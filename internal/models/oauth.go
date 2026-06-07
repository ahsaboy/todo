package models

type OAuthAccount struct {
	ID             int64  `json:"id"`
	UserID         int64  `json:"user_id"`
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
	DisplayName    string `json:"display_name"`
	AvatarURL      string `json:"avatar_url"`
	CreatedAt      string `json:"created_at"`
}

// OAuthAccountResponse 返回给前端的 OAuth 账号信息（隐藏 user_id 和 provider_user_id）。
type OAuthAccountResponse struct {
	ID          int64  `json:"id"`
	Provider    string `json:"provider"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	LinkedAt    string `json:"linked_at"`
}

// ProfileResponse 聚合个人资料页所需的全部数据。
type ProfileResponse struct {
	User          UserResponse           `json:"user"`
	OAuthAccounts []OAuthAccountResponse `json:"oauth_accounts"`
	HasPassword   bool                   `json:"has_password"`
}

func (a *OAuthAccount) ToResponse() OAuthAccountResponse {
	return OAuthAccountResponse{
		ID:          a.ID,
		Provider:    a.Provider,
		DisplayName: a.DisplayName,
		AvatarURL:   a.AvatarURL,
		LinkedAt:    a.CreatedAt,
	}
}
