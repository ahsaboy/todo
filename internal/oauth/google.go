package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"todo/internal/config"
)

type GoogleProvider struct {
	cfg config.OAuthProviderConfig
}

func NewGoogleProvider(cfg config.OAuthProviderConfig) *GoogleProvider {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid", "email", "profile"}
	}
	return &GoogleProvider{cfg: cfg}
}

func (p *GoogleProvider) Name() string { return "google" }

func (p *GoogleProvider) IsEnabled() bool {
	return p.cfg.Enabled && p.cfg.ClientID != "" && p.cfg.ClientSecret != ""
}

func (p *GoogleProvider) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		Scopes:       p.cfg.Scopes,
		Endpoint:     google.Endpoint,
	}
}

func (p *GoogleProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetch google user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo API returned %d", resp.StatusCode)
	}

	var user struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Picture   string `json:"picture"`
		Verified  bool   `json:"verified_email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode google user: %w", err)
	}

	return &UserInfo{
		Provider:   "google",
		ProviderID: user.ID,
		Username:   user.Email, // Google 没有 username，用 email 前缀
		Email:      user.Email,
		AvatarURL:  user.Picture,
	}, nil
}
