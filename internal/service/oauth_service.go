package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"todo/internal/models"
	"todo/internal/oauth"
	"todo/internal/repository"
)

type OAuthService struct {
	userRepo   repository.UserRepository
	oauthRepo  repository.OAuthAccountRepository
	apiKeyRepo repository.APIKeyRepository
	oauthReg   *oauth.Registry
	authSvc    *AuthService
}

func NewOAuthService(
	userRepo repository.UserRepository,
	oauthRepo repository.OAuthAccountRepository,
	apiKeyRepo repository.APIKeyRepository,
	oauthReg *oauth.Registry,
	authSvc *AuthService,
) *OAuthService {
	return &OAuthService{
		userRepo:   userRepo,
		oauthRepo:  oauthRepo,
		apiKeyRepo: apiKeyRepo,
		oauthReg:   oauthReg,
		authSvc:    authSvc,
	}
}

// GetAvailableProviders 返回已启用的 OAuth provider 展示信息。
func (s *OAuthService) GetAvailableProviders() []oauth.ProviderDisplayInfo {
	return s.oauthReg.GetDisplayInfo()
}

// HandleCallback 处理 OAuth 回调：换取 token → 获取用户信息 → 创建/关联用户 → 生成 API Key。
func (s *OAuthService) HandleCallback(ctx context.Context, providerName string, code string) (*models.UserResponse, string, error) {
	// 整个 OAuth 流程 10 秒超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	provider, ok := s.oauthReg.Get(providerName)
	if !ok || !provider.IsEnabled() {
		return nil, "", fmt.Errorf("provider %s not available", providerName)
	}

	// 用 code 换取 token
	token, err := provider.OAuth2Config().Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchange code: %w", err)
	}

	// 获取用户信息
	info, err := provider.FetchUserInfo(ctx, token)
	if err != nil {
		return nil, "", fmt.Errorf("fetch user info: %w", err)
	}

	// 1. 查找已关联的 OAuth 账号
	existing, err := s.oauthRepo.GetByProvider(ctx, providerName, info.ProviderID)
	if err != nil {
		return nil, "", fmt.Errorf("check oauth account: %w", err)
	}

	if existing != nil {
		user, err := s.userRepo.GetByID(ctx, existing.UserID)
		if err != nil || user == nil {
			return nil, "", fmt.Errorf("get linked user: %w", err)
		}
		apiKey, err := s.generateOAuthLoginKey(ctx, user.ID, providerName)
		if err != nil {
			return nil, "", err
		}
		resp := user.ToResponse()
		return &resp, apiKey, nil
	}

	// 2. 未关联，尝试邮箱匹配
	var user *models.User
	if info.Email != "" {
		user, err = s.userRepo.GetByEmail(ctx, info.Email)
		if err != nil {
			return nil, "", fmt.Errorf("check email: %w", err)
		}
	}

	if user != nil {
		if err := s.oauthRepo.Create(ctx, user.ID, providerName, info.ProviderID, info.Username, info.AvatarURL); err != nil {
			return nil, "", fmt.Errorf("link oauth account: %w", err)
		}
		apiKey, err := s.generateOAuthLoginKey(ctx, user.ID, providerName)
		if err != nil {
			return nil, "", err
		}
		resp := user.ToResponse()
		return &resp, apiKey, nil
	}

	// 3. 全新用户 → 创建（带重试以防用户名冲突）
	for attempt := 0; attempt < 5; attempt++ {
		username := s.generateOAuthUsername(providerName, info.Username)
		user, err = s.userRepo.CreateOAuthUser(ctx, username, info.Email, info.AvatarURL)
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, "", fmt.Errorf("create oauth user: %w", err)
		}
	}
	if user == nil {
		return nil, "", fmt.Errorf("create oauth user: failed after retries")
	}

	if err := s.oauthRepo.Create(ctx, user.ID, providerName, info.ProviderID, info.Username, info.AvatarURL); err != nil {
		return nil, "", fmt.Errorf("link oauth account: %w", err)
	}

	apiKey, err := s.generateOAuthLoginKey(ctx, user.ID, providerName)
	if err != nil {
		return nil, "", err
	}

	resp := user.ToResponse()
	return &resp, apiKey, nil
}

// generateOAuthLoginKey 清理旧 key 后生成新的 OAuth 登录 key。
func (s *OAuthService) generateOAuthLoginKey(ctx context.Context, userID int64, providerName string) (string, error) {
	s.cleanOldOAuthKey(ctx, userID, providerName)
	apiKey, err := s.authSvc.GenerateLoginKey(ctx, userID, "oauth-"+providerName)
	if err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return apiKey, nil
}

// cleanOldOAuthKey 删除该用户指定 provider 的旧 OAuth API Key，避免无限积累。
func (s *OAuthService) cleanOldOAuthKey(ctx context.Context, userID int64, providerName string) {
	keys, err := s.apiKeyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return
	}
	targetName := "oauth-" + providerName
	for _, k := range keys {
		if k.Name == targetName {
			s.apiKeyRepo.Delete(ctx, k.ID, userID)
		}
	}
}

// generateOAuthUsername 生成唯一用户名，格式：{prefix}_{random8}。
func (s *OAuthService) generateOAuthUsername(provider, name string) string {
	prefix := provider
	if name != "" {
		name = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
				return r
			}
			return '_'
		}, name)
		if len(name) > 16 {
			name = name[:16]
		}
		prefix = name
	}

	b := make([]byte, 4)
	rand.Read(b)
	suffix := hex.EncodeToString(b)

	return fmt.Sprintf("%s_%s", prefix, suffix)
}

// ListUserAccounts 返回指定用户的所有 OAuth 绑定。
func (s *OAuthService) ListUserAccounts(ctx context.Context, userID int64) ([]models.OAuthAccount, error) {
	return s.oauthRepo.GetByUserID(ctx, userID)
}

// LinkAccount 将 OAuth 账号绑定到已认证用户。
func (s *OAuthService) LinkAccount(ctx context.Context, userID int64, providerName, code string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	provider, ok := s.oauthReg.Get(providerName)
	if !ok || !provider.IsEnabled() {
		return fmt.Errorf("provider %s not available", providerName)
	}

	token, err := provider.OAuth2Config().Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("exchange code: %w", err)
	}

	info, err := provider.FetchUserInfo(ctx, token)
	if err != nil {
		return fmt.Errorf("fetch user info: %w", err)
	}

	existing, err := s.oauthRepo.GetByProvider(ctx, providerName, info.ProviderID)
	if err != nil {
		return fmt.Errorf("check oauth account: %w", err)
	}
	if existing != nil {
		if existing.UserID == userID {
			return ErrOAuthAlreadyLinked
		}
		return ErrOAuthLinkedToOther
	}

	if err := s.oauthRepo.Create(ctx, userID, providerName, info.ProviderID, info.Username, info.AvatarURL); err != nil {
		return fmt.Errorf("link oauth account: %w", err)
	}

	// 若用户无头像，自动使用 OAuth 头像
	if info.AvatarURL != "" {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err == nil && user != nil && user.AvatarURL == "" {
			_ = s.userRepo.UpdateAvatar(ctx, userID, info.AvatarURL)
		}
	}

	return nil
}

// UnlinkAccount 解除 OAuth 绑定，保证至少保留一种登录方式。
func (s *OAuthService) UnlinkAccount(ctx context.Context, userID, accountID int64) error {
	accounts, err := s.oauthRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("list oauth accounts: %w", err)
	}

	found := false
	for _, a := range accounts {
		if a.ID == accountID {
			found = true
			break
		}
	}
	if !found {
		return ErrOAuthAccountNotFound
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return ErrOAuthAccountNotFound
	}

	hasPassword := user.PasswordHash != ""
	remainingOAuth := len(accounts) - 1
	if !hasPassword && remainingOAuth == 0 {
		return ErrLastAuthMethod
	}

	_, err = s.oauthRepo.Delete(ctx, accountID, userID)
	return err
}
