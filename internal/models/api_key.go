package models

type APIKey struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	KeyHash    string `json:"-"`
	Name       string `json:"name"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string `json:"created_at"`
}

type CreateKeyRequest struct {
	Name string `json:"name" binding:"omitempty,max=64"`
}

type APIKeyResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"` // 仅创建时返回明文
	CreatedAt string `json:"created_at"`
}

type APIKeyInfo struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
}
