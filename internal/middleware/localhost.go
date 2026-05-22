package middleware

import (
	"net"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

// LocalhostOnly rejects requests whose TCP RemoteAddr is not 127.0.0.1 / ::1
// or a private/loopback network (localhost, 10.x.x.x, 172.16-31.x.x, 192.168.x.x).
func LocalhostOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			utils.RespondError(c, 403, "admin interface is localhost-only", utils.CodeUnauthorized)
			c.Abort()
			return
		}

		ip := net.ParseIP(host)
		if ip == nil {
			utils.RespondError(c, 403, "admin interface is localhost-only", utils.CodeUnauthorized)
			c.Abort()
			return
		}

		// Allow loopback addresses
		if ip.IsLoopback() {
			c.Next()
			return
		}

		// Allow private networks (10.x.x.x, 172.16-31.x.x, 192.168.x.x)
		if ip.IsPrivate() {
			c.Next()
			return
		}

		utils.RespondError(c, 403, "admin interface is localhost-only", utils.CodeUnauthorized)
		c.Abort()
	}
}
