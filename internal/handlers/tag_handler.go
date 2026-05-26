package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// Create 创建标签
// @Summary      创建标签
// @Description  为当前用户创建一个新标签;标签名同用户下唯一
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        body body models.CreateTagRequest true "标签信息"
// @Success      201  {object} utils.SuccessResponse{data=models.UserTag}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      409  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	tag, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTagColor) || errors.Is(err, service.ErrInvalidTagIcon) || errors.Is(err, service.ErrTagNameEmpty) {
			utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
			return
		}
		if errors.Is(err, repository.ErrTagNameTaken) {
			utils.RespondError(c, http.StatusConflict, err.Error(), utils.CodeInvalidInput)
			return
		}
		utils.RespondInternalError(c, "failed to create tag", err)
		return
	}

	utils.RespondCreated(c, views.UserTagView(tag, timezone.Get()))
}

// List 列出当前用户的所有标签
// @Summary      列出标签
// @Description  按 sort_order, id 顺序返回当前用户的所有标签
// @Tags         tags
// @Produce      json
// @Success      200  {object} utils.SuccessResponse{data=[]models.UserTag}
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	tags, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, "failed to list tags", err)
		return
	}
	utils.RespondSuccess(c, views.UserTagsView(tags, timezone.Get()))
}

// GetByID 获取单个标签
// @Summary      获取单个标签
// @Tags         tags
// @Produce      json
// @Param        id path int true "标签 ID"
// @Success      200  {object} utils.SuccessResponse{data=models.UserTag}
// @Failure      404  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tags/{id} [get]
func (h *TagHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid tag id", utils.CodeInvalidInput)
		return
	}

	tag, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondInternalError(c, "failed to get tag", err)
		return
	}
	if tag == nil {
		utils.RespondError(c, http.StatusNotFound, "tag not found", utils.CodeNotFound)
		return
	}
	utils.RespondSuccess(c, views.UserTagView(tag, timezone.Get()))
}

// Update 更新标签
// @Summary      更新标签
// @Description  改名时会同步更新该用户所有任务上引用的标签名(事务保证一致)
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path int                       true "标签 ID"
// @Param        body body models.UpdateTagRequest   true "要更新的字段"
// @Success      200  {object} utils.SuccessResponse{data=models.UserTag}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      404  {object} utils.ErrorResponse
// @Failure      409  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tags/{id} [put]
func (h *TagHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid tag id", utils.CodeInvalidInput)
		return
	}

	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	tag, err := h.svc.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTagColor) || errors.Is(err, service.ErrInvalidTagIcon) || errors.Is(err, service.ErrTagNameEmpty) {
			utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
			return
		}
		if errors.Is(err, repository.ErrTagNameTaken) {
			utils.RespondError(c, http.StatusConflict, err.Error(), utils.CodeInvalidInput)
			return
		}
		utils.RespondInternalError(c, "failed to update tag", err)
		return
	}
	if tag == nil {
		utils.RespondError(c, http.StatusNotFound, "tag not found", utils.CodeNotFound)
		return
	}
	utils.RespondSuccess(c, views.UserTagView(tag, timezone.Get()))
}

// Delete 删除标签
// @Summary      删除标签
// @Description  删除后,该用户所有任务上引用的此标签会被同步摘除
// @Tags         tags
// @Produce      json
// @Param        id path int true "标签 ID"
// @Success      200  {object} map[string]interface{}
// @Failure      404  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid tag id", utils.CodeInvalidInput)
		return
	}

	deleted, tasksAffected, err := h.svc.Delete(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondInternalError(c, "failed to delete tag", err)
		return
	}
	if !deleted {
		utils.RespondError(c, http.StatusNotFound, "tag not found", utils.CodeNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "tag deleted",
		"tasks_affected":  tasksAffected,
	})
}
