package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type ReminderConfigHandler struct {
	svc *service.ReminderConfigService
}

func NewReminderConfigHandler(svc *service.ReminderConfigService) *ReminderConfigHandler {
	return &ReminderConfigHandler{svc: svc}
}

// Create 创建提醒配置
// @Summary      创建提醒配置
// @Description  为当前用户创建一个新的通知渠道配置
// @Tags         reminder-config
// @Accept       json
// @Produce      json
// @Param        body body models.CreateReminderConfigRequest true "配置信息"
// @Success      201  {object} utils.SuccessResponse{data=models.UserReminderConfig}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/reminder-configs [post]
func (h *ReminderConfigHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	var req models.CreateReminderConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	cfg, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to create config", utils.CodeInternalError)
		return
	}

	utils.RespondCreated(c, views.UserReminderConfigView(cfg, timezone.Get()))
}

// List 列出提醒配置
// @Summary      列出提醒配置
// @Description  获取当前用户的所有通知渠道配置
// @Tags         reminder-config
// @Produce      json
// @Success      200  {object} utils.SuccessResponse{data=[]models.UserReminderConfig}
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/reminder-configs [get]
func (h *ReminderConfigHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	configs, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to list configs", utils.CodeInternalError)
		return
	}

	utils.RespondSuccess(c, views.UserReminderConfigsView(configs, timezone.Get()))
}

// GetByID 获取单个提醒配置
// @Summary      获取单个提醒配置
// @Tags         reminder-config
// @Produce      json
// @Param        id path int true "配置 ID"
// @Success      200  {object} utils.SuccessResponse{data=models.UserReminderConfig}
// @Failure      404  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/reminder-configs/{id} [get]
func (h *ReminderConfigHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid config id", utils.CodeInvalidInput)
		return
	}

	cfg, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to get config", utils.CodeInternalError)
		return
	}
	if cfg == nil {
		utils.RespondError(c, http.StatusNotFound, "config not found", utils.CodeNotFound)
		return
	}

	utils.RespondSuccess(c, views.UserReminderConfigView(cfg, timezone.Get()))
}

// Update 更新提醒配置
// @Summary      更新提醒配置
// @Tags         reminder-config
// @Accept       json
// @Produce      json
// @Param        id   path int                             true "配置 ID"
// @Param        body body models.UpdateReminderConfigRequest true "要更新的字段"
// @Success      200  {object} utils.SuccessResponse{data=models.UserReminderConfig}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      404  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/reminder-configs/{id} [put]
func (h *ReminderConfigHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid config id", utils.CodeInvalidInput)
		return
	}

	var req models.UpdateReminderConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	cfg, err := h.svc.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update config", utils.CodeInternalError)
		return
	}
	if cfg == nil {
		utils.RespondError(c, http.StatusNotFound, "config not found", utils.CodeNotFound)
		return
	}

	utils.RespondSuccess(c, views.UserReminderConfigView(cfg, timezone.Get()))
}

// Delete 删除提醒配置
// @Summary      删除提醒配置
// @Tags         reminder-config
// @Produce      json
// @Param        id path int true "配置 ID"
// @Success      200  {object} map[string]interface{}
// @Failure      404  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/reminder-configs/{id} [delete]
func (h *ReminderConfigHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid config id", utils.CodeInvalidInput)
		return
	}

	deleted, err := h.svc.Delete(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to delete config", utils.CodeInternalError)
		return
	}
	if !deleted {
		utils.RespondError(c, http.StatusNotFound, "config not found", utils.CodeNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "config deleted"})
}
