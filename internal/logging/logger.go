package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"todo/internal/config"
)

const defaultLogPath = "./logs"

type appendFileWriter struct {
	path string
	mu   sync.Mutex
}

func (w *appendFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.Write(p)
}

func (w *appendFileWriter) Sync() error {
	return nil
}

func NewLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	level := parseLevel(cfg.Level)

	encoding := "json"
	if !strings.EqualFold(strings.TrimSpace(cfg.Format), "json") {
		encoding = "console"
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	encoder, err := newEncoder(encoding, encoderCfg)
	if err != nil {
		return nil, err
	}

	outputs := make([]zapcore.WriteSyncer, 0, 2)
	if cfg.Backend.ConsoleEnabled {
		outputs = append(outputs, zapcore.AddSync(os.Stdout))
	}
	if cfg.Backend.FileEnabled {
		logDir := cfg.Path
		if strings.TrimSpace(logDir) == "" {
			logDir = defaultLogPath
		}
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return nil, err
		}
		outputs = append(outputs, zapcore.AddSync(&appendFileWriter{
			path: BackendLogPath(logDir, time.Now()),
		}))
	}
	if len(outputs) == 0 {
		outputs = append(outputs, zapcore.AddSync(io.Discard))
	}

	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(outputs...), zap.NewAtomicLevelAt(level))
	return zap.New(core, zap.AddCaller()), nil
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func newEncoder(encoding string, cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	switch encoding {
	case "json":
		return zapcore.NewJSONEncoder(cfg), nil
	case "console":
		return zapcore.NewConsoleEncoder(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported encoding %q", encoding)
	}
}

func BackendLogPath(path string, now time.Time) string {
	logDir := path
	if strings.TrimSpace(logDir) == "" {
		logDir = defaultLogPath
	}
	return filepath.Join(logDir, "backend-"+now.Format("2006-01-02")+".log")
}

func FrontendLogPath(path string, now time.Time) string {
	logDir := path
	if strings.TrimSpace(logDir) == "" {
		logDir = defaultLogPath
	}
	return filepath.Join(logDir, "frontend-"+now.Format("2006-01-02")+".log")
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
	patterns := []string{
		filepath.Join(logDir, "backend-*.log"),
		filepath.Join(logDir, "frontend-*.log"),
	}

	for _, pattern := range patterns {
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
			var prefix string
			switch {
			case strings.HasPrefix(base, "backend-") && strings.HasSuffix(base, ".log"):
				prefix = "backend-"
			case strings.HasPrefix(base, "frontend-") && strings.HasSuffix(base, ".log"):
				prefix = "frontend-"
			default:
				continue
			}
			dateStr := strings.TrimSuffix(strings.TrimPrefix(base, prefix), ".log")
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
	}

	return nil
}
