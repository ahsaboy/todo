package main

import (
	"strings"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

func isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/admin/api")
}

func writeAPINotFound(c *gin.Context) {
	utils.RespondLocalizedError(c, 404, "system.endpoint_not_found")
}
