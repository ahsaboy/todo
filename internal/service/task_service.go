package service

import (
	"context"
	"fmt"

	"todo/internal/models"
	"todo/internal/repository"
)

type TaskService struct {
	repo *repository.TaskRepo
}

func NewTaskService(repo *repository.TaskRepo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error) {
	return s.repo.Create(ctx, req)
}

func (s *TaskService) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error) {
	return s.repo.List(ctx, filters, page, limit, sortField, sortOrder)
}

func (s *TaskService) Update(ctx context.Context, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *TaskService) Delete(ctx context.Context, id int64) (bool, error) {
	return s.repo.Delete(ctx, id)
}

func (s *TaskService) ToggleComplete(ctx context.Context, id int64) (*models.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	updated, err := s.repo.ToggleComplete(ctx, id)
	if err != nil {
		return nil, err
	}

	// 完成时自动生成下一次重复任务
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

	// 检查是否超过结束日期
	if t.RepeatEndDate != nil && nextDue != nil {
		if *nextDue > *t.RepeatEndDate {
			return nil
		}
	}

	newTask := &models.Task{
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
