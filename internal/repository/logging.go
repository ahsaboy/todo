package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go.uber.org/zap"

	appLogging "todo/internal/logging"
)

const repositoryLoggerName = "repository"

type dbOperationLog struct {
	logger *zap.Logger
	start  time.Time
	fields []zap.Field
}

func beginDBOperation(ctx context.Context, repositoryName, operation string, fields ...zap.Field) *dbOperationLog {
	baseLogger := appLogging.Logger(ctx, nil).Named(repositoryLoggerName)
	baseFields := make([]zap.Field, 0, len(fields)+2)
	baseFields = append(baseFields,
		zap.String("repository", repositoryName),
		zap.String("operation", operation),
	)
	baseFields = append(baseFields, fields...)

	return &dbOperationLog{
		logger: baseLogger,
		start:  time.Now(),
		fields: baseFields,
	}
}

func (l *dbOperationLog) complete(err error, resultFields ...zap.Field) {
	fields := make([]zap.Field, 0, len(l.fields)+len(resultFields)+2)
	fields = append(fields, l.fields...)
	fields = append(fields, zap.Duration("duration", time.Since(l.start)))
	fields = append(fields, resultFields...)

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.logger.Error("repository db operation", fields...)
		return
	}

	l.logger.Info("repository db operation", fields...)
}

func rowsAffected(result sql.Result) int64 {
	if result == nil {
		return 0
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0
	}
	return rows
}

func lastInsertID(result sql.Result) int64 {
	if result == nil {
		return 0
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0
	}
	return id
}

func apiKeyFingerprint(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}
