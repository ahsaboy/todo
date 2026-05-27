// Package testutil 提供测试辅助工具，包括手写的 mock 实现。
// 这些 mock 实现了 repository 层的接口，可在 service 单元测试中替代真实数据库访问。
package testutil

import (
	"context"
	"time"

	"todo/internal/models"
	"todo/internal/repository"
)

// ---- TaskRepository mock ----

type MockTaskRepository struct {
	CreateFn                        func(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error)
	GetByIDFn                       func(ctx context.Context, userID, id int64) (*models.Task, error)
	ListFn                          func(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error)
	UpdateFn                        func(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error)
	DeleteFn                        func(ctx context.Context, userID, id int64) (bool, error)
	ToggleCompleteFn                func(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error)
	ToggleCompleteAndCreateRepeatFn func(ctx context.Context, userID, id int64, next *models.Task, focusDuration *int) (*models.Task, error)
	GetPendingRemindersFn           func(ctx context.Context, now time.Time) ([]models.Task, error)
	MarkReminderSentFn              func(ctx context.Context, id int64) (bool, error)
	CreateRepeatTaskFn              func(ctx context.Context, t *models.Task) error
	ListAllFn                       func(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int) ([]models.Task, int64, error)
	AdminGetByIDFn                  func(ctx context.Context, id int64) (*models.Task, error)
	AdminToggleCompleteFn           func(ctx context.Context, id int64) (*models.Task, error)
	AdminUpdateFn                   func(ctx context.Context, id int64, req models.UpdateTaskRequest) (*models.Task, error)
	AdminDeleteFn                   func(ctx context.Context, id int64) (bool, error)
}

func (m *MockTaskRepository) Create(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
	return m.CreateFn(ctx, userID, req)
}
func (m *MockTaskRepository) GetByID(ctx context.Context, userID, id int64) (*models.Task, error) {
	return m.GetByIDFn(ctx, userID, id)
}
func (m *MockTaskRepository) List(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error) {
	return m.ListFn(ctx, userID, filters, page, limit, sortField, sortOrder)
}
func (m *MockTaskRepository) Update(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	return m.UpdateFn(ctx, userID, id, req)
}
func (m *MockTaskRepository) Delete(ctx context.Context, userID, id int64) (bool, error) {
	return m.DeleteFn(ctx, userID, id)
}
func (m *MockTaskRepository) ToggleComplete(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error) {
	return m.ToggleCompleteFn(ctx, userID, id, focusDuration)
}
func (m *MockTaskRepository) ToggleCompleteAndCreateRepeat(ctx context.Context, userID, id int64, next *models.Task, focusDuration *int) (*models.Task, error) {
	return m.ToggleCompleteAndCreateRepeatFn(ctx, userID, id, next, focusDuration)
}
func (m *MockTaskRepository) GetPendingReminders(ctx context.Context, now time.Time) ([]models.Task, error) {
	return m.GetPendingRemindersFn(ctx, now)
}
func (m *MockTaskRepository) MarkReminderSent(ctx context.Context, id int64) (bool, error) {
	return m.MarkReminderSentFn(ctx, id)
}
func (m *MockTaskRepository) CreateRepeatTask(ctx context.Context, t *models.Task) error {
	return m.CreateRepeatTaskFn(ctx, t)
}
func (m *MockTaskRepository) ListAll(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int) ([]models.Task, int64, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx, userID, filters, page, limit)
	}
	return nil, 0, nil
}
func (m *MockTaskRepository) AdminGetByID(ctx context.Context, id int64) (*models.Task, error) {
	if m.AdminGetByIDFn != nil {
		return m.AdminGetByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *MockTaskRepository) AdminToggleComplete(ctx context.Context, id int64) (*models.Task, error) {
	if m.AdminToggleCompleteFn != nil {
		return m.AdminToggleCompleteFn(ctx, id)
	}
	return nil, nil
}
func (m *MockTaskRepository) AdminUpdate(ctx context.Context, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	if m.AdminUpdateFn != nil {
		return m.AdminUpdateFn(ctx, id, req)
	}
	return nil, nil
}
func (m *MockTaskRepository) AdminDelete(ctx context.Context, id int64) (bool, error) {
	if m.AdminDeleteFn != nil {
		return m.AdminDeleteFn(ctx, id)
	}
	return false, nil
}

// ---- UserRepository mock ----

type MockUserRepository struct {
	CreateFn             func(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	GetByIDFn            func(ctx context.Context, id int64) (*models.User, error)
	GetByUsernameFn      func(ctx context.Context, username string) (*models.User, error)
	GetByEmailFn         func(ctx context.Context, email string) (*models.User, error)
	UpdateProfileFn      func(ctx context.Context, id int64, email string) error
	UpdatePasswordFn     func(ctx context.Context, id int64, passwordHash string) error
	ListAllFn            func(ctx context.Context, page, limit int, search string) ([]models.User, int64, error)
	DeleteFn             func(ctx context.Context, id int64) error
	ForceResetPasswordFn func(ctx context.Context, id int64, passwordHash string) error
	IsAdminFn            func(ctx context.Context, userID int64) (bool, error)
	SetIsAdminFn         func(ctx context.Context, userID int64, isAdmin bool) error
}

func (m *MockUserRepository) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	return m.CreateFn(ctx, username, email, passwordHash)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return m.GetByUsernameFn(ctx, username)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.GetByEmailFn(ctx, email)
}
func (m *MockUserRepository) UpdateProfile(ctx context.Context, id int64, email string) error {
	return m.UpdateProfileFn(ctx, id, email)
}
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	return m.UpdatePasswordFn(ctx, id, passwordHash)
}
func (m *MockUserRepository) ListAll(ctx context.Context, page, limit int, search string) ([]models.User, int64, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx, page, limit, search)
	}
	return nil, 0, nil
}
func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
func (m *MockUserRepository) ForceResetPassword(ctx context.Context, id int64, passwordHash string) error {
	if m.ForceResetPasswordFn != nil {
		return m.ForceResetPasswordFn(ctx, id, passwordHash)
	}
	return nil
}
func (m *MockUserRepository) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	if m.IsAdminFn != nil {
		return m.IsAdminFn(ctx, userID)
	}
	return false, nil
}
func (m *MockUserRepository) SetIsAdmin(ctx context.Context, userID int64, isAdmin bool) error {
	if m.SetIsAdminFn != nil {
		return m.SetIsAdminFn(ctx, userID, isAdmin)
	}
	return nil
}

// ---- APIKeyRepository mock ----

type MockAPIKeyRepository struct {
	CreateFn                  func(ctx context.Context, userID int64, keyHash, name string) (*models.APIKey, error)
	GetByUserIDFn             func(ctx context.Context, userID int64) ([]models.APIKey, error)
	DeleteFn                  func(ctx context.Context, id, userID int64) (bool, error)
	ValidateKeyFn             func(ctx context.Context, keyHash string) (int64, error)
	CleanupExpiredLoginKeysFn func(ctx context.Context, userID int64) (int64, error)
}

func (m *MockAPIKeyRepository) Create(ctx context.Context, userID int64, keyHash, name string) (*models.APIKey, error) {
	return m.CreateFn(ctx, userID, keyHash, name)
}
func (m *MockAPIKeyRepository) GetByUserID(ctx context.Context, userID int64) ([]models.APIKey, error) {
	return m.GetByUserIDFn(ctx, userID)
}
func (m *MockAPIKeyRepository) Delete(ctx context.Context, id, userID int64) (bool, error) {
	return m.DeleteFn(ctx, id, userID)
}
func (m *MockAPIKeyRepository) ValidateKey(ctx context.Context, keyHash string) (int64, error) {
	return m.ValidateKeyFn(ctx, keyHash)
}
func (m *MockAPIKeyRepository) CleanupExpiredLoginKeys(ctx context.Context, userID int64) (int64, error) {
	return m.CleanupExpiredLoginKeysFn(ctx, userID)
}

// ---- ReminderConfigRepository mock ----

type MockReminderConfigRepository struct {
	CreateFn             func(ctx context.Context, cfg *models.UserReminderConfig) (*models.UserReminderConfig, error)
	GetByIDFn            func(ctx context.Context, id, userID int64) (*models.UserReminderConfig, error)
	GetByUserIDFn        func(ctx context.Context, userID int64) ([]models.UserReminderConfig, error)
	UpdateFn             func(ctx context.Context, id, userID int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error)
	DeleteFn             func(ctx context.Context, id, userID int64) (bool, error)
	HasEnabledByUserIDFn func(ctx context.Context, userID int64) (bool, error)
	ListAllFn            func(ctx context.Context, page, limit int) ([]models.UserReminderConfig, int64, error)
	AdminToggleEnabledFn func(ctx context.Context, id int64) (bool, error)
	AdminDeleteCfgFn     func(ctx context.Context, id int64) (bool, error)
}

func (m *MockReminderConfigRepository) Create(ctx context.Context, cfg *models.UserReminderConfig) (*models.UserReminderConfig, error) {
	return m.CreateFn(ctx, cfg)
}
func (m *MockReminderConfigRepository) GetByID(ctx context.Context, id, userID int64) (*models.UserReminderConfig, error) {
	return m.GetByIDFn(ctx, id, userID)
}
func (m *MockReminderConfigRepository) GetByUserID(ctx context.Context, userID int64) ([]models.UserReminderConfig, error) {
	return m.GetByUserIDFn(ctx, userID)
}
func (m *MockReminderConfigRepository) Update(ctx context.Context, id, userID int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error) {
	return m.UpdateFn(ctx, id, userID, req)
}
func (m *MockReminderConfigRepository) Delete(ctx context.Context, id, userID int64) (bool, error) {
	return m.DeleteFn(ctx, id, userID)
}
func (m *MockReminderConfigRepository) HasEnabledByUserID(ctx context.Context, userID int64) (bool, error) {
	return m.HasEnabledByUserIDFn(ctx, userID)
}
func (m *MockReminderConfigRepository) ListAll(ctx context.Context, page, limit int) ([]models.UserReminderConfig, int64, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx, page, limit)
	}
	return nil, 0, nil
}
func (m *MockReminderConfigRepository) AdminToggleEnabled(ctx context.Context, id int64) (bool, error) {
	if m.AdminToggleEnabledFn != nil {
		return m.AdminToggleEnabledFn(ctx, id)
	}
	return false, nil
}
func (m *MockReminderConfigRepository) AdminDelete(ctx context.Context, id int64) (bool, error) {
	if m.AdminDeleteCfgFn != nil {
		return m.AdminDeleteCfgFn(ctx, id)
	}
	return false, nil
}

// ---- ReminderLogRepository mock ----

type MockReminderLogRepository struct {
	UpsertFn                 func(ctx context.Context, p repository.CreateReminderLogParams) error
	HasResultForTaskConfigFn func(ctx context.Context, taskID, configID int64) (bool, error)
	ListByUserIDFn           func(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error)
	DeleteByTaskIDFn         func(ctx context.Context, taskID int64) error
	ListAllFn                func(ctx context.Context, page, limit int) ([]models.ReminderLog, int64, error)
	AdminListFilteredFn      func(ctx context.Context, page, limit int, userID int64, status string) ([]models.ReminderLog, int64, error)
}

func (m *MockReminderLogRepository) Upsert(ctx context.Context, p repository.CreateReminderLogParams) error {
	return m.UpsertFn(ctx, p)
}
func (m *MockReminderLogRepository) HasResultForTaskConfig(ctx context.Context, taskID, configID int64) (bool, error) {
	return m.HasResultForTaskConfigFn(ctx, taskID, configID)
}
func (m *MockReminderLogRepository) ListByUserID(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error) {
	return m.ListByUserIDFn(ctx, userID, page, limit)
}
func (m *MockReminderLogRepository) DeleteByTaskID(ctx context.Context, taskID int64) error {
	return m.DeleteByTaskIDFn(ctx, taskID)
}
func (m *MockReminderLogRepository) ListAll(ctx context.Context, page, limit int) ([]models.ReminderLog, int64, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx, page, limit)
	}
	return nil, 0, nil
}
func (m *MockReminderLogRepository) AdminListFiltered(ctx context.Context, page, limit int, userID int64, status string) ([]models.ReminderLog, int64, error) {
	if m.AdminListFilteredFn != nil {
		return m.AdminListFilteredFn(ctx, page, limit, userID, status)
	}
	return nil, 0, nil
}
