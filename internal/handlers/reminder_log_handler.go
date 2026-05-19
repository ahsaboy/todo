package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo/internal/middleware"
	"todo/internal/repository"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type ReminderLogHandler struct {
	repo *repository.ReminderLogRepo
}

func NewReminderLogHandler(repo *repository.ReminderLogRepo) *ReminderLogHandler {
	return &ReminderLogHandler{repo: repo}
}

func (h *ReminderLogHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	limit := parsePositiveInt(c.DefaultQuery("limit", "20"), 20)
	if limit > 100 {
		limit = 100
	}

	logs, total, err := h.repo.ListByUserID(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.RespondInternalError(c, "failed to list reminder logs", err)
		return
	}

	utils.RespondPaginated(c, views.ReminderLogsView(logs, timezone.Get()), page, limit, total)
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
