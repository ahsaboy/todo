package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"todo/internal/config"
)

type GitHubProvider struct {
	cfg config.OAuthProviderConfig
}

func NewGitHubProvider(cfg config.OAuthProviderConfig) *GitHubProvider {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"read:user", "user:email"}
	}
	return &GitHubProvider{cfg: cfg}
}

func (p *GitHubProvider) Name() string { return "github" }

func (p *GitHubProvider) IsEnabled() bool {
	return p.cfg.Enabled && p.cfg.ClientID != "" && p.cfg.ClientSecret != ""
}

func (p *GitHubProvider) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		Scopes:       p.cfg.Scopes,
		Endpoint:     github.Endpoint,
	}
}

func (p *GitHubProvider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// 获取用户基本信息
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user API returned %d", userResp.StatusCode)
	}

	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode github user: %w", err)
	}

	email := user.Email

	// GitHub 可能不公开邮箱，需要额外查询
	if email == "" {
		email = p.fetchPrimaryEmail(client)
	}

	displayName := user.Name
	if displayName == "" {
		displayName = user.Login
	}

	return &UserInfo{
		Provider:   "github",
		ProviderID: fmt.Sprintf("%d", user.ID),
		Username:   user.Login,
		Email:      email,
		AvatarURL:  user.AvatarURL,
	}, nil
}

func (p *GitHubProvider) fetchPrimaryEmail(client *http.Client) string {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email
		}
	}
	return ""
}
