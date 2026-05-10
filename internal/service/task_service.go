package service

import (
	"context"
	"fmt"

	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/utils"
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
	// 标准化时间字段
	if err := normalizeCreateTaskTimes(&req); err != nil {
		return nil, err
	}

	if err := s.requireEnabledReminderChannel(ctx, userID, req.RemindAt); err != nil {
		return nil, err
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
	if err := normalizeUpdateTaskTimes(&req); err != nil {
		return nil, err
	}
	if err := s.requireEnabledReminderChannel(ctx, userID, req.RemindAt); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, userID, id, req)
}

func (s *TaskService) Delete(ctx context.Context, userID, id int64) (bool, error) {
	return s.repo.Delete(ctx, userID, id)
}

func (s *TaskService) requireEnabledReminderChannel(ctx context.Context, userID int64, remindAt *string) error {
	// 仅当设置了提醒时间时，才要求存在已启用的提醒通道。
	if remindAt == nil || *remindAt == "" {
		return nil
	}

	hasEnabledReminder, err := s.reminderConfigRepo.HasEnabledByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if !hasEnabledReminder {
		return ErrReminderChannelMissing
	}
	return nil
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
		next, err := models.CalculateNextDueDate(*t.DueAt, t.RepeatType, t.RepeatInterval)
		if err != nil {
			return fmt.Errorf("calculate next due_at: %w", err)
		}
		nextDue = &next
	}
	if t.RemindAt != nil {
		next, err := models.CalculateNextDueDate(*t.RemindAt, t.RepeatType, t.RepeatInterval)
		if err != nil {
			return fmt.Errorf("calculate next remind_at: %w", err)
		}
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

// normalizeCreateTaskTimes 标准化创建任务请求中的时间字段。
// nil 和空字符串视为未设置；RFC3339 字符串转为 UTC RFC3339；非法字符串返回错误。
func normalizeCreateTaskTimes(req *models.CreateTaskRequest) error {
	var err error
	if req.DueAt, err = normalizeOptionalTime(req.DueAt, true); err != nil {
		return fmt.Errorf("due_at: %w", err)
	}
	if req.RemindAt, err = normalizeOptionalTime(req.RemindAt, true); err != nil {
		return fmt.Errorf("remind_at: %w", err)
	}
	if req.RepeatEndDate, err = normalizeOptionalTime(req.RepeatEndDate, true); err != nil {
		return fmt.Errorf("repeat_end_date: %w", err)
	}
	return nil
}

// normalizeUpdateTaskTimes 标准化更新任务请求中的时间字段。
// nil 表示不修改；空字符串表示清空；RFC3339 字符串转为 UTC RFC3339；非法字符串返回错误。
func normalizeUpdateTaskTimes(req *models.UpdateTaskRequest) error {
	var err error
	if req.DueAt, err = normalizeOptionalTime(req.DueAt, false); err != nil {
		return fmt.Errorf("due_at: %w", err)
	}
	if req.RemindAt, err = normalizeOptionalTime(req.RemindAt, false); err != nil {
		return fmt.Errorf("remind_at: %w", err)
	}
	if req.RepeatEndDate, err = normalizeOptionalTime(req.RepeatEndDate, false); err != nil {
		return fmt.Errorf("repeat_end_date: %w", err)
	}
	return nil
}

// normalizeOptionalTime 规范化可选时间字段。
// isCreate 为 true 时，空字符串视为 nil（未设置）；为 false 时，空字符串保留（清空字段）。
func normalizeOptionalTime(p *string, isCreate bool) (*string, error) {
	if p == nil {
		return nil, nil
	}
	if *p == "" {
		if isCreate {
			return nil, nil
		}
		return p, nil
	}
	normalized, err := utils.NormalizeAPITime(*p)
	if err != nil {
		return nil, ErrInvalidTime
	}
	return &normalized, nil
}
