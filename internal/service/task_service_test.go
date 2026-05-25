package service_test

import (
	"context"
	"errors"
	"testing"

	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/testutil"
)

func TestTaskService_Create_NoRemindAt(t *testing.T) {
	repo := &testutil.MockTaskRepository{
		CreateFn: func(_ context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
			return &models.Task{ID: 1, UserID: userID, Title: req.Title}, nil
		},
	}
	cfgRepo := &testutil.MockReminderConfigRepository{}
	svc := service.NewTaskService(repo, cfgRepo)

	task, err := svc.Create(context.Background(), 1, models.CreateTaskRequest{Title: "buy milk"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.Title != "buy milk" {
		t.Error("expected task with title 'buy milk'")
	}
}

func TestTaskService_Create_RemindAt_NoChannel(t *testing.T) {
	repo := &testutil.MockTaskRepository{}
	cfgRepo := &testutil.MockReminderConfigRepository{
		HasEnabledByUserIDFn: func(_ context.Context, userID int64) (bool, error) {
			return false, nil
		},
	}
	svc := service.NewTaskService(repo, cfgRepo)

	remindAt := "2099-01-01T00:00:00Z"
	_, err := svc.Create(context.Background(), 1, models.CreateTaskRequest{
		Title:    "remind me",
		RemindAt: &remindAt,
	})
	if !errors.Is(err, service.ErrReminderChannelMissing) {
		t.Errorf("expected ErrReminderChannelMissing, got %v", err)
	}
}

func TestTaskService_Create_RemindAt_WithChannel(t *testing.T) {
	remindAt := "2099-01-01T00:00:00Z"
	repo := &testutil.MockTaskRepository{
		CreateFn: func(_ context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
			return &models.Task{ID: 1, UserID: userID, Title: req.Title}, nil
		},
	}
	cfgRepo := &testutil.MockReminderConfigRepository{
		HasEnabledByUserIDFn: func(_ context.Context, userID int64) (bool, error) {
			return true, nil
		},
	}
	svc := service.NewTaskService(repo, cfgRepo)

	task, err := svc.Create(context.Background(), 1, models.CreateTaskRequest{
		Title:    "remind me",
		RemindAt: &remindAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Error("expected non-nil task")
	}
}

func TestTaskService_ToggleComplete_NotFound(t *testing.T) {
	repo := &testutil.MockTaskRepository{
		GetByIDFn: func(_ context.Context, userID, id int64) (*models.Task, error) {
			return nil, nil
		},
	}
	svc := service.NewTaskService(repo, &testutil.MockReminderConfigRepository{})

	task, err := svc.ToggleComplete(context.Background(), 1, 99, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task != nil {
		t.Error("expected nil for not-found task")
	}
}

func TestTaskService_ToggleComplete_NoRepeat(t *testing.T) {
	existing := &models.Task{ID: 1, UserID: 1, Title: "once", Completed: false, RepeatType: "none"}
	var capturedNext *models.Task
	repo := &testutil.MockTaskRepository{
		GetByIDFn: func(_ context.Context, _, _ int64) (*models.Task, error) { return existing, nil },
		ToggleCompleteAndCreateRepeatFn: func(_ context.Context, _, _ int64, next *models.Task, focusDuration *int) (*models.Task, error) {
			capturedNext = next
			done := *existing
			done.Completed = true
			return &done, nil
		},
	}
	svc := service.NewTaskService(repo, &testutil.MockReminderConfigRepository{})

	task, err := svc.ToggleComplete(context.Background(), 1, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !task.Completed {
		t.Error("expected task to be marked complete")
	}
	if capturedNext != nil {
		t.Error("expected no next occurrence for non-repeating task")
	}
}

func TestTaskService_ToggleComplete_WithRepeat(t *testing.T) {
	dueAt := "2026-05-22T10:00:00Z"
	existing := &models.Task{
		ID: 1, UserID: 1, Title: "daily",
		Completed: false, RepeatType: "daily", RepeatInterval: 1, DueAt: &dueAt,
	}
	var capturedNext *models.Task
	repo := &testutil.MockTaskRepository{
		GetByIDFn: func(_ context.Context, _, _ int64) (*models.Task, error) { return existing, nil },
		ToggleCompleteAndCreateRepeatFn: func(_ context.Context, _, _ int64, next *models.Task, focusDuration *int) (*models.Task, error) {
			capturedNext = next
			done := *existing
			done.Completed = true
			return &done, nil
		},
	}
	svc := service.NewTaskService(repo, &testutil.MockReminderConfigRepository{})

	task, err := svc.ToggleComplete(context.Background(), 1, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !task.Completed {
		t.Error("expected task to be marked complete")
	}
	if capturedNext == nil {
		t.Fatal("expected a next occurrence for daily repeating task")
	}
	if capturedNext.Title != "daily" {
		t.Errorf("next occurrence title: want 'daily', got %q", capturedNext.Title)
	}
	if capturedNext.DueAt == nil {
		t.Error("expected next occurrence to have due_at")
	}
}

func TestTaskService_ToggleComplete_UncompletingDoesNotGenerateNext(t *testing.T) {
	dueAt := "2026-05-22T10:00:00Z"
	existing := &models.Task{
		ID: 1, UserID: 1, Completed: true, RepeatType: "daily", RepeatInterval: 1, DueAt: &dueAt,
	}
	var capturedNext *models.Task
	repo := &testutil.MockTaskRepository{
		GetByIDFn: func(_ context.Context, _, _ int64) (*models.Task, error) { return existing, nil },
		ToggleCompleteAndCreateRepeatFn: func(_ context.Context, _, _ int64, next *models.Task, focusDuration *int) (*models.Task, error) {
			capturedNext = next
			undone := *existing
			undone.Completed = false
			return &undone, nil
		},
	}
	svc := service.NewTaskService(repo, &testutil.MockReminderConfigRepository{})

	_, err := svc.ToggleComplete(context.Background(), 1, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedNext != nil {
		t.Error("should not create next occurrence when un-completing a task")
	}
}
