package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

// UserInfo 是从 OAuth 提供商获取的用户信息。
type UserInfo struct {
	Provider   string
	ProviderID string
	Username   string
	Email      string
	AvatarURL  string
}

// Provider 定义 OAuth 提供商的统一接口。
type Provider interface {
	Name() string
	IsEnabled() bool
	OAuth2Config() *oauth2.Config
	FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}
