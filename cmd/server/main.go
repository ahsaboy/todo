package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "todo/docs"
	"todo/internal/config"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/logging"
	"todo/internal/middleware"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/web"
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

const version = "2.0.0"

func main() {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	port := flag.Int("port", 0, "覆盖服务端口号")
	host := flag.String("host", "", "覆盖监听地址")
	mode := flag.String("mode", "", "覆盖运行模式 (debug/release)")
	logPath := flag.String("log-path", "", "覆盖日志存储路径")
	logMaxDays := flag.Int("log-max-days", 0, "覆盖日志保留天数")
	backendLog := flag.String("backend-log", "", "覆盖后端日志输出模式 (console/file/both/off)")
	frontendLog := flag.String("frontend-log", "", "覆盖前端日志输出模式 (console/file/both/off)")
	frontendLogLevel := flag.String("frontend-log-level", "", "覆盖前端日志级别")
	showVersion := flag.Bool("version", false, "显示版本号")
	showHelp := flag.Bool("help", false, "显示帮助信息")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "TODO 任务管理系统 v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  todo-server [选项]\n\n")
		fmt.Fprintf(os.Stderr, "选项:\n")
		fmt.Fprintf(os.Stderr, "  -c, --config <path>  配置文件路径 (默认: config.yaml)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port <port>    覆盖服务端口号\n")
		fmt.Fprintf(os.Stderr, "  --host <addr>        覆盖监听地址\n")
		fmt.Fprintf(os.Stderr, "  --mode <mode>        覆盖运行模式 (debug/release)\n")
		fmt.Fprintf(os.Stderr, "  --log-path <path>    覆盖日志存储路径\n")
		fmt.Fprintf(os.Stderr, "  --log-max-days <n>   覆盖日志保留天数\n")
		fmt.Fprintf(os.Stderr, "  --backend-log <mode> 覆盖后端日志输出模式 (console/file/both/off)\n")
		fmt.Fprintf(os.Stderr, "  --frontend-log <mode> 覆盖前端日志输出模式 (console/file/both/off)\n")
		fmt.Fprintf(os.Stderr, "  --frontend-log-level <level>  覆盖前端日志级别\n")
		fmt.Fprintf(os.Stderr, "  -v, --version        显示版本号\n")
		fmt.Fprintf(os.Stderr, "  -h, --help           显示此帮助信息\n")
	}

	flag.StringVar(cfgPath, "c", "config.yaml", "配置文件路径 (short)")
	flag.IntVar(port, "p", 0, "覆盖服务端口号 (short)")
	flag.BoolVar(showVersion, "v", false, "显示版本号 (short)")
	flag.BoolVar(showHelp, "h", false, "显示帮助信息 (short)")

	flag.Parse()

	if *showVersion {
		fmt.Printf("todo-server v%s\n", version)
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

	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *mode == "debug" || *mode == "release" {
		cfg.Server.Mode = *mode
	}
	if err := applyLoggingOverrides(cfg, *logPath, *logMaxDays, *backendLog, *frontendLog, *frontendLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := logging.CleanupOldLogs(cfg.Logging.Path, cfg.Logging.MaxDays, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "清理旧日志失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("配置加载完成",
		zap.String("config", *cfgPath),
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

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
	logHandler := handlers.NewLogHandler(cfg.Logging, logger)
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
	r.Use(gin.Recovery())

	// CORS
	if cfg.CORS.Enabled {
		r.Use(corsMiddleware(cfg.CORS.AllowedOrigins))
	}

	// 公开端点
	r.GET("/api/v1/health", handlers.HealthCheck(db))
	r.GET("/api/v1/runtime-config", logHandler.RuntimeConfig)
	r.POST("/api/v1/logs/frontend", logHandler.FrontendLogs)
	r.GET("/api/v1/templates", reminderTemplatesHandler(cfg))
	r.GET("/docs/*any", func(c *gin.Context) {
		httpSwagger.WrapHandler.ServeHTTP(c.Writer, c.Request)
	})

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

	// 静态资源服务
	distFS, err := fs.Sub(web.Files, "dist")
	if err != nil {
		logger.Fatal("无法加载静态资源", zap.Error(err))
	}
	assetsFS, err := fs.Sub(distFS, "assets")
	if err != nil {
		logger.Fatal("无法加载 assets 资源", zap.Error(err))
	}
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		logger.Fatal("无法加载 index.html", zap.Error(err))
	}

	// 静态文件（优先级高于 NoRoute）
	r.StaticFS("/assets", http.FS(assetsFS))
	r.StaticFileFS("/favicon.svg", "favicon.svg", http.FS(distFS))
	r.StaticFileFS("/icons.svg", "icons.svg", http.FS(distFS))

	// SPA fallback
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 路由返回 JSON
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "endpoint not found",
				"code":    "NOT_FOUND",
			})
			return
		}

		// 其他路由返回 index.html。不要用 FileFromFS("index.html")，它底层的
		// http.FileServer 会把 /index.html 规范化重定向到 ./，导致根路径循环跳转。
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

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

func applyLoggingOverrides(cfg *config.Config, logPath string, logMaxDays int, backendLog string, frontendLog string, frontendLogLevel string) error {
	if logPath != "" {
		cfg.Logging.Path = logPath
	}
	if logMaxDays > 0 {
		cfg.Logging.MaxDays = logMaxDays
	}
	if backendLog != "" {
		consoleEnabled, fileEnabled, err := applyLogOutputMode(backendLog)
		if err != nil {
			return fmt.Errorf("无效的 --backend-log 值: %w", err)
		}
		cfg.Logging.Backend.ConsoleEnabled = consoleEnabled
		cfg.Logging.Backend.FileEnabled = fileEnabled
	}
	if frontendLog != "" {
		consoleEnabled, fileEnabled, err := applyLogOutputMode(frontendLog)
		if err != nil {
			return fmt.Errorf("无效的 --frontend-log 值: %w", err)
		}
		cfg.Logging.Frontend.ConsoleEnabled = consoleEnabled
		cfg.Logging.Frontend.FileEnabled = fileEnabled
	}
	if frontendLogLevel != "" {
		cfg.Logging.Frontend.Level = frontendLogLevel
	}
	return nil
}

func applyLogOutputMode(mode string) (bool, bool, error) {
	switch mode {
	case "console":
		return true, false, nil
	case "file":
		return false, true, nil
	case "both":
		return true, true, nil
	case "off":
		return false, false, nil
	default:
		return false, false, fmt.Errorf("可选值: console, file, both, off")
	}
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := "*"
		if len(origins) > 0 {
			origin = origins[0]
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, api-key, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
