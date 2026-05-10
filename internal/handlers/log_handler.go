package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todo/internal/config"
	"todo/internal/logging"
	"todo/internal/utils"
)

const (
	maxFrontendLogRequestBytes = 64 * 1024
	maxFrontendLogEntries      = 20
	maxFrontendLogMessageBytes = 2048
	maxFrontendLogStackBytes   = 8192
	maxFrontendLogContextBytes = 4096
)

type FrontendLogEntry struct {
	Level   string          `json:"level"`
	Message string          `json:"message"`
	Stack   string          `json:"stack,omitempty"`
	Context json.RawMessage `json:"context,omitempty"`
}

type RuntimeConfigResponse struct {
	Logging RuntimeConfigLogging `json:"logging"`
}

type RuntimeConfigLogging struct {
	Frontend RuntimeConfigFrontend `json:"frontend"`
}

type RuntimeConfigFrontend struct {
	ConsoleEnabled bool   `json:"console_enabled"`
	FileEnabled    bool   `json:"file_enabled"`
	Level          string `json:"level"`
}

type LogHandler struct {
	cfg    config.LoggingConfig
	logger *zap.Logger
}

func NewLogHandler(cfg config.LoggingConfig, logger *zap.Logger) *LogHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LogHandler{cfg: cfg, logger: logger}
}

// RuntimeConfig returns the safe runtime configuration for the browser.
func (h *LogHandler) RuntimeConfig(c *gin.Context) {
	c.JSON(http.StatusOK, RuntimeConfigResponse{
		Logging: RuntimeConfigLogging{
			Frontend: RuntimeConfigFrontend{
				ConsoleEnabled: h.cfg.Frontend.ConsoleEnabled,
				FileEnabled:    h.cfg.Frontend.FileEnabled,
				Level:          h.cfg.Frontend.Level,
			},
		},
	})
}

// FrontendLogs accepts single or batched frontend log entries.
func (h *LogHandler) FrontendLogs(c *gin.Context) {
	if !h.cfg.Frontend.FileEnabled {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	entries, err := decodeFrontendLogEntries(c, maxFrontendLogRequestBytes)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if err := validateFrontendLogEntries(entries); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if err := writeFrontendLogEntries(h.cfg, entries, time.Now()); err != nil {
		h.logger.Error("failed to write frontend log", zap.Error(err))
		utils.RespondError(c, http.StatusInternalServerError, "failed to store frontend log", utils.CodeInternalError)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

func decodeFrontendLogEntries(c *gin.Context, maxBytes int64) ([]FrontendLogEntry, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
	defer c.Request.Body.Close()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, errors.New("request body cannot be empty")
	}

	var batch []FrontendLogEntry
	if err := json.Unmarshal(body, &batch); err == nil {
		if len(batch) == 0 {
			return nil, errors.New("request body cannot be empty")
		}
		if len(batch) > maxFrontendLogEntries {
			return nil, errors.New("too many log entries")
		}
		return batch, nil
	}

	var single FrontendLogEntry
	if err := json.Unmarshal(body, &single); err != nil {
		return nil, errors.New("invalid frontend log payload")
	}

	return []FrontendLogEntry{single}, nil
}

func validateFrontendLogEntries(entries []FrontendLogEntry) error {
	for _, entry := range entries {
		if err := validateFrontendLogEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

func validateFrontendLogEntry(entry FrontendLogEntry) error {
	switch strings.ToLower(strings.TrimSpace(entry.Level)) {
	case "debug", "info", "warn", "error":
	default:
		return errors.New("invalid log level")
	}

	if strings.TrimSpace(entry.Message) == "" {
		return errors.New("message is required")
	}
	if len(entry.Message) > maxFrontendLogMessageBytes {
		return errors.New("message is too long")
	}
	if len(entry.Stack) > maxFrontendLogStackBytes {
		return errors.New("stack is too long")
	}
	if len(entry.Context) > maxFrontendLogContextBytes {
		return errors.New("context is too long")
	}

	return nil
}

type frontendLogRecord struct {
	Timestamp string          `json:"timestamp"`
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	Stack     string          `json:"stack,omitempty"`
	Context   json.RawMessage `json:"context,omitempty"`
}

func writeFrontendLogEntries(cfg config.LoggingConfig, entries []FrontendLogEntry, now time.Time) error {
	logPath := logging.FrontendLogPath(cfg.Path, now)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, entry := range entries {
		record := frontendLogRecord{
			Timestamp: now.UTC().Format(time.RFC3339Nano),
			Level:     strings.ToLower(strings.TrimSpace(entry.Level)),
			Message:   entry.Message,
			Stack:     entry.Stack,
			Context:   entry.Context,
		}
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}

	return nil
}
