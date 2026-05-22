package middleware

import (
	"net"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

// LocalhostOnly rejects requests whose TCP RemoteAddr is not 127.0.0.1 / ::1.
// This relies on a direct connection and intentionally ignores X-Forwarded-For.
// If the server runs behind a reverse proxy, the proxy must be on the same host
// and the admin port must not be publicly exposed.
func LocalhostOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil || (host != "127.0.0.1" && host != "::1") {
			utils.RespondError(c, 403, "admin interface is localhost-only", utils.CodeUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}
