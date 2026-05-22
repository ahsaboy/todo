package service_test

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/internal/testutil"
)

func bcryptHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(hash)
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
		CreateFn: func(_ context.Context, username, email, _ string) (*models.User, error) {
			return &models.User{ID: 1, Username: username, Email: email}, nil
		},
	}
	apiKeyRepo := &testutil.MockAPIKeyRepository{
		CreateFn: func(_ context.Context, userID int64, _, name string) (*models.APIKey, error) {
			return &models.APIKey{ID: 1, UserID: userID, Name: name}, nil
		},
	}
	svc := service.NewAuthService(userRepo, apiKeyRepo)

	resp, key, err := svc.Register(context.Background(), models.RegisterRequest{
		Username: "alice", Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Username != "alice" {
		t.Error("expected user response with username 'alice'")
	}
	if key == "" {
		t.Error("expected non-empty API key")
	}
}

func TestAuthService_Register_UsernameTaken_PreCheck(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) {
			return &models.User{ID: 1, Username: "alice"}, nil
		},
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	_, _, err := svc.Register(context.Background(), models.RegisterRequest{
		Username: "alice", Password: "password123",
	})
	if !errors.Is(err, service.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestAuthService_Register_UsernameTaken_DBConflict(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
		CreateFn: func(_ context.Context, _, _, _ string) (*models.User, error) {
			return nil, repository.ErrUsernameTaken
		},
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	_, _, err := svc.Register(context.Background(), models.RegisterRequest{
		Username: "alice", Password: "password123",
	})
	if !errors.Is(err, service.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	hash := bcryptHash(t, "password123")
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, username string) (*models.User, error) {
			return &models.User{ID: 1, Username: username, PasswordHash: hash}, nil
		},
	}
	apiKeyRepo := &testutil.MockAPIKeyRepository{
		CleanupExpiredLoginKeysFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
		CreateFn: func(_ context.Context, userID int64, _, name string) (*models.APIKey, error) {
			return &models.APIKey{ID: 1, UserID: userID, Name: name}, nil
		},
	}
	svc := service.NewAuthService(userRepo, apiKeyRepo)

	resp, key, err := svc.Login(context.Background(), models.LoginRequest{
		Account: "alice", Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil user response")
	}
	if key == "" {
		t.Error("expected non-empty API key")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
		GetByEmailFn:    func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	_, _, err := svc.Login(context.Background(), models.LoginRequest{
		Account: "ghost", Password: "pass",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hash := bcryptHash(t, "correct")
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, username string) (*models.User, error) {
			return &models.User{ID: 1, Username: username, PasswordHash: hash}, nil
		},
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	_, _, err := svc.Login(context.Background(), models.LoginRequest{
		Account: "alice", Password: "wrong",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_ByEmail(t *testing.T) {
	hash := bcryptHash(t, "pass")
	userRepo := &testutil.MockUserRepository{
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
		GetByEmailFn: func(_ context.Context, email string) (*models.User, error) {
			return &models.User{ID: 1, Username: "alice", Email: email, PasswordHash: hash}, nil
		},
	}
	apiKeyRepo := &testutil.MockAPIKeyRepository{
		CleanupExpiredLoginKeysFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
		CreateFn: func(_ context.Context, userID int64, _, name string) (*models.APIKey, error) {
			return &models.APIKey{ID: 1, UserID: userID}, nil
		},
	}
	svc := service.NewAuthService(userRepo, apiKeyRepo)

	resp, _, err := svc.Login(context.Background(), models.LoginRequest{
		Account: "alice@example.com", Password: "pass",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil user response")
	}
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	hash := bcryptHash(t, "oldpass")
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: hash}, nil
		},
		UpdatePasswordFn: func(_ context.Context, _ int64, _ string) error { return nil },
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	if err := svc.ChangePassword(context.Background(), 1, "oldpass", "newpass"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, _ int64) (*models.User, error) { return nil, nil },
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	if err := svc.ChangePassword(context.Background(), 99, "old", "new"); !errors.Is(err, service.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	hash := bcryptHash(t, "correctold")
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: hash}, nil
		},
	}
	svc := service.NewAuthService(userRepo, &testutil.MockAPIKeyRepository{})

	if err := svc.ChangePassword(context.Background(), 1, "wrongold", "new"); !errors.Is(err, service.ErrInvalidOldPassword) {
		t.Errorf("expected ErrInvalidOldPassword, got %v", err)
	}
}
