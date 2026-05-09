package service

import (
	"context"
	"fmt"

	"todo/internal/models"
	"todo/internal/repository"
)

type TaskService struct {
	repo               *repository.TaskRepo
	reminderConfigRepo *repository.ReminderConfigRepo
}

func NewTaskService(repo *repository.TaskRepo, reminderConfigRepo *repository.ReminderConfigRepo) *TaskService {
	return &TaskService{
		repo:               repo,
		reminderConfigRepo: reminderConfigRepo,
	}
}

func (s *TaskService) Create(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
	// 仅当设置了提醒时间时，才要求存在已启用的提醒通道
	if req.RemindAt != nil && *req.RemindAt != "" {
		hasEnabledReminder, err := s.reminderConfigRepo.HasEnabledByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !hasEnabledReminder {
			return nil, ErrReminderChannelMissing
		}
	}
	return s.repo.Create(ctx, userID, req)
}

func (s *TaskService) GetByID(ctx context.Context, userID, id int64) (*models.Task, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *TaskService) List(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error) {
	return s.repo.List(ctx, userID, filters, page, limit, sortField, sortOrder)
}

func (s *TaskService) Update(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	return s.repo.Update(ctx, userID, id, req)
}

func (s *TaskService) Delete(ctx context.Context, userID, id int64) (bool, error) {
	return s.repo.Delete(ctx, userID, id)
}

func (s *TaskService) ToggleComplete(ctx context.Context, userID, id int64) (*models.Task, error) {
	task, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	updated, err := s.repo.ToggleComplete(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if updated.Completed && task.RepeatType != "none" {
		if err := s.createNextOccurrence(ctx, task); err != nil {
			return updated, fmt.Errorf("create next occurrence: %w", err)
		}
	}

	return updated, nil
}

func (s *TaskService) createNextOccurrence(ctx context.Context, t *models.Task) error {
	var nextDue, nextRemind *string

	if t.DueAt != nil {
		next := models.CalculateNextDueDate(*t.DueAt, t.RepeatType, t.RepeatInterval)
		nextDue = &next
	}
	if t.RemindAt != nil {
		next := models.CalculateNextDueDate(*t.RemindAt, t.RepeatType, t.RepeatInterval)
		nextRemind = &next
	}

	if t.RepeatEndDate != nil && nextDue != nil {
		if *nextDue > *t.RepeatEndDate {
			return nil
		}
	}

	newTask := &models.Task{
		UserID:         t.UserID,
		Title:          t.Title,
		Description:    t.Description,
		Priority:       t.Priority,
		DueAt:          nextDue,
		RemindAt:       nextRemind,
		RepeatType:     t.RepeatType,
		RepeatInterval: t.RepeatInterval,
		RepeatEndDate:  t.RepeatEndDate,
	}
	return s.repo.CreateRepeatTask(ctx, newTask)
}
