package service

import (
	"context"
	"fmt"
	"time"

	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/timezone"
	"todo/internal/utils"
)

type TaskService struct {
	repo               repository.TaskRepository
	reminderConfigRepo repository.ReminderConfigRepository
	tagService         *TagService
}

func NewTaskService(repo repository.TaskRepository, reminderConfigRepo repository.ReminderConfigRepository, tagService *TagService) *TaskService {
	return &TaskService{
		repo:               repo,
		reminderConfigRepo: reminderConfigRepo,
		tagService:         tagService,
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

	// 校验 tags 全部存在于该用户字典
	if s.tagService != nil {
		clean, err := s.tagService.ValidateTagsExist(ctx, userID, req.Tags)
		if err != nil {
			return nil, err
		}
		req.Tags = clean
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
	if req.Tags != nil && s.tagService != nil {
		clean, err := s.tagService.ValidateTagsExist(ctx, userID, *req.Tags)
		if err != nil {
			return nil, err
		}
		req.Tags = &clean
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

func (s *TaskService) ToggleComplete(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error) {
	task, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	var next *models.Task
	// 任务完成且有重复规则时，预计算下一次实例（但还不知道 completed 的新值）
	// 先判断当前状态：若当前未完成，切换后会变完成，此时需要生成下一次
	if !task.Completed && task.RepeatType != "none" {
		next, err = s.buildNextOccurrence(task)
		if err != nil {
			return nil, fmt.Errorf("build next occurrence: %w", err)
		}
	}

	return s.repo.ToggleCompleteAndCreateRepeat(ctx, userID, id, next, focusDuration)
}

func (s *TaskService) buildNextOccurrence(t *models.Task) (*models.Task, error) {
	var nextDue, nextRemind *string

	if t.DueAt != nil {
		next, err := models.CalculateNextDueDate(*t.DueAt, t.RepeatType, t.RepeatInterval)
		if err != nil {
			return nil, fmt.Errorf("calculate next due_at: %w", err)
		}
		nextDue = &next
	}
	if t.RemindAt != nil {
		next, err := models.CalculateNextDueDate(*t.RemindAt, t.RepeatType, t.RepeatInterval)
		if err != nil {
			return nil, fmt.Errorf("calculate next remind_at: %w", err)
		}
		nextRemind = &next
	}

	if t.RepeatEndDate != nil && nextDue != nil {
		if *nextDue > *t.RepeatEndDate {
			return nil, nil
		}
	}

	return &models.Task{
		UserID:         t.UserID,
		Title:          t.Title,
		Description:    t.Description,
		Priority:       t.Priority,
		DueAt:          nextDue,
		RemindAt:       nextRemind,
		RepeatType:     t.RepeatType,
		RepeatInterval: t.RepeatInterval,
		RepeatEndDate:  t.RepeatEndDate,
		Tags:           t.Tags,
	}, nil
}

// normalizeCreateTaskTimes 标准化创建任务请求中的时间字段。
// nil 和空字符串视为未设置；RFC3339 字符串转为 UTC RFC3339；
// 无时区时间按 server.timezone 解释;非法字符串返回错误。
func normalizeCreateTaskTimes(req *models.CreateTaskRequest) error {
	loc := timezone.Get()
	var err error
	if req.DueAt, err = normalizeOptionalTime(req.DueAt, true, loc); err != nil {
		return fmt.Errorf("due_at: %w", err)
	}
	if req.RemindAt, err = normalizeOptionalTime(req.RemindAt, true, loc); err != nil {
		return fmt.Errorf("remind_at: %w", err)
	}
	if req.RepeatEndDate, err = normalizeOptionalTime(req.RepeatEndDate, true, loc); err != nil {
		return fmt.Errorf("repeat_end_date: %w", err)
	}
	return nil
}

// normalizeUpdateTaskTimes 标准化更新任务请求中的时间字段。
// nil 表示不修改；空字符串表示清空；RFC3339 字符串转为 UTC RFC3339；
// 无时区时间按 server.timezone 解释;非法字符串返回错误。
func normalizeUpdateTaskTimes(req *models.UpdateTaskRequest) error {
	loc := timezone.Get()
	var err error
	if req.DueAt, err = normalizeOptionalTime(req.DueAt, false, loc); err != nil {
		return fmt.Errorf("due_at: %w", err)
	}
	if req.RemindAt, err = normalizeOptionalTime(req.RemindAt, false, loc); err != nil {
		return fmt.Errorf("remind_at: %w", err)
	}
	if req.RepeatEndDate, err = normalizeOptionalTime(req.RepeatEndDate, false, loc); err != nil {
		return fmt.Errorf("repeat_end_date: %w", err)
	}
	return nil
}

// normalizeOptionalTime 规范化可选时间字段。
// isCreate 为 true 时，空字符串视为 nil（未设置）；为 false 时，空字符串保留（清空字段）。
// defaultLoc 用于解释不带时区的时间字符串。
func normalizeOptionalTime(p *string, isCreate bool, defaultLoc *time.Location) (*string, error) {
	if p == nil {
		return nil, nil
	}
	if *p == "" {
		if isCreate {
			return nil, nil
		}
		return p, nil
	}
	normalized, err := utils.NormalizeAPITime(*p, defaultLoc)
	if err != nil {
		return nil, ErrInvalidTime
	}
	return &normalized, nil
}
