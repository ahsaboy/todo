package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todo/internal/i18n"
	"todo/internal/logging"
)

const (
	CodeNotFound      = "NOT_FOUND"
	CodeUnauthorized  = "UNAUTHORIZED"
	CodeForbidden     = "FORBIDDEN"
	CodeInvalidInput  = "INVALID_INPUT"
	CodeInternalError = "INTERNAL_ERROR"
	CodeRateLimited   = "RATE_LIMITED"
)

// SuccessResponse 通用成功响应
type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
}

// ErrorResponse 通用错误响应
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"task not found"`
	Code    string `json:"code" example:"NOT_FOUND"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    any         `json:"data"`
	Meta    PageMeta    `json:"meta"`
}

// PageMeta 分页元信息
type PageMeta struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"20"`
	TotalItems int64 `json:"total_items" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}

func RespondSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func RespondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func RespondError(c *gin.Context, code int, message string, errCode string) {
	logging.SetResponseLogMeta(c, errCode, message)
	c.JSON(code, gin.H{"success": false, "error": message, "code": errCode})
}

func RespondInternalError(c *gin.Context, message string, err error) {
	if err != nil {
		logging.LoggerFromContext(c).Error(message, zap.Error(err))
	}
	RespondError(c, http.StatusInternalServerError, message, CodeInternalError)
}

func RespondPaginated(c *gin.Context, data any, page, limit int, total int64) {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

// RespondLocalizedError 发送本地化的错误响应
// errCode 是 i18n 注册表中的消息键，不是HTTP错误码
func RespondLocalizedError(c *gin.Context, httpStatus int, errCode string, params ...interface{}) {
	msg := i18n.TL(c, errCode, params...)
	logging.SetResponseLogMeta(c, errCode, msg)
	c.JSON(httpStatus, gin.H{"success": false, "error": msg, "code": errCodeToResponseCode(errCode)})
}

// RespondLocalizedInternalError 记录Go错误并发送本地化的内部错误
func RespondLocalizedInternalError(c *gin.Context, errKey string, err error, params ...interface{}) {
	if err != nil {
		logging.LoggerFromContext(c).Error(i18n.TL(c, errKey, params...), zap.Error(err))
	}
	msg := i18n.TL(c, errKey, params...)
	logging.SetResponseLogMeta(c, "INTERNAL_ERROR", msg)
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": msg, "code": CodeInternalError})
}

// errCodeToResponseCode 将i18n消息键映射到现有的6个响应码
func errCodeToResponseCode(key string) string {
	switch {
	case strings.HasSuffix(key, ".not_found") || key == "not_found" || strings.Contains(key, "not_found"):
		return CodeNotFound
	case strings.Contains(key, "unauthorized") || strings.HasSuffix(key, ".invalid_api_key") ||
		strings.HasSuffix(key, ".missing_api_key") || strings.HasSuffix(key, ".invalid_old_password") ||
		strings.HasSuffix(key, ".invalid_credentials") || strings.HasSuffix(key, ".username_taken"):
		return CodeUnauthorized
	case strings.Contains(key, "forbidden") || strings.HasPrefix(key, "admin."):
		return CodeForbidden
	case key == "rate_limited" || key == "rate_limited_admin_login":
		return CodeRateLimited
	default:
		return CodeInvalidInput
	}
}
