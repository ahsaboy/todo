package logging

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"todo/internal/config"
)

const (
	defaultLogPath          = "./logs"
	logCodeContextKey       = "log_code"
	logMessageContextKey    = "log_message"
	requestLoggerContextKey = "logger"
)

var staticAssetExtensions = map[string]struct{}{
	".avif":        {},
	".css":         {},
	".eot":         {},
	".gif":         {},
	".ico":         {},
	".jpeg":        {},
	".jpg":         {},
	".js":          {},
	".map":         {},
	".mjs":         {},
	".png":         {},
	".svg":         {},
	".ttf":         {},
	".txt":         {},
	".webmanifest": {},
	".webp":        {},
	".woff":        {},
	".woff2":       {},
	".xml":         {},
}

type dateNowFunc func() time.Time

type SyncLogger interface {
	Sync() error
}

type dailyFileWriter struct {
	dir     string
	now     dateNowFunc
	mu      sync.Mutex
	current string
	file    *os.File
}

func newDailyFileWriter(dir string, now dateNowFunc) *dailyFileWriter {
	if now == nil {
		now = time.Now
	}
	return &dailyFileWriter{dir: dir, now: now}
}

func (w *dailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := w.currentFileLocked()
	if err != nil {
		return 0, err
	}
	return file.Write(p)
}

func (w *dailyFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *dailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.closeLocked()
}

func (w *dailyFileWriter) currentFileLocked() (*os.File, error) {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return nil, err
	}

	path := BackendLogPath(w.dir, w.now())
	if w.file != nil && w.current == path {
		return w.file, nil
	}
	if err := w.closeLocked(); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	w.file = file
	w.current = path
	return file, nil
}

func (w *dailyFileWriter) closeLocked() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.current = ""
	return err
}

func NewLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	return newLoggerWithClock(cfg, time.Now)
}

func NewManagedLogger(cfg config.LoggingConfig) (*zap.Logger, SyncLogger, error) {
	return newManagedLoggerWithClock(cfg, time.Now)
}

func newLoggerWithClock(cfg config.LoggingConfig, now dateNowFunc) (*zap.Logger, error) {
	logger, _, err := newManagedLoggerWithClock(cfg, now)
	return logger, err
}

func newManagedLoggerWithClock(cfg config.LoggingConfig, now dateNowFunc) (*zap.Logger, SyncLogger, error) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	outputs := make([]zapcore.WriteSyncer, 0, 2)
	outputs = append(outputs, zapcore.AddSync(os.Stdout))

	var writer *dailyFileWriter
	if cfg.FileEnabled {
		logDir := cfg.Path
		if strings.TrimSpace(logDir) == "" {
			logDir = defaultLogPath
		}
		writer = newDailyFileWriter(logDir, now)
		outputs = append(outputs, zapcore.AddSync(writer))
	}

	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(outputs...), zap.NewAtomicLevelAt(zap.DebugLevel))
	logger := zap.New(core, zap.AddCaller())
	if writer == nil {
		return logger, noopSyncer{}, nil
	}
	return logger, &managedSyncer{logger: logger, closer: writer}, nil
}

func BackendLogPath(path string, now time.Time) string {
	logDir := path
	if strings.TrimSpace(logDir) == "" {
		logDir = defaultLogPath
	}
	return filepath.Join(logDir, "backend-"+now.Format("2006-01-02")+".log")
}

func CleanupOldLogs(path string, maxDays int, now time.Time) error {
	if maxDays < 1 {
		return nil
	}

	logDir := path
	if strings.TrimSpace(logDir) == "" {
		logDir = defaultLogPath
	}

	info, err := os.Stat(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	cutoff := now.Add(-time.Duration(maxDays) * 24 * time.Hour)
	pattern := filepath.Join(logDir, "backend-*.log")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, filePath := range matches {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if fileInfo.IsDir() {
			continue
		}
		base := filepath.Base(filePath)
		if !strings.HasPrefix(base, "backend-") || !strings.HasSuffix(base, ".log") {
			continue
		}
		dateStr := strings.TrimSuffix(strings.TrimPrefix(base, "backend-"), ".log")
		fileDate, err := time.ParseInLocation("2006-01-02", dateStr, now.Location())
		if err != nil {
			continue
		}
		if !fileDate.Before(cutoff) {
			continue
		}
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func AccessLogger(base *zap.Logger) gin.HandlerFunc {
	if base == nil {
		base = zap.NewNop()
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		isStaticAsset := isStaticAssetPath(path)
		requestLogger := base.With(
			zap.String("method", c.Request.Method),
			zap.String("path", path),
		)
		c.Set(requestLoggerContextKey, requestLogger)
		c.Request = c.Request.WithContext(WithContext(c.Request.Context(), requestLogger))

		c.Next()

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.Int("response_bytes", c.Writer.Size()),
		}

		if !isStaticAsset {
			fields = append(fields,
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("referer", c.Request.Referer()),
				zap.Int64("body_bytes", c.Request.ContentLength),
			)
		}

		if route := c.FullPath(); route != "" && !isStaticAsset {
			fields = append(fields, zap.String("route", route))
		}
		if userID, ok := c.Get("user_id"); ok {
			fields = append(fields, zap.Any("user_id", userID))
		}
		if code, ok := c.Get(logCodeContextKey); ok {
			fields = append(fields, zap.Any("code", code))
		}
		if msg, ok := c.Get(logMessageContextKey); ok {
			fields = append(fields, zap.Any("error_message", msg))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("gin_errors", c.Errors.String()))
		}

		entry := requestLogger.With(fields...)
		status := c.Writer.Status()
		switch {
		case status >= http.StatusInternalServerError:
			entry.Error("http request completed")
		case status >= http.StatusBadRequest:
			entry.Warn("http request completed")
		default:
			entry.Info("http request completed")
		}
	}
}

func isStaticAssetPath(path string) bool {
	normalized := strings.ToLower(strings.TrimSpace(path))
	if normalized == "" {
		return false
	}
	if strings.HasPrefix(normalized, "/assets/") {
		return true
	}
	switch normalized {
	case "/favicon.ico", "/favicon.svg", "/icons.svg", "/manifest.webmanifest", "/robots.txt", "/site.webmanifest":
		return true
	}
	ext := filepath.Ext(normalized)
	_, ok := staticAssetExtensions[ext]
	return ok
}

func Recovery(base *zap.Logger) gin.HandlerFunc {
	if base == nil {
		base = zap.NewNop()
	}

	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger := LoggerFromContext(c)
		if logger == nil {
			logger = base
		}
		SetResponseLogMeta(c, "INTERNAL_ERROR", "panic recovered")
		logger.Error("panic recovered",
			zap.Any("panic", recovered),
			zap.ByteString("stack", debug.Stack()),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, requestLoggerContextKey, logger)
}

func Logger(ctx context.Context, fallback *zap.Logger) *zap.Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(requestLoggerContextKey).(*zap.Logger); ok && logger != nil {
			return logger
		}
	}
	if fallback == nil {
		return zap.NewNop()
	}
	return fallback
}

func LoggerFromContext(c *gin.Context) *zap.Logger {
	if c != nil {
		if logger, ok := c.Get(requestLoggerContextKey); ok {
			if zl, ok := logger.(*zap.Logger); ok && zl != nil {
				return zl
			}
		}
	}
	return zap.NewNop()
}

func SetResponseLogMeta(c *gin.Context, code string, message string) {
	if c == nil {
		return
	}
	if code != "" {
		c.Set(logCodeContextKey, code)
	}
	if message != "" {
		c.Set(logMessageContextKey, message)
	}
}

type noopSyncer struct{}

func (noopSyncer) Sync() error {
	return nil
}

type managedSyncer struct {
	logger *zap.Logger
	closer io.Closer
}

func (m *managedSyncer) Sync() error {
	var errs []error
	if m.logger != nil {
		if err := m.logger.Sync(); err != nil && !isIgnorableSyncError(err) {
			errs = append(errs, err)
		}
	}
	if m.closer != nil {
		if err := m.closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func isIgnorableSyncError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "invalid argument")
}
