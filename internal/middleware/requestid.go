package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"todo/internal/i18n"
)

const RequestIDHeader = "X-Request-ID"
const RequestIDKey = "request_id"

// RequestID 读取或生成 X-Request-ID，写回响应头并存入 gin context。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(RequestIDKey, id)
		c.Header(RequestIDHeader, id)

		// 设置语言（从Accept-Language header解析）
		i18n.SetLang(c, i18n.AcceptLanguage(c.GetHeader("Accept-Language")))

		c.Next()
	}
}
