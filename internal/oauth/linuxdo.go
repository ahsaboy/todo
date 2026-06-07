package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"todo/internal/config"
)

const linuxDoDiscoveryURL = "https://connect.linuxdo.net/.well-known/openid-configuration"

type LinuxDoProvider struct {
	cfg            config.OAuthProviderConfig
	endpoint       oauth2.Endpoint
	userInfoURL    string
	tokenAuthStyle oauth2.AuthStyle
}

type oidcDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

func NewLinuxDoProvider(cfg config.OAuthProviderConfig) *LinuxDoProvider {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid", "profile", "email"}
	}
	return &LinuxDoProvider{cfg: cfg}
}

// Init 从 OIDC discovery 文档获取端点信息。应在启动时调用。
func (p *LinuxDoProvider) Init(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, linuxDoDiscoveryURL, nil)
	if err != nil {
		return fmt.Errorf("create discovery request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch linuxdo discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("linuxdo discovery returned %d", resp.StatusCode)
	}

	var disc oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&disc); err != nil {
		return fmt.Errorf("decode linuxdo discovery: %w", err)
	}

	p.endpoint = oauth2.Endpoint{
		AuthURL:  disc.AuthorizationEndpoint,
		TokenURL: disc.TokenEndpoint,
	}
	p.userInfoURL = disc.UserinfoEndpoint
	// LinuxDo 可能需要 client_secret_post 认证方式
	p.tokenAuthStyle = oauth2.AuthStyleInParams

	return nil
}

func (p *LinuxDoProvider) Name() string { return "linuxdo" }

func (p *LinuxDoProvider) IsEnabled() bool {
	return p.cfg.Enabled && p.cfg.ClientID != "" && p.cfg.ClientSecret != ""
}

func (p *LinuxDoProvider) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		Scopes:       p.cfg.Scopes,
		Endpoint:     p.endpoint,
	}
}

func (p *LinuxDoProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// 优先尝试 userinfo endpoint
	if p.userInfoURL != "" {
		return p.fetchFromUserinfoEndpoint(client)
	}

	// fallback: 从 ID token 解析（如果有的话）
	idToken, ok := token.Extra("id_token").(string)
	if ok && idToken != "" {
		return p.parseIDToken(idToken)
	}

	return nil, fmt.Errorf("no userinfo endpoint and no id_token available")
}

func (p *LinuxDoProvider) fetchFromUserinfoEndpoint(client *http.Client) (*UserInfo, error) {
	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("fetch linuxdo userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linuxdo userinfo returned %d", resp.StatusCode)
	}

	var user struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		PreferredUser string `json:"preferred_username"`
		Email         string `json:"email"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode linuxdo userinfo: %w", err)
	}

	username := user.PreferredUser
	if username == "" {
		username = user.Name
	}

	return &UserInfo{
		Provider:   "linuxdo",
		ProviderID: user.Sub,
		Username:   username,
		Email:      user.Email,
		AvatarURL:  user.Picture,
	}, nil
}

// parseIDToken 简单解析 JWT payload（不验证签名，因为来自可信的 token endpoint）。
func (p *LinuxDoProvider) parseIDToken(idToken string) (*UserInfo, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	// base64url 解码 payload
	payload := parts[1]
	// 补齐 padding
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("decode id_token payload: %w", err)
	}

	var claims struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		PreferredUser string `json:"preferred_username"`
		Email         string `json:"email"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}

	username := claims.PreferredUser
	if username == "" {
		username = claims.Name
	}

	return &UserInfo{
		Provider:   "linuxdo",
		ProviderID: claims.Sub,
		Username:   username,
		Email:      claims.Email,
		AvatarURL:  claims.Picture,
	}, nil
}
