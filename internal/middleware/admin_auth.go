package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

// AdminOnlyMiddleware 检查当前用户(由 AuthMiddleware 注入)是否有 is_admin=1。
// 必须在 AuthMiddleware 之后链式调用。
func AdminOnlyMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := GetUserID(c)
		if !ok {
			utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		var isAdmin int
		err := db.QueryRowContext(c.Request.Context(),
			`SELECT is_admin FROM users WHERE id = ?`, uid,
		).Scan(&isAdmin)
		if err != nil || isAdmin != 1 {
			utils.RespondLocalizedError(c, http.StatusForbidden, "admin.access_required")
			c.Abort()
			return
		}
		c.Next()
	}
}
