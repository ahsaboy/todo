package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api")
}

func writeAPINotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "endpoint not found",
		"code":    "NOT_FOUND",
	})
}
