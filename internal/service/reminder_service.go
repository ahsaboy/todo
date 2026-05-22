package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"go.uber.org/zap"

	"todo/internal/config"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/timezone"
)

type ReminderService struct {
	taskRepo     repository.TaskRepository
	configRepo   repository.ReminderConfigRepository
	logRepo      repository.ReminderLogRepository
	logger       *zap.Logger
	client       *http.Client
	enabled      bool
	scanInterval time.Duration
	defaultTmpl  *template.Template
	workerCount  int
}

func NewReminderService(
	taskRepo repository.TaskRepository,
	configRepo repository.ReminderConfigRepository,
	logRepo repository.ReminderLogRepository,
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

	workerCount := cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = 5
	}

	return &ReminderService{
		taskRepo:     taskRepo,
		configRepo:   configRepo,
		logRepo:      logRepo,
		logger:       logger,
		client:       &http.Client{Timeout: timeout},
		enabled:      cfg.Enabled,
		scanInterval: scanInterval,
		defaultTmpl:  tmpl,
		workerCount:  workerCount,
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
	tasks, err := s.taskRepo.GetPendingReminders(ctx, time.Now())
	if err != nil {
		s.logger.Error("get pending reminders failed", zap.Error(err))
		return
	}

	sem := make(chan struct{}, s.workerCount)
	var wg sync.WaitGroup

	for i := range tasks {
		task := &tasks[i]
		select {
		case <-ctx.Done():
			break
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			s.processTaskReminder(ctx, task)
		}()
	}
	wg.Wait()
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

	allAttempted := true
	for _, cfg := range configs {
		hasResult, err := s.logRepo.HasResultForTaskConfig(ctx, task.ID, cfg.ID)
		if err != nil {
			allAttempted = false
			s.logger.Error("check reminder log failed",
				zap.Int64("task_id", task.ID),
				zap.Int64("user_id", task.UserID),
				zap.String("channel", cfg.Name),
				zap.Error(err),
			)
			continue
		}
		if hasResult {
			continue
		}

		attempts, err := s.sendToChannel(ctx, task, &cfg)
		status := "success"
		errorMessage := ""
		if err != nil {
			status = "failed"
			errorMessage = err.Error()
			s.logger.Error("send reminder to channel failed",
				zap.Int64("task_id", task.ID),
				zap.Int64("user_id", task.UserID),
				zap.String("channel", cfg.Name),
				zap.Error(err),
			)
		}

		if err := s.logRepo.Upsert(ctx, repository.CreateReminderLogParams{
			UserID:           task.UserID,
			TaskID:           task.ID,
			ReminderConfigID: cfg.ID,
			ChannelName:      cfg.Name,
			ChannelType:      cfg.ChannelType,
			Status:           status,
			Attempts:         attempts,
			ErrorMessage:     errorMessage,
		}); err != nil {
			allAttempted = false
			s.logger.Error("write reminder log failed",
				zap.Int64("task_id", task.ID),
				zap.Int64("user_id", task.UserID),
				zap.String("channel", cfg.Name),
				zap.Error(err),
			)
		}
	}

	if !allAttempted {
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

func (s *ReminderService) sendToChannel(ctx context.Context, task *models.Task, cfg *models.UserReminderConfig) (int, error) {
	tmpl := s.defaultTmpl
	if cfg.WebhookBodyTemplate != "" {
		var err error
		tmpl, err = template.New("webhook").Parse(cfg.WebhookBodyTemplate)
		if err != nil {
			return 0, fmt.Errorf("parse webhook template: %w", err)
		}
	}

	// 渲染模板,按全局 server.timezone 格式化时间字段
	data := task.ToTemplateData(timezone.Get())
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return 0, fmt.Errorf("execute template: %w", err)
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
			return attempt, nil
		}
		lastErr = err
		s.logger.Warn("reminder send failed, retrying",
			zap.Int64("task_id", task.ID),
			zap.String("channel", cfg.Name),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		if attempt == maxRetries {
			break
		}
		select {
		case <-ctx.Done():
			return attempt, ctx.Err()
		case <-time.After(time.Duration(retryDelay*attempt) * time.Second):
		}
	}

	return maxRetries, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
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
