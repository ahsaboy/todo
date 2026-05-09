package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeNotFound      = "NOT_FOUND"
	CodeUnauthorized  = "UNAUTHORIZED"
	CodeInvalidInput  = "INVALID_INPUT"
	CodeInternalError = "INTERNAL_ERROR"
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
	c.JSON(code, gin.H{"success": false, "error": message, "code": errCode})
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
