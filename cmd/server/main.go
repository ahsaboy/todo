package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	docs "todo/docs"
	"todo/internal/config"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/logging"
	mcpserver "todo/internal/mcp"
	"todo/internal/middleware"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
)

// @title           TODO 任务管理系统 API
// @version         2.0.0
// @description     多用户 TODO 任务管理系统，支持用户注册/登录、个人 API Key、多渠道提醒推送
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name api-key

var version = "dev"

func main() {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	port := flag.Int("port", 0, "覆盖服务端口号")
	host := flag.String("host", "", "覆盖监听地址")
	mode := flag.String("mode", "", "覆盖运行模式 (debug/release)")
	logPath := flag.String("log-path", "", "覆盖日志存储路径")
	logMaxDays := flag.Int("log-max-days", 0, "覆盖日志保留天数")
	logFileEnabled := flag.Bool("log-file-enabled", false, "启用日志文件输出")
	logFileDisabled := flag.Bool("log-file-disabled", false, "禁用日志文件输出")
	staticFilesEnabled := flag.Bool("static-files-enabled", false, "启用前端静态文件与 Swagger 路由")
	staticFilesDisabled := flag.Bool("static-files-disabled", false, "禁用前端静态文件与 Swagger 路由")
	showVersion := flag.Bool("version", false, "显示版本号")
	showHelp := flag.Bool("help", false, "显示帮助信息")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "TODO 任务管理系统 %s\n\n", version)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  todo-server [选项]\n\n")
		fmt.Fprintf(os.Stderr, "选项:\n")
		fmt.Fprintf(os.Stderr, "  -c, --config <path>  配置文件路径 (默认: config.yaml)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port <port>    覆盖服务端口号\n")
		fmt.Fprintf(os.Stderr, "  --host <addr>        覆盖监听地址\n")
		fmt.Fprintf(os.Stderr, "  --mode <mode>        覆盖运行模式 (debug/release)\n")
		fmt.Fprintf(os.Stderr, "  --log-path <path>    覆盖日志存储路径\n")
		fmt.Fprintf(os.Stderr, "  --log-max-days <n>   覆盖日志保留天数\n")
		fmt.Fprintf(os.Stderr, "  --log-file-enabled   启用日志文件输出\n")
		fmt.Fprintf(os.Stderr, "  --log-file-disabled  禁用日志文件输出\n")
		fmt.Fprintf(os.Stderr, "  --static-files-enabled   启用前端静态文件与 Swagger 路由\n")
		fmt.Fprintf(os.Stderr, "  --static-files-disabled  禁用前端静态文件与 Swagger 路由\n")
		fmt.Fprintf(os.Stderr, "  -v, --version        显示版本号\n")
		fmt.Fprintf(os.Stderr, "  -h, --help           显示此帮助信息\n")
	}

	flag.StringVar(cfgPath, "c", "config.yaml", "配置文件路径 (short)")
	flag.IntVar(port, "p", 0, "覆盖服务端口号 (short)")
	flag.BoolVar(showVersion, "v", false, "显示版本号 (short)")
	flag.BoolVar(showHelp, "h", false, "显示帮助信息 (short)")

	flag.Parse()

	if *showVersion {
		fmt.Printf("todo-server %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	if err := applyEnvOverrides(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *mode == "debug" || *mode == "release" {
		cfg.Server.Mode = *mode
	}
	if err := applyLoggingOverrides(cfg, *logPath, *logMaxDays, *logFileEnabled, *logFileDisabled); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := applyStaticFilesOverrides(cfg, *staticFilesEnabled, *staticFilesDisabled); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if cfg.Logging.FileEnabled {
		if err := logging.CleanupOldLogs(cfg.Logging.Path, cfg.Logging.MaxDays, time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "清理旧日志失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 初始化日志
	logger, logSyncer, err := logging.NewManagedLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logSyncer.Sync()

	logger.Info("配置加载完成",
		zap.String("config", *cfgPath),
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

	// 初始化进程级时区(供 view 函数和提醒模板使用)。
	// 失败时回退到 time.Local 并 warn,不阻断启动。
	loc, tzErr := utils.ResolveTimezone(cfg.Server.Timezone)
	timezone.Init(loc)
	if tzErr != nil {
		logger.Warn("解析 server.timezone 失败,已回退到 Local",
			zap.String("config", cfg.Server.Timezone),
			zap.Error(tzErr))
	} else {
		logger.Info("时区配置完成", zap.String("location", loc.String()))
	}

	docs.SwaggerInfo.Version = version

	// 初始化数据库
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	defer db.Close()

	// 初始化各层
	userRepo := repository.NewUserRepo(db)
	apiKeyRepo := repository.NewAPIKeyRepo(db)
	reminderConfigRepo := repository.NewReminderConfigRepo(db)
	reminderLogRepo := repository.NewReminderLogRepo(db)
	taskRepo := repository.NewTaskRepo(db)

	authSvc := service.NewAuthService(userRepo, apiKeyRepo)
	reminderConfigSvc := service.NewReminderConfigService(reminderConfigRepo)
	taskSvc := service.NewTaskService(taskRepo, reminderConfigRepo)

	authHandler := handlers.NewAuthHandler(authSvc)
	reminderConfigHandler := handlers.NewReminderConfigHandler(reminderConfigSvc)
	reminderLogHandler := handlers.NewReminderLogHandler(reminderLogRepo)
	taskHandler := handlers.NewTaskHandler(taskSvc)

	// 初始化提醒服务
	reminderSvc, err := service.NewReminderService(
		taskRepo, reminderConfigRepo, reminderLogRepo,
		cfg.Reminder,
		logger,
	)
	if err != nil {
		logger.Fatal("提醒服务初始化失败", zap.Error(err))
	}

	// 启动后台提醒
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go reminderSvc.Start(ctx)

	// 设置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(logging.AccessLogger(logger))
	r.Use(logging.Recovery(logger))

	// CORS
	if cfg.CORS.Enabled {
		r.Use(corsMiddleware(cfg.CORS.AllowedOrigins))
	}

	// 公开端点
	r.GET("/api/v1/health", handlers.HealthCheck(db))
	r.GET("/api/v1/templates", reminderTemplatesHandler(cfg))

	// 公开路由（无需认证）
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// 需认证路由
	api := r.Group("/api/v1", middleware.AuthMiddleware(apiKeyRepo))
	{
		// 任务管理
		api.POST("/tasks", taskHandler.Create)
		api.GET("/tasks", taskHandler.List)
		api.GET("/tasks/:id", taskHandler.GetByID)
		api.PUT("/tasks/:id", taskHandler.Update)
		api.DELETE("/tasks/:id", taskHandler.Delete)
		api.PATCH("/tasks/:id/complete", taskHandler.ToggleComplete)

		// 用户管理
		api.GET("/user/profile", authHandler.GetProfile)
		api.PUT("/user/profile", authHandler.UpdateProfile)
		api.PUT("/user/password", authHandler.ChangePassword)

		// API Key 管理
		api.GET("/user/keys", authHandler.ListAPIKeys)
		api.POST("/user/keys", authHandler.GenerateAPIKey)
		api.DELETE("/user/keys/:id", authHandler.RevokeAPIKey)

		// 提醒配置管理
		api.GET("/user/reminder-configs", reminderConfigHandler.List)
		api.POST("/user/reminder-configs", reminderConfigHandler.Create)
		api.GET("/user/reminder-configs/:id", reminderConfigHandler.GetByID)
		api.PUT("/user/reminder-configs/:id", reminderConfigHandler.Update)
		api.DELETE("/user/reminder-configs/:id", reminderConfigHandler.Delete)
		api.GET("/user/reminder-logs", reminderLogHandler.List)
	}

	// MCP(Model Context Protocol)端点。
	// 单一路径 /mcp 同时承载 POST(JSON-RPC 请求)、GET(SSE 事件流)、DELETE(关闭 session),
	// 故用 r.Any 挂载,而不放进 /api/v1 的 AuthMiddleware 分组 —— handler 内部已自带
	// api-key 认证,见 internal/mcp/auth.go。
	mcpHandler := mcpserver.NewMCPServer(mcpserver.Dependencies{
		TaskSvc:     taskSvc,
		ReminderSvc: reminderConfigSvc,
		AuthSvc:     authSvc,
		APIKeyRepo:  apiKeyRepo,
	})
	r.Any("/mcp", gin.WrapH(mcpHandler))

	registerOptionalRoutes(r, logger, cfg.StaticFiles)

	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		logger.Info("服务启动", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务异常退出", zap.Error(err))
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("关闭服务出错", zap.Error(err))
	}
	logger.Info("服务已停止")
}

// reminderTemplatesHandler 查看预置提醒模板
// @Summary      查看预置提醒模板
// @Description  返回 config.yaml 中 reminder.default_templates 配置的预置模板；模板只作为创建用户通知渠道时的参考，不会直接用于发送
// @Tags         templates
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Router       /api/v1/templates [get]
func reminderTemplatesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		templates := cfg.Reminder.DefaultTemplates
		if templates == nil {
			templates = map[string]config.DefaultTemplate{}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": templates})
	}
}

func applyLoggingOverrides(cfg *config.Config, logPath string, logMaxDays int, logFileEnabled bool, logFileDisabled bool) error {
	if logPath != "" {
		cfg.Logging.Path = logPath
	}
	if logMaxDays > 0 {
		cfg.Logging.MaxDays = logMaxDays
	}
	if logFileEnabled && logFileDisabled {
		return fmt.Errorf("--log-file-enabled 与 --log-file-disabled 不能同时使用")
	}
	if logFileEnabled {
		cfg.Logging.FileEnabled = true
	}
	if logFileDisabled {
		cfg.Logging.FileEnabled = false
	}
	return nil
}

func applyEnvOverrides(cfg *config.Config) error {
	if value, ok := lookupEnvAny("HOST", "host"); ok {
		cfg.Server.Host = value
	}

	if value, ok := lookupEnvAny("PORT", "port"); ok {
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 {
			return fmt.Errorf("PORT 必须是有效的正整数, 当前值: %q", value)
		}
		cfg.Server.Port = port
	}

	if value, ok := lookupEnvAny("STATIC_FILES", "static_files"); ok {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("STATIC_FILES 必须是布尔值, 当前值: %q", value)
		}
		cfg.StaticFiles = enabled
	}

	if value, ok := lookupEnvAny("CORS", "cors"); ok {
		if enabled, err := strconv.ParseBool(value); err == nil {
			cfg.CORS.Enabled = enabled
			if !enabled {
				cfg.CORS.AllowedOrigins = nil
			}
		} else {
			origins := splitCommaSeparated(value)
			if len(origins) == 0 {
				return fmt.Errorf("CORS 必须是布尔值或逗号分隔的来源列表, 当前值: %q", value)
			}
			cfg.CORS.Enabled = true
			cfg.CORS.AllowedOrigins = origins
		}
	}

	return nil
}

func applyStaticFilesOverrides(cfg *config.Config, staticFilesEnabled bool, staticFilesDisabled bool) error {
	if staticFilesEnabled && staticFilesDisabled {
		return fmt.Errorf("--static-files-enabled 与 --static-files-disabled 不能同时使用")
	}
	if staticFilesEnabled {
		cfg.StaticFiles = true
	}
	if staticFilesDisabled {
		cfg.StaticFiles = false
	}
	return nil
}

func lookupEnvAny(keys ...string) (string, bool) {
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		return value, true
	}
	return "", false
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func registerOptionalRoutes(r *gin.Engine, logger *zap.Logger, staticFilesEnabled bool) {
	if staticFilesEnabled {
		r.GET("/docs/*any", func(c *gin.Context) {
			httpSwagger.WrapHandler.ServeHTTP(c.Writer, c.Request)
		})
		registerStaticRoutes(r, logger)
		return
	}

	logger.Info("静态资源开关已关闭，跳过前端静态资源与 Swagger 路由注册")
	r.NoRoute(func(c *gin.Context) {
		writeAPINotFound(c)
	})
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin, ok := allowedCORSOrigin(c.GetHeader("Origin"), origins); ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		// Mcp-Session-Id / Last-Event-ID 由 MCP Streamable HTTP transport 使用(2025-11-25 规范)。
		// X-MCP-Structured-Output / X-MCP-Include-Reminders / X-MCP-Timezone 是本服务自定义的 MCP 客户端选项 header。
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, api-key, X-API-Key, Mcp-Session-Id, Last-Event-ID, X-MCP-Structured-Output, X-MCP-Include-Reminders, X-MCP-Timezone")
		// 让浏览器客户端能读到 initialize 响应里的 Mcp-Session-Id。
		c.Header("Access-Control-Expose-Headers", "Mcp-Session-Id")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func allowedCORSOrigin(requestOrigin string, allowedOrigins []string) (string, bool) {
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			if requestOrigin != "" {
				return requestOrigin, true
			}
			return "*", true
		}
		if requestOrigin != "" && origin == requestOrigin {
			return requestOrigin, true
		}
	}
	return "", false
}
