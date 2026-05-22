package service

import (
	"context"

	"todo/internal/models"
	"todo/internal/repository"
)

type ReminderConfigService struct {
	repo repository.ReminderConfigRepository
}

func NewReminderConfigService(repo repository.ReminderConfigRepository) *ReminderConfigService {
	return &ReminderConfigService{repo: repo}
}

func (s *ReminderConfigService) Create(ctx context.Context, userID int64, req models.CreateReminderConfigRequest) (*models.UserReminderConfig, error) {
	cfg := &models.UserReminderConfig{
		UserID:              userID,
		Name:                req.Name,
		ChannelType:         req.ChannelType,
		WebhookURL:          req.WebhookURL,
		WebhookMethod:       req.WebhookMethod,
		WebhookHeaders:      req.WebhookHeaders,
		WebhookBodyTemplate: req.WebhookBodyTemplate,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		Enabled:             true,
	}
	if cfg.WebhookMethod == "" {
		cfg.WebhookMethod = "POST"
	}
	if req.MaxRetries != nil {
		cfg.MaxRetries = *req.MaxRetries
	}
	if req.RetryDelaySeconds != nil {
		cfg.RetryDelaySeconds = *req.RetryDelaySeconds
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}

	return s.repo.Create(ctx, cfg)
}

func (s *ReminderConfigService) GetByID(ctx context.Context, userID, id int64) (*models.UserReminderConfig, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *ReminderConfigService) List(ctx context.Context, userID int64) ([]models.UserReminderConfig, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ReminderConfigService) Update(ctx context.Context, userID, id int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error) {
	return s.repo.Update(ctx, id, userID, req)
}

func (s *ReminderConfigService) Delete(ctx context.Context, userID, id int64) (bool, error) {
	return s.repo.Delete(ctx, id, userID)
}

func (s *ReminderConfigService) HasEnabled(ctx context.Context, userID int64) (bool, error) {
	return s.repo.HasEnabledByUserID(ctx, userID)
}
