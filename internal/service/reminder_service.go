package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"go.uber.org/zap"

	"todo/internal/config"
	"todo/internal/models"
	"todo/internal/repository"
)

type ReminderService struct {
	taskRepo     *repository.TaskRepo
	configRepo   *repository.ReminderConfigRepo
	logger       *zap.Logger
	client       *http.Client
	enabled      bool
	scanInterval time.Duration
	defaultTmpl  *template.Template
}

func NewReminderService(
	taskRepo *repository.TaskRepo,
	configRepo *repository.ReminderConfigRepo,
	cfg config.ReminderConfig,
	logger *zap.Logger,
) (*ReminderService, error) {
	tmpl, err := template.New("webhook").Parse(cfg.WebhookBodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse default webhook template: %w", err)
	}

	scanInterval := time.Duration(cfg.ScanIntervalSeconds) * time.Second
	if scanInterval <= 0 {
		scanInterval = 30 * time.Second
	}

	timeout := time.Duration(cfg.WebhookTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &ReminderService{
		taskRepo:     taskRepo,
		configRepo:   configRepo,
		logger:       logger,
		client:       &http.Client{Timeout: timeout},
		enabled:      cfg.Enabled,
		scanInterval: scanInterval,
		defaultTmpl:  tmpl,
	}, nil
}

func (s *ReminderService) Start(ctx context.Context) {
	if !s.enabled {
		s.logger.Info("reminder service disabled")
		return
	}

	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	s.logger.Info("reminder service started", zap.Duration("interval", s.scanInterval))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("reminder service stopped")
			return
		case <-ticker.C:
			s.processReminders(ctx)
		}
	}
}

func (s *ReminderService) processReminders(ctx context.Context) {
	tasks, err := s.taskRepo.GetPendingReminders(ctx)
	if err != nil {
		s.logger.Error("get pending reminders failed", zap.Error(err))
		return
	}

	for _, task := range tasks {
		s.processTaskReminder(ctx, &task)
	}
}

func (s *ReminderService) processTaskReminder(ctx context.Context, task *models.Task) {
	configs, err := s.deliveryConfigs(ctx, task.UserID)
	if err != nil {
		s.logger.Error("get reminder delivery config failed",
			zap.Int64("task_id", task.ID),
			zap.Int64("user_id", task.UserID),
			zap.Error(err),
		)
		return
	}
	if len(configs) == 0 {
		s.logger.Debug("no enabled reminder config for task, skipping",
			zap.Int64("task_id", task.ID),
			zap.Int64("user_id", task.UserID),
		)
		return
	}

	allSucceeded := true
	for _, cfg := range configs {
		if err := s.sendToChannel(ctx, task, &cfg); err != nil {
			allSucceeded = false
			s.logger.Error("send reminder to channel failed",
				zap.Int64("task_id", task.ID),
				zap.Int64("user_id", task.UserID),
				zap.String("channel", cfg.Name),
				zap.Error(err),
			)
		}
	}

	if !allSucceeded {
		return
	}

	if _, err := s.taskRepo.MarkReminderSent(ctx, task.ID); err != nil {
		s.logger.Error("mark reminder sent failed", zap.Int64("task_id", task.ID), zap.Error(err))
	}
}

func (s *ReminderService) deliveryConfigs(ctx context.Context, userID int64) ([]models.UserReminderConfig, error) {
	configs, err := s.configRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	enabledConfigs := make([]models.UserReminderConfig, 0, len(configs))
	for _, cfg := range configs {
		if cfg.Enabled {
			enabledConfigs = append(enabledConfigs, cfg)
		}
	}
	if len(enabledConfigs) > 0 {
		return enabledConfigs, nil
	}
	return nil, nil
}

func (s *ReminderService) sendToChannel(ctx context.Context, task *models.Task, cfg *models.UserReminderConfig) error {
	tmpl := s.defaultTmpl
	if cfg.WebhookBodyTemplate != "" {
		var err error
		tmpl, err = template.New("webhook").Parse(cfg.WebhookBodyTemplate)
		if err != nil {
			return fmt.Errorf("parse webhook template: %w", err)
		}
	}

	// 渲染模板
	data := task.ToTemplateData()
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// 带重试的 HTTP 推送
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	retryDelay := cfg.RetryDelaySeconds
	if retryDelay <= 0 {
		retryDelay = 5
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := s.doHTTPRequest(ctx, cfg, body.String())
		if err == nil {
			s.logger.Info("reminder sent",
				zap.Int64("task_id", task.ID),
				zap.String("title", task.Title),
				zap.String("channel", cfg.Name),
				zap.Int("attempt", attempt),
			)
			return nil
		}
		lastErr = err
		s.logger.Warn("reminder send failed, retrying",
			zap.Int64("task_id", task.ID),
			zap.String("channel", cfg.Name),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		time.Sleep(time.Duration(retryDelay*attempt) * time.Second)
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (s *ReminderService) doHTTPRequest(ctx context.Context, cfg *models.UserReminderConfig, body string) error {
	method := cfg.WebhookMethod
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, cfg.WebhookURL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	for k, v := range cfg.WebhookHeaders {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *ReminderService) UpdateTemplate(tmplStr string) error {
	tmpl, err := template.New("webhook").Parse(tmplStr)
	if err != nil {
		return err
	}
	s.defaultTmpl = tmpl
	return nil
}
