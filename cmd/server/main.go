package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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
	"todo/internal/i18n"
	"todo/internal/logging"
	mcpserver "todo/internal/mcp"
	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/oauth"
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

	// 收集命令行显式设置的 flag → 锁定对应配置项(必须在任何 CLI 赋值之前)。
	// 被锁定的项不会被数据库配置覆盖,保证命令行优先级最高。
	lockedKeys := collectLockedKeys()

	// 初始化数据库(提前到此:数据库配置需先连库才能读取并覆盖到 cfg)。
	// 数据库路径只来自默认/配置文件/环境变量,绝不入库,避免循环依赖。
	// 此时尚未建立 zap logger,初始化失败用 stderr 退出。
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 读取数据库配置并覆盖到 cfg(优先级: 数据库 > 环境变量 > 配置文件)。
	// 读取失败时降级为空,不阻断启动(配置回退到配置文件/环境变量)。
	appConfigRepo := repository.NewAppConfigRepo(db)
	appConfigSvc := service.NewAppConfigService(appConfigRepo)
	var dbConfigSkipped []string
	if dbValues, derr := appConfigRepo.LoadAll(context.Background()); derr != nil {
		fmt.Fprintf(os.Stderr, "读取数据库配置失败,已忽略: %v\n", derr)
	} else {
		dbConfigSkipped = config.ApplyDBOverrides(cfg, dbValues, lockedKeys)
	}

	// 命令行 flag 覆盖(最后执行,优先级最高)。
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

	// 初始化i18n默认语言
	if cfg.I18n.DefaultLang != "" {
		i18n.SetDefaultLang(cfg.I18n.DefaultLang)
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
	if len(dbConfigSkipped) > 0 {
		logger.Warn("部分数据库配置项无效已忽略", zap.Strings("keys", dbConfigSkipped))
	}

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

	// 初始化各层
	userRepo := repository.NewUserRepo(db)
	apiKeyRepo := repository.NewAPIKeyRepo(db)
	reminderConfigRepo := repository.NewReminderConfigRepo(db)
	reminderLogRepo := repository.NewReminderLogRepo(db)
	tagRepo := repository.NewTagRepo(db)
	taskRepo := repository.NewTaskRepo(db, time.Duration(cfg.Reminder.GracePeriodMinutes)*time.Minute)

	authSvc := service.NewAuthService(userRepo, apiKeyRepo)
	emailSvc := service.NewEmailService(db, cfg, logger)

	// OAuth 初始化
	oauthRepo := repository.NewOAuthRepo(db)
	oauthReg := oauth.NewRegistry(cfg)
	var oauthHandler *handlers.OAuthHandler
	var oauthSvcInstance *service.OAuthService
	if cfg.OAuth.Enabled {
		if cfg.OAuth.StateSecret == "" {
			secret := make([]byte, 32)
			rand.Read(secret)
			cfg.OAuth.StateSecret = hex.EncodeToString(secret)
			logger.Warn("OAuth state_secret 未配置，已自动生成随机密钥（重启后失效）")
		}
		if err := oauthReg.Init(context.Background()); err != nil {
			logger.Error("OAuth provider 初始化失败", zap.Error(err))
		}
		oauthSvcInstance = service.NewOAuthService(userRepo, oauthRepo, apiKeyRepo, oauthReg, authSvc)
		oauthHandler = handlers.NewOAuthHandler(oauthSvcInstance, oauthReg, cfg)
	}

	// ProfileHandler 聚合用户资料（含 OAuth 绑定信息）
	var profileHandler *handlers.ProfileHandler
	if oauthSvcInstance != nil {
		profileHandler = handlers.NewProfileHandler(authSvc, oauthSvcInstance)
	} else {
		profileHandler = handlers.NewProfileHandler(authSvc, nil)
	}
	reminderConfigSvc := service.NewReminderConfigService(reminderConfigRepo)
	tagSvc := service.NewTagService(tagRepo)
	taskSvc := service.NewTaskService(taskRepo, reminderConfigRepo, tagSvc)

	authHandler := handlers.NewAuthHandler(authSvc, emailSvc)
	reminderConfigHandler := handlers.NewReminderConfigHandler(reminderConfigSvc)
	reminderLogSvc := service.NewReminderLogService(reminderLogRepo)
	reminderLogHandler := handlers.NewReminderLogHandler(reminderLogSvc)
	taskHandler := handlers.NewTaskHandler(taskSvc)
	tagHandler := handlers.NewTagHandler(tagSvc)

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

	// 初始管理员账号:首次启动时自动创建
	seedAdminUser(ctx, cfg, authSvc, userRepo, logger)

	go reminderSvc.Start(ctx)

	// 设置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(logging.AccessLogger(logger))
	r.Use(logging.Recovery(logger))

	// CORS（始终注册,内部根据 cfg.CORS.Enabled 动态决定是否生效）
	r.Use(corsMiddleware(cfg))

	// 公开端点
	r.GET("/api/v1/health", handlers.HealthCheck(db))
	r.GET("/api/v1/templates", reminderTemplatesHandler(cfg))

	// 公开路由（无需认证）
	auth := r.Group("/api/v1/auth")
	if cfg.RateLimit.Enabled {
		auth.Use(middleware.IPRateLimit(cfg.RateLimit.AuthReqsPerSecond, cfg.RateLimit.AuthBurst))
	}
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/email-status", authHandler.EmailStatus)
		auth.POST("/send-code", authHandler.SendCode)
		auth.POST("/verify-code", authHandler.VerifyCode)
		auth.POST("/reset-password", authHandler.ResetPassword)

		// OAuth 社交登录
		if oauthHandler != nil {
			auth.GET("/oauth/providers", oauthHandler.GetProviders)
			auth.GET("/oauth/:provider", oauthHandler.InitiateOAuth)
			auth.GET("/oauth/:provider/callback", oauthHandler.HandleCallback)
		}
	}

	// 需认证路由
	api := r.Group("/api/v1", middleware.AuthMiddleware(apiKeyRepo))
	if cfg.RateLimit.Enabled {
		api.Use(middleware.IPRateLimit(cfg.RateLimit.ReqsPerSecond, cfg.RateLimit.Burst))
	}
	{
		// 任务管理
		api.POST("/tasks", taskHandler.Create)
		api.GET("/tasks", taskHandler.List)
		api.GET("/tasks/:id", taskHandler.GetByID)
		api.PUT("/tasks/:id", taskHandler.Update)
		api.DELETE("/tasks/:id", taskHandler.Delete)
		api.PATCH("/tasks/:id/complete", taskHandler.ToggleComplete)

		// 用户管理
		api.GET("/user/profile", profileHandler.GetFullProfile)
		api.PUT("/user/profile", authHandler.UpdateProfile)
		api.PUT("/user/password", authHandler.ChangePassword)
		api.POST("/user/set-password", authHandler.SetPassword)

		// OAuth 账号管理
		if oauthHandler != nil {
			api.GET("/user/oauth-accounts", oauthHandler.GetProfileAccounts)
			api.DELETE("/user/oauth-accounts/:id", oauthHandler.UnlinkOAuthAccount)
			api.GET("/user/oauth/:provider/link", oauthHandler.InitiateLinkOAuth)
		}

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

		api.GET("/tags", tagHandler.List)
		api.POST("/tags", tagHandler.Create)
		api.GET("/tags/:id", tagHandler.GetByID)
		api.PUT("/tags/:id", tagHandler.Update)
		api.DELETE("/tags/:id", tagHandler.Delete)
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

	registerOptionalRoutes(r, logger, cfg.StaticFiles, cfg.Admin.Enabled)
	registerAdminRoutes(r, db, cfg, userRepo, apiKeyRepo, taskRepo, reminderConfigRepo, reminderLogRepo, appConfigSvc, reminderSvc, emailSvc, lockedKeys, logger, oauthHandler, oauthReg)

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
			origins := config.SplitCommaSeparated(value)
			if len(origins) == 0 {
				return fmt.Errorf("CORS 必须是布尔值或逗号分隔的来源列表, 当前值: %q", value)
			}
			cfg.CORS.Enabled = true
			cfg.CORS.AllowedOrigins = origins
		}
	}

	// OAuth 环境变量覆盖
	if value, ok := lookupEnvAny("OAUTH_ENABLED", "oauth_enabled"); ok {
		if enabled, err := strconv.ParseBool(value); err == nil {
			cfg.OAuth.Enabled = enabled
		}
	}
	if value, ok := lookupEnvAny("OAUTH_FRONTEND_URL", "oauth_frontend_url"); ok {
		cfg.OAuth.FrontendURL = value
	}
	if value, ok := lookupEnvAny("OAUTH_ADMIN_URL", "oauth_admin_url"); ok {
		cfg.OAuth.AdminURL = value
	}
	if value, ok := lookupEnvAny("OAUTH_STATE_SECRET", "oauth_state_secret"); ok {
		cfg.OAuth.StateSecret = value
	}
	if value, ok := lookupEnvAny("OAUTH_GITHUB_CLIENT_ID", "oauth_github_client_id"); ok {
		cfg.OAuth.GitHub.ClientID = value
	}
	if value, ok := lookupEnvAny("OAUTH_GITHUB_CLIENT_SECRET", "oauth_github_client_secret"); ok {
		cfg.OAuth.GitHub.ClientSecret = value
		cfg.OAuth.GitHub.Enabled = true
	}
	if value, ok := lookupEnvAny("OAUTH_GOOGLE_CLIENT_ID", "oauth_google_client_id"); ok {
		cfg.OAuth.Google.ClientID = value
	}
	if value, ok := lookupEnvAny("OAUTH_GOOGLE_CLIENT_SECRET", "oauth_google_client_secret"); ok {
		cfg.OAuth.Google.ClientSecret = value
		cfg.OAuth.Google.Enabled = true
	}
	if value, ok := lookupEnvAny("OAUTH_LINUXDO_CLIENT_ID", "oauth_linuxdo_client_id"); ok {
		cfg.OAuth.LinuxDo.ClientID = value
	}
	if value, ok := lookupEnvAny("OAUTH_LINUXDO_CLIENT_SECRET", "oauth_linuxdo_client_secret"); ok {
		cfg.OAuth.LinuxDo.ClientSecret = value
		cfg.OAuth.LinuxDo.Enabled = true
	}

	return nil
}

// collectLockedKeys 返回被命令行显式设置过的 flag 对应的配置项 key。
// 这些项不会被数据库配置覆盖,保证命令行优先级最高。
func collectLockedKeys() map[string]bool {
	locked := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port", "p":
			locked["server.port"] = true
		case "host":
			locked["server.host"] = true
		case "mode":
			locked["server.mode"] = true
		case "log-path":
			locked["logging.path"] = true
		case "log-max-days":
			locked["logging.max_days"] = true
		case "log-file-enabled", "log-file-disabled":
			locked["logging.file_enabled"] = true
		case "static-files-enabled", "static-files-disabled":
			locked["static_files"] = true
		}
	})
	return locked
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

func registerOptionalRoutes(r *gin.Engine, logger *zap.Logger, staticFilesEnabled, adminEnabled bool) {
	if staticFilesEnabled {
		r.GET("/docs/*any", func(c *gin.Context) {
			httpSwagger.WrapHandler.ServeHTTP(c.Writer, c.Request)
		})
	}

	// 前端 SPA(含管理后台界面)在 静态资源开启 或 管理后台启用 时都注册,
	// 保证 admin.enabled 时管理后台界面始终可登录,不被 static_files 锁死。
	if staticFilesEnabled || adminEnabled {
		registerStaticRoutes(r, logger)
		return
	}

	logger.Info("静态资源与管理后台均关闭，跳过前端静态资源与 Swagger 路由注册")
	r.NoRoute(func(c *gin.Context) {
		writeAPINotFound(c)
	})
}

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.CORS.Enabled {
			c.Next()
			return
		}
		if origin, ok := allowedCORSOrigin(c.GetHeader("Origin"), cfg.CORS.AllowedOrigins); ok {
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

func registerAdminRoutes(
	r *gin.Engine,
	db *sql.DB,
	cfg *config.Config,
	userRepo repository.UserRepository,
	apiKeyRepo repository.APIKeyRepository,
	taskRepo repository.TaskRepository,
	reminderConfigRepo repository.ReminderConfigRepository,
	reminderLogRepo repository.ReminderLogRepository,
	appConfigSvc *service.AppConfigService,
	reminderSvc *service.ReminderService,
	emailSvc service.EmailServiceInterface,
	lockedKeys map[string]bool,
	logger *zap.Logger,
	oauthHandler *handlers.OAuthHandler,
	oauthReg *oauth.Registry,
) {
	if !cfg.Admin.Enabled {
		logger.Warn("管理后台已禁用，请在 config.yaml 中配置 admin.enabled=true")
		return
	}
	auditLogRepo := repository.NewAuditLogRepo(db)
	adminH := handlers.NewAdminHandler(db, userRepo, apiKeyRepo, taskRepo, reminderConfigRepo, reminderLogRepo, auditLogRepo, cfg, appConfigSvc, reminderSvc, emailSvc, lockedKeys, func(lang string) {
		i18n.SetDefaultLang(lang)
	}, func() {
		if oauthReg != nil {
			oauthReg.Reload(context.Background(), cfg)
		}
	})
	adm := r.Group("/admin/api")
	adm.Use(middleware.LocalhostOnly())
	adm.POST("/auth/login", middleware.AdminLoginRateLimit(10, time.Minute), adminH.AdminLogin)

	// 管理后台 OAuth 社交登录
	if oauthHandler != nil {
		adm.GET("/auth/oauth/providers", oauthHandler.GetProviders)
		adm.GET("/auth/oauth/:provider", oauthHandler.AdminInitiateOAuth)
		adm.GET("/auth/oauth/:provider/callback", oauthHandler.AdminHandleCallback)
	}

	authed := adm.Group("", middleware.AuthMiddleware(apiKeyRepo), middleware.AdminOnlyMiddleware(db))
	{
		authed.GET("/stats", adminH.GetStats)
		authed.GET("/stats/trends", adminH.GetTrends)
		authed.GET("/users", adminH.ListUsers)
		authed.GET("/users/:id", adminH.GetUser)
		authed.DELETE("/users/:id", adminH.DeleteUser)
		authed.POST("/users/:id/reset-password", adminH.ForceResetPassword)
		authed.PATCH("/users/:id/admin", adminH.ToggleAdmin)
		authed.GET("/tasks", adminH.ListAllTasks)
		authed.PATCH("/tasks/:id/toggle", adminH.AdminToggleComplete)
		authed.PUT("/tasks/:id", adminH.AdminUpdateTask)
		authed.DELETE("/tasks/:id", adminH.AdminDeleteTask)
		authed.GET("/reminder-configs", adminH.ListAllReminderConfigs)
		authed.PATCH("/reminder-configs/:id/toggle", adminH.AdminToggleReminderConfig)
		authed.DELETE("/reminder-configs/:id", adminH.AdminDeleteReminderConfig)
		authed.GET("/reminder-logs", adminH.ListAllReminderLogs)
		authed.GET("/config", adminH.GetConfig)
		authed.PUT("/config", adminH.UpdateConfig)
		authed.DELETE("/config/*key", adminH.ResetConfig)
		authed.POST("/config/test-email", adminH.TestEmail)
		authed.GET("/oauth/callback-urls", adminH.GetOAuthCallbackURLs)
		authed.GET("/audit-logs", adminH.ListAuditLogs)

		systemLogH := handlers.NewSystemLogHandler(cfg)
		authed.GET("/system-logs", systemLogH.ListLogFiles)
		authed.GET("/system-logs/:filename/entries", systemLogH.GetLogEntries)
		authed.GET("/system-logs/:filename/download", systemLogH.DownloadLogFile)
	}
	logger.Info("管理后台已启动", zap.String("prefix", "/admin/api"))
}

// seedAdminUser 根据配置文件的 admin.username/password 自动创建初始管理员。
// 幂等：用户名已存在则跳过；username 或 password 为空则不执行。
func seedAdminUser(ctx context.Context, cfg *config.Config, authSvc *service.AuthService, userRepo repository.UserRepository, logger *zap.Logger) {
	adminUser := cfg.Admin.Username
	adminPass := cfg.Admin.Password
	if adminUser == "" || adminPass == "" {
		return
	}

	_, _, err := authSvc.Register(ctx, models.RegisterRequest{
		Username: adminUser,
		Password: adminPass,
		Email:    cfg.Admin.Email,
	})
	if err == service.ErrUsernameTaken {
		logger.Info("管理员账号已存在,跳过自动创建", zap.String("username", adminUser))
	} else if err != nil {
		logger.Fatal("自动创建管理员失败", zap.String("username", adminUser), zap.Error(err))
	} else {
		logger.Info("初始管理员账号已创建", zap.String("username", adminUser))
	}

	// 确保管理员用户有 is_admin 标记
	user, err := userRepo.GetByUsername(ctx, adminUser)
	if err != nil || user == nil {
		logger.Error("查找管理员用户失败", zap.String("username", adminUser), zap.Error(err))
		return
	}
	if !user.IsAdmin {
		if err := userRepo.SetIsAdmin(ctx, user.ID, true); err != nil {
			logger.Error("设置管理员标记失败", zap.String("username", adminUser), zap.Error(err))
		} else {
			logger.Info("已设置 is_admin 标记", zap.String("username", adminUser))
		}
	}
}
