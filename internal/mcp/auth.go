package mcp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"todo/internal/repository"
)

// apiKeyAuthMiddleware 从请求头(Authorization: Bearer / api-key / X-API-Key)取出 API Key,
// SHA-256 哈希后用 APIKeyRepo.ValidateKey 校验。
// 失败时写入 401 JSON(格式对齐 internal/utils/response.go);成功时把 user_id 注入 ctx。
func apiKeyAuthMiddleware(repo *repository.APIKeyRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractAPIKey(r)
			if key == "" {
				writeUnauthorized(w, "missing API key (use Authorization: Bearer <key> or api-key: <key>)")
				return
			}

			hash := hashAPIKey(key)
			userID, err := repo.ValidateKey(r.Context(), hash)
			if err != nil || userID <= 0 {
				writeUnauthorized(w, "invalid API key")
				return
			}

			ctx := WithUserID(r.Context(), userID)
			ctx = WithStructuredOutput(ctx, headerEnabled(r.Header.Get("X-MCP-Structured-Output")))
			ctx = WithRemindersEnabled(ctx, headerEnabled(r.Header.Get("X-MCP-Include-Reminders")))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractAPIKey 与 internal/middleware/auth.go 中的 extractAPIKey 保持一致的三种来源:
//  1. Authorization: Bearer <key>
//  2. api-key: <key>
//  3. X-API-Key: <key>(legacy)
func extractAPIKey(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if key := r.Header.Get("api-key"); key != "" {
		return key
	}
	return r.Header.Get("X-API-Key")
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// writeUnauthorized 返回与 utils.ErrorResponse 一致的 401 JSON 结构。
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error":   message,
		"code":    "UNAUTHORIZED",
	})
}
