package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"todo/internal/config"
	"todo/internal/utils"
)

func AdminAuthMiddleware(cfg config.AdminConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-Admin-Token")
		if raw == "" {
			utils.RespondError(c, 401, "missing admin token", utils.CodeUnauthorized)
			c.Abort()
			return
		}
		h := sha256.Sum256([]byte(raw))
		if subtle.ConstantTimeCompare([]byte(hex.EncodeToString(h[:])), []byte(cfg.TokenHash)) != 1 {
			utils.RespondError(c, 401, "invalid admin token", utils.CodeUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}
