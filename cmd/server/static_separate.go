//go:build separate_frontend

package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func registerStaticRoutes(r *gin.Engine, logger *zap.Logger) {
	logger.Info("separate_frontend 构建模式已启用，跳过前端静态资源注册")

	r.NoRoute(func(c *gin.Context) {
		writeAPINotFound(c)
	})
}
