package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"todo/internal/repository"
	"todo/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(apiKeyRepo repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := extractAPIKey(c)

		if key == "" {
			utils.RespondError(c, http.StatusUnauthorized, "missing API key (use Authorization: Bearer <key> or api-key: <key>)", utils.CodeUnauthorized)
			c.Abort()
			return
		}

		hash := HashAPIKey(key)
		userID, err := apiKeyRepo.ValidateKey(c.Request.Context(), hash)
		if err != nil || userID <= 0 {
			utils.RespondError(c, http.StatusUnauthorized, "invalid API key", utils.CodeUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func extractAPIKey(c *gin.Context) string {
	// 1. Authorization: Bearer <key>
	if auth := c.GetHeader("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// 2. api-key: <key>
	if key := c.GetHeader("api-key"); key != "" {
		return key
	}

	// 3. X-API-Key: <key> (legacy)
	return c.GetHeader("X-API-Key")
}

func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func GetUserID(c *gin.Context) (int64, bool) {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := uid.(int64)
	return id, ok
}
