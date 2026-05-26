package repository

import (
	"context"
	"time"

	"todo/internal/models"
)

type TaskRepository interface {
	Create(ctx context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error)
	GetByID(ctx context.Context, userID, id int64) (*models.Task, error)
	List(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int, sortField, sortOrder string) ([]models.Task, int64, error)
	Update(ctx context.Context, userID, id int64, req models.UpdateTaskRequest) (*models.Task, error)
	Delete(ctx context.Context, userID, id int64) (bool, error)
	ToggleComplete(ctx context.Context, userID, id int64, focusDuration *int) (*models.Task, error)
	ToggleCompleteAndCreateRepeat(ctx context.Context, userID, id int64, next *models.Task, focusDuration *int) (*models.Task, error)
	GetPendingReminders(ctx context.Context, now time.Time) ([]models.Task, error)
	MarkReminderSent(ctx context.Context, id int64) (bool, error)
	CreateRepeatTask(ctx context.Context, t *models.Task) error
	// Admin methods: userID=0 queries all users, >0 queries that user only
	ListAll(ctx context.Context, userID int64, filters models.TaskFilters, page, limit int) ([]models.Task, int64, error)
	AdminDelete(ctx context.Context, id int64) (bool, error)
}

type UserRepository interface {
	Create(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateProfile(ctx context.Context, id int64, email string) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	// Admin methods
	ListAll(ctx context.Context, page, limit int, search string) ([]models.User, int64, error)
	Delete(ctx context.Context, id int64) error
	ForceResetPassword(ctx context.Context, id int64, passwordHash string) error
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	SetIsAdmin(ctx context.Context, userID int64, isAdmin bool) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, userID int64, keyHash, name string) (*models.APIKey, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.APIKey, error)
	Delete(ctx context.Context, id, userID int64) (bool, error)
	ValidateKey(ctx context.Context, keyHash string) (int64, error)
	CleanupExpiredLoginKeys(ctx context.Context, userID int64) (int64, error)
}

type ReminderConfigRepository interface {
	Create(ctx context.Context, cfg *models.UserReminderConfig) (*models.UserReminderConfig, error)
	GetByID(ctx context.Context, id, userID int64) (*models.UserReminderConfig, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.UserReminderConfig, error)
	Update(ctx context.Context, id, userID int64, req models.UpdateReminderConfigRequest) (*models.UserReminderConfig, error)
	Delete(ctx context.Context, id, userID int64) (bool, error)
	HasEnabledByUserID(ctx context.Context, userID int64) (bool, error)
	// Admin methods
	ListAll(ctx context.Context, page, limit int) ([]models.UserReminderConfig, int64, error)
}

type ReminderLogRepository interface {
	Upsert(ctx context.Context, p CreateReminderLogParams) error
	HasResultForTaskConfig(ctx context.Context, taskID, configID int64) (bool, error)
	ListByUserID(ctx context.Context, userID int64, page, limit int) ([]models.ReminderLog, int64, error)
	DeleteByTaskID(ctx context.Context, taskID int64) error
	// Admin methods
	ListAll(ctx context.Context, page, limit int) ([]models.ReminderLog, int64, error)
}

type TagRepository interface {
	Create(ctx context.Context, t *models.UserTag) (*models.UserTag, error)
	GetByID(ctx context.Context, id, userID int64) (*models.UserTag, error)
	GetByName(ctx context.Context, userID int64, name string) (*models.UserTag, error)
	ListByUserID(ctx context.Context, userID int64) ([]models.UserTag, error)
	GetNamesSet(ctx context.Context, userID int64) (map[string]struct{}, error)
	Update(ctx context.Context, id, userID int64, name *string, color *string, icon *string, sortOrder *int) (*models.UserTag, error)
	Delete(ctx context.Context, id, userID int64) (bool, error)
	RenameWithTaskSync(ctx context.Context, id, userID int64, newName string) (*models.UserTag, error)
	DeleteWithTaskSync(ctx context.Context, id, userID int64) (bool, int64, error)
}
