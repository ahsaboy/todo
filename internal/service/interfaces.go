package service

import (
	"context"

	"todo/internal/models"
	"todo/internal/oauth"
)

type TaskServiceInterface interface {
	Create(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error)
	GetByID(ctx context.Context, userID, id int64) (*models.Task, error)
	List(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error)
	Update(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error)
	Delete(ctx context.Context, userID, id int64) (bool, error)
	ToggleComplete(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error)
}

type AuthServiceInterface interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, string, error)
	Login(ctx context.Context, req models.LoginRequest) (*models.UserResponse, string, error)
	GenerateAPIKey(ctx context.Context, userID int64, name string) (string, error)
	RevokeAPIKey(ctx context.Context, id, userID int64) (bool, error)
	UpdateProfile(ctx context.Context, userID int64, email string) error
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ResetPassword(ctx context.Context, userID int64, newPassword string) error
	ListAPIKeys(ctx context.Context, userID int64) ([]models.APIKey, error)
	HasPassword(ctx context.Context, userID int64) (bool, error)
}

type EmailServiceInterface interface {
	IsEnabled() bool
	SetEnabled(b bool)
	SendVerificationCode(ctx context.Context, email, purpose string) error
	VerifyCode(ctx context.Context, email, code, purpose string) error
	TestConnection(ctx context.Context) error
}

type ReminderConfigServiceInterface interface {
	Create(ctx context.Context, userID int64, req models.CreateReminderConfigRequest) (*models.UserReminderConfig, error)
	GetByID(ctx context.Context, userID, id int64) (*models.UserReminderConfig, error)
	List(ctx context.Context, userID int64) ([]models.UserReminderConfig, error)
	Update(ctx context.Context, userID, id int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error)
	Delete(ctx context.Context, userID, id int64) (bool, error)
	HasEnabled(ctx context.Context, userID int64) (bool, error)
}

type ReminderLogServiceInterface interface {
	List(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error)
}

type OAuthServiceInterface interface {
	HandleCallback(ctx context.Context, providerName string, code string) (*models.UserResponse, string, error)
	GetAvailableProviders() []oauth.ProviderDisplayInfo
	ListUserAccounts(ctx context.Context, userID int64) ([]models.OAuthAccount, error)
	LinkAccount(ctx context.Context, userID int64, providerName, code string) error
	UnlinkAccount(ctx context.Context, userID, accountID int64) error
}
