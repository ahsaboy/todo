package main

import (
	"strings"

	"github.com/gin-gonic/gin"

	"todo/internal/utils"
)

func isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api")
}

func writeAPINotFound(c *gin.Context) {
	utils.RespondError(c, 404, "endpoint not found", utils.CodeNotFound)
}
