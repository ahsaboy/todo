package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo/internal/middleware"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type ReminderLogHandler struct {
	svc service.ReminderLogServiceInterface
}

func NewReminderLogHandler(svc service.ReminderLogServiceInterface) *ReminderLogHandler {
	return &ReminderLogHandler{svc: svc}
}

func (h *ReminderLogHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	limit := parsePositiveInt(c.DefaultQuery("limit", "20"), 20)
	if limit > 100 {
		limit = 100
	}

	logs, total, err := h.svc.List(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "reminder_log.list", err)
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

