package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"todo/internal/models"
	"todo/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   repository.UserRepository
	apiKeyRepo repository.APIKeyRepository
}

func NewAuthService(userRepo repository.UserRepository, apiKeyRepo repository.APIKeyRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, string, error) {
	// 检查用户名是否已存在
	existing, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, "", ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, req.Username, req.Email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrUsernameTaken) {
			return nil, "", ErrUsernameTaken
		}
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	// 自动生成首个 API Key
	apiKey, err := s.generateKey(ctx, user.ID, "default")
	if err != nil {
		return nil, "", fmt.Errorf("generate api key: %w", err)
	}

	resp := user.ToResponse()
	return &resp, apiKey, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.UserResponse, string, error) {
	// 支持用户名或邮箱登录
	user, err := s.userRepo.GetByUsername(ctx, req.Account)
	if err != nil {
		return nil, "", fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		user, err = s.userRepo.GetByEmail(ctx, req.Account)
		if err != nil {
			return nil, "", fmt.Errorf("get user: %w", err)
		}
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// 清理过期的 login API Key（last_used_at 超过 24 小时）
	s.apiKeyRepo.CleanupExpiredLoginKeys(ctx, user.ID)

	// 登录时生成新的 API Key
	apiKey, err := s.generateKey(ctx, user.ID, "login")
	if err != nil {
		return nil, "", fmt.Errorf("generate api key: %w", err)
	}

	resp := user.ToResponse()
	return &resp, apiKey, nil
}

func (s *AuthService) GenerateAPIKey(ctx context.Context, userID int64, name string) (string, error) {
	if name == "" {
		name = "default"
	}
	return s.generateKey(ctx, userID, name)
}

func (s *AuthService) RevokeAPIKey(ctx context.Context, id, userID int64) (bool, error) {
	return s.apiKeyRepo.Delete(ctx, id, userID)
}

func (s *AuthService) generateKey(ctx context.Context, userID int64, name string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random key: %w", err)
	}
	apiKey := base64.URLEncoding.EncodeToString(b)

	h := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(h[:])

	_, err := s.apiKeyRepo.Create(ctx, userID, keyHash, name)
	if err != nil {
		return "", fmt.Errorf("save api key: %w", err)
	}

	return apiKey, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, email string) error {
	return s.userRepo.UpdateProfile(ctx, userID, email)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidOldPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(hash))
}

func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) ListAPIKeys(ctx context.Context, userID int64) ([]models.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(ctx, userID)
}
