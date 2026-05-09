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
	repo    *repository.TaskRepo
	cfg     config.ReminderConfig
	logger  *zap.Logger
	client  *http.Client
	tmpl    *template.Template
}

func NewReminderService(repo *repository.TaskRepo, cfg config.ReminderConfig, logger *zap.Logger) (*ReminderService, error) {
	tmpl, err := template.New("webhook").Parse(cfg.WebhookBodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse webhook template: %w", err)
	}

	return &ReminderService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: time.Duration(cfg.WebhookTimeoutSeconds) * time.Second},
		tmpl:   tmpl,
	}, nil
}

func (s *ReminderService) Start(ctx context.Context) {
	if !s.cfg.Enabled {
		s.logger.Info("reminder service disabled")
		return
	}

	interval := time.Duration(s.cfg.ScanIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("reminder service started", zap.Duration("interval", interval))

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
	tasks, err := s.repo.GetPendingReminders(ctx)
	if err != nil {
		s.logger.Error("get pending reminders failed", zap.Error(err))
		return
	}

	for _, task := range tasks {
		if err := s.sendReminder(ctx, &task); err != nil {
			s.logger.Error("send reminder failed",
				zap.Int64("task_id", task.ID),
				zap.String("title", task.Title),
				zap.Error(err),
			)
		}
	}
}

func (s *ReminderService) sendReminder(ctx context.Context, task *models.Task) error {
	// 原子标记，防止重复发送
	sent, err := s.repo.MarkReminderSent(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("mark reminder sent: %w", err)
	}
	if !sent {
		return nil // 已被其他实例处理
	}

	// 渲染模板
	data := task.ToTemplateData()
	var body bytes.Buffer
	if err := s.tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// 带重试的 HTTP 推送
	var lastErr error
	for attempt := 1; attempt <= s.cfg.MaxRetries; attempt++ {
		err := s.doHTTPRequest(body.String())
		if err == nil {
			s.logger.Info("reminder sent",
				zap.Int64("task_id", task.ID),
				zap.String("title", task.Title),
				zap.Int("attempt", attempt),
			)
			return nil
		}
		lastErr = err
		s.logger.Warn("reminder send failed, retrying",
			zap.Int64("task_id", task.ID),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		time.Sleep(time.Duration(s.cfg.RetryDelaySeconds*attempt) * time.Second)
	}

	return fmt.Errorf("failed after %d retries: %w", s.cfg.MaxRetries, lastErr)
}

func (s *ReminderService) doHTTPRequest(body string) error {
	method := s.cfg.WebhookMethod
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequest(method, s.cfg.WebhookURL, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	for k, v := range s.cfg.WebhookHeaders {
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

// UpdateTemplate 动态更新 Webhook 模板
func (s *ReminderService) UpdateTemplate(tmplStr string) error {
	tmpl, err := template.New("webhook").Parse(tmplStr)
	if err != nil {
		return err
	}
	s.tmpl = tmpl
	return nil
}
