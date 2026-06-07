package testutil

import (
	"context"

	"todo/internal/models"
)

// ---- TaskServiceInterface mock ----

type MockTaskService struct {
	CreateFn         func(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error)
	GetByIDFn        func(ctx context.Context, userID, id int64) (*models.Task, error)
	ListFn           func(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error)
	UpdateFn         func(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error)
	DeleteFn         func(ctx context.Context, userID, id int64) (bool, error)
	ToggleCompleteFn func(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error)
}

func (m *MockTaskService) Create(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
	return m.CreateFn(ctx, userID, req)
}
func (m *MockTaskService) GetByID(ctx context.Context, userID, id int64) (*models.Task, error) {
	return m.GetByIDFn(ctx, userID, id)
}
func (m *MockTaskService) List(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error) {
	return m.ListFn(ctx, userID, filters, page, limit, sortField, sortOrder)
}
func (m *MockTaskService) Update(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error) {
	return m.UpdateFn(ctx, userID, id, req)
}
func (m *MockTaskService) Delete(ctx context.Context, userID, id int64) (bool, error) {
	return m.DeleteFn(ctx, userID, id)
}
func (m *MockTaskService) ToggleComplete(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error) {
	return m.ToggleCompleteFn(ctx, userID, id, focusDuration)
}

// ---- AuthServiceInterface mock ----

type MockAuthService struct {
	RegisterFn         func(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, string, error)
	LoginFn            func(ctx context.Context, req models.LoginRequest) (*models.UserResponse, string, error)
	GenerateAPIKeyFn   func(ctx context.Context, userID int64, name string) (string, error)
	RevokeAPIKeyFn     func(ctx context.Context, id, userID int64) (bool, error)
	UpdateProfileFn    func(ctx context.Context, userID int64, email string) error
	ChangePasswordFn   func(ctx context.Context, userID int64, oldPassword, newPassword string) error
	GetUserByIDFn      func(ctx context.Context, id int64) (*models.User, error)
	GetUserByEmailFn   func(ctx context.Context, email string) (*models.User, error)
	ResetPasswordFn    func(ctx context.Context, userID int64, newPassword string) error
	ListAPIKeysFn      func(ctx context.Context, userID int64) ([]models.APIKey, error)
	HasPasswordFn      func(ctx context.Context, userID int64) (bool, error)
}

func (m *MockAuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, string, error) {
	return m.RegisterFn(ctx, req)
}
func (m *MockAuthService) Login(ctx context.Context, req models.LoginRequest) (*models.UserResponse, string, error) {
	return m.LoginFn(ctx, req)
}
func (m *MockAuthService) GenerateAPIKey(ctx context.Context, userID int64, name string) (string, error) {
	return m.GenerateAPIKeyFn(ctx, userID, name)
}
func (m *MockAuthService) RevokeAPIKey(ctx context.Context, id, userID int64) (bool, error) {
	return m.RevokeAPIKeyFn(ctx, id, userID)
}
func (m *MockAuthService) UpdateProfile(ctx context.Context, userID int64, email string) error {
	return m.UpdateProfileFn(ctx, userID, email)
}
func (m *MockAuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	return m.ChangePasswordFn(ctx, userID, oldPassword, newPassword)
}
func (m *MockAuthService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return m.GetUserByIDFn(ctx, id)
}
func (m *MockAuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return nil, nil
}
func (m *MockAuthService) ResetPassword(ctx context.Context, userID int64, newPassword string) error {
	if m.ResetPasswordFn != nil {
		return m.ResetPasswordFn(ctx, userID, newPassword)
	}
	return nil
}
func (m *MockAuthService) ListAPIKeys(ctx context.Context, userID int64) ([]models.APIKey, error) {
	return m.ListAPIKeysFn(ctx, userID)
}
func (m *MockAuthService) HasPassword(ctx context.Context, userID int64) (bool, error) {
	if m.HasPasswordFn != nil {
		return m.HasPasswordFn(ctx, userID)
	}
	return true, nil
}

// ---- ReminderConfigServiceInterface mock ----

type MockReminderConfigService struct {
	CreateFn     func(ctx context.Context, userID int64, req models.CreateReminderConfigRequest) (*models.UserReminderConfig, error)
	GetByIDFn    func(ctx context.Context, userID, id int64) (*models.UserReminderConfig, error)
	ListFn       func(ctx context.Context, userID int64) ([]models.UserReminderConfig, error)
	UpdateFn     func(ctx context.Context, userID, id int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error)
	DeleteFn     func(ctx context.Context, userID, id int64) (bool, error)
	HasEnabledFn func(ctx context.Context, userID int64) (bool, error)
}

func (m *MockReminderConfigService) Create(ctx context.Context, userID int64, req models.CreateReminderConfigRequest) (*models.UserReminderConfig, error) {
	return m.CreateFn(ctx, userID, req)
}
func (m *MockReminderConfigService) GetByID(ctx context.Context, userID, id int64) (*models.UserReminderConfig, error) {
	return m.GetByIDFn(ctx, userID, id)
}
func (m *MockReminderConfigService) List(ctx context.Context, userID int64) ([]models.UserReminderConfig, error) {
	return m.ListFn(ctx, userID)
}
func (m *MockReminderConfigService) Update(ctx context.Context, userID, id int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error) {
	return m.UpdateFn(ctx, userID, id, req)
}
func (m *MockReminderConfigService) Delete(ctx context.Context, userID, id int64) (bool, error) {
	return m.DeleteFn(ctx, userID, id)
}
func (m *MockReminderConfigService) HasEnabled(ctx context.Context, userID int64) (bool, error) {
	return m.HasEnabledFn(ctx, userID)
}

// ---- ReminderLogServiceInterface mock ----

type MockReminderLogService struct {
	ListFn func(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error)
}

func (m *MockReminderLogService) List(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error) {
	return m.ListFn(ctx, userID, page, limit)
}

// ---- EmailServiceInterface mock ----

type MockEmailService struct {
	IsEnabledFn              func() bool
	SetEnabledFn             func(b bool)
	SendVerificationCodeFn   func(ctx context.Context, email string) error
	VerifyCodeFn             func(ctx context.Context, email, code string) error
	TestConnectionFn         func(ctx context.Context) error
}

func (m *MockEmailService) IsEnabled() bool {
	if m.IsEnabledFn != nil {
		return m.IsEnabledFn()
	}
	return false
}
func (m *MockEmailService) SetEnabled(b bool) {
	if m.SetEnabledFn != nil {
		m.SetEnabledFn(b)
	}
}
func (m *MockEmailService) SendVerificationCode(ctx context.Context, email string) error {
	return m.SendVerificationCodeFn(ctx, email)
}
func (m *MockEmailService) VerifyCode(ctx context.Context, email, code string) error {
	return m.VerifyCodeFn(ctx, email, code)
}
func (m *MockEmailService) TestConnection(ctx context.Context) error {
	if m.TestConnectionFn != nil {
		return m.TestConnectionFn(ctx)
	}
	return nil
}
