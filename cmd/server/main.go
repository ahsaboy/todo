package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "todo/docs"
	"todo/internal/config"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/middleware"
	"todo/internal/repository"
	"todo/internal/service"
)

// @title           TODO 任务管理系统 API
// @version         2.0.0
// @description     多用户 TODO 任务管理系统，支持用户注册/登录、个人 API Key、多渠道提醒推送
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name api-key

const version = "2.0.0"

func main() {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	port := flag.Int("port", 0, "覆盖服务端口号")
	host := flag.String("host", "", "覆盖监听地址")
	mode := flag.String("mode", "", "覆盖运行模式 (debug/release)")
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

	// 初始化日志
	logger, err := initLogger(cfg.Logging)
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
	taskRepo := repository.NewTaskRepo(db)

	authSvc := service.NewAuthService(userRepo, apiKeyRepo)
	reminderConfigSvc := service.NewReminderConfigService(reminderConfigRepo)
	taskSvc := service.NewTaskService(taskRepo, reminderConfigRepo)

	authHandler := handlers.NewAuthHandler(authSvc)
	reminderConfigHandler := handlers.NewReminderConfigHandler(reminderConfigSvc)
	taskHandler := handlers.NewTaskHandler(taskSvc)

	// 初始化提醒服务
	reminderSvc, err := service.NewReminderService(
		taskRepo, reminderConfigRepo,
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
	r.GET("/api/v1/templates", func(c *gin.Context) {
		templates := cfg.Reminder.DefaultTemplates
		if templates == nil {
			templates = map[string]config.DefaultTemplate{}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": templates})
	})
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
	}

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

func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Encoding:    cfg.Format,
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
	}

	if cfg.Format != "json" {
		zapCfg.Encoding = "console"
	}

	return zapCfg.Build()
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
