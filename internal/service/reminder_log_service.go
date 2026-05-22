package service

import (
	"context"

	"todo/internal/models"
	"todo/internal/repository"
)

type ReminderLogService struct {
	repo repository.ReminderLogRepository
}

func NewReminderLogService(repo repository.ReminderLogRepository) *ReminderLogService {
	return &ReminderLogService{repo: repo}
}

func (s *ReminderLogService) List(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error) {
	return s.repo.ListByUserID(ctx, userID, page, limit)
}
