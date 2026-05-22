package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todo/internal/logging"
	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// Create 创建任务
// @Summary      创建一个新任务
// @Description  创建一个新任务，支持设置标题、优先级、截止时间、提醒时间、重复规则等；若设置了提醒时间（remind_at），则必须至少存在一个已启用的通知渠道
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        body body models.CreateTaskRequest true "任务信息"
// @Success      201  {object} utils.SuccessResponse{data=models.Task}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	task, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrReminderChannelMissing) || errors.Is(err, service.ErrInvalidTime) {
			utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
			return
		}
		utils.RespondInternalError(c, "failed to create task", err)
		return
	}

	utils.RespondCreated(c, views.TaskView(task, timezone.Get()))
}

// GetByID 获取单个任务
// @Summary      根据 ID 获取任务
// @Description  根据任务 ID 返回任务详情
// @Tags         tasks
// @Produce      json
// @Param        id path int true "任务 ID"
// @Success      200  {object} utils.SuccessResponse{data=models.Task}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      404  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid task id", utils.CodeInvalidInput)
		return
	}

	task, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondInternalError(c, "failed to get task", err)
		return
	}
	if task == nil {
		utils.RespondError(c, http.StatusNotFound, "task not found", utils.CodeNotFound)
		return
	}

	utils.RespondSuccess(c, views.TaskView(task, timezone.Get()))
}

// List 获取任务列表
// @Summary      获取任务列表（分页、筛选、排序）
// @Description  支持分页查询、按状态/优先级筛选、按关键字搜索、自定义排序
// @Tags         tasks
// @Produce      json
// @Param        page      query int    false "页码"       default(1)
// @Param        limit     query int    false "每页数量"    default(20)
// @Param        sort      query string false "排序字段"    default(created_at)  Enums(created_at, updated_at, due_at, priority, task_center)
// @Param        order     query string false "排序方向"    default(desc)        Enums(asc, desc)
// @Param        status    query string false "任务状态"                         Enums(pending, completed, all)
// @Param        priority  query int    false "优先级筛选"                       Enums(1, 2, 3)
// @Param        due_before query string false "截止时间上限 (RFC3339 UTC，例如 2026-05-10T10:30:00Z)"
// @Param        due_after  query string false "截止时间下限 (RFC3339 UTC，例如 2026-05-10T10:30:00Z)"
// @Param        search    query string false "搜索关键字（标题/描述）"
// @Success      200  {object} utils.PaginatedResponse{data=[]models.Task}
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	priority, _ := strconv.Atoi(c.Query("priority"))

	dueBefore, dueAfter := c.Query("due_before"), c.Query("due_after")
	if dueBefore != "" {
		normalized, err := utils.NormalizeAPITime(dueBefore, timezone.Get())
		if err != nil {
			utils.RespondError(c, http.StatusBadRequest, "due_before: "+err.Error(), utils.CodeInvalidInput)
			return
		}
		dueBefore = normalized
	}
	if dueAfter != "" {
		normalized, err := utils.NormalizeAPITime(dueAfter, timezone.Get())
		if err != nil {
			utils.RespondError(c, http.StatusBadRequest, "due_after: "+err.Error(), utils.CodeInvalidInput)
			return
		}
		dueAfter = normalized
	}

	filters := models.TaskFilters{
		Status:    c.Query("status"),
		Priority:  priority,
		DueBefore: dueBefore,
		DueAfter:  dueAfter,
		Search:    c.Query("search"),
	}

	sortField := c.DefaultQuery("sort", "created_at")
	sortOrder := c.DefaultQuery("order", "desc")

	tasks, total, err := h.svc.List(c.Request.Context(), userID, filters, page, limit, sortField, sortOrder)
	if err != nil {
		utils.RespondInternalError(c, "failed to list tasks", err)
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}
	utils.RespondPaginated(c, views.TasksView(tasks, timezone.Get()), page, limit, total)
}

// Update 更新任务
// @Summary      更新任务信息
// @Description  部分更新任务字段，只需传入需要修改的字段
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path int                true "任务 ID"
// @Param        body body models.UpdateTaskRequest true "要更新的字段"
// @Success      200  {object} utils.SuccessResponse{data=models.Task}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      404  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tasks/{id} [put]
func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid task id", utils.CodeInvalidInput)
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	task, err := h.svc.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTime) || errors.Is(err, service.ErrReminderChannelMissing) {
			utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
			return
		}
		utils.RespondInternalError(c, "failed to update task", err)
		return
	}
	if task == nil {
		utils.RespondError(c, http.StatusNotFound, "task not found", utils.CodeNotFound)
		return
	}

	utils.RespondSuccess(c, views.TaskView(task, timezone.Get()))
}

// Delete 删除任务
// @Summary      删除指定任务
// @Description  根据任务 ID 永久删除一个任务
// @Tags         tasks
// @Produce      json
// @Param        id path int true "任务 ID"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      404  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tasks/{id} [delete]
func (h *TaskHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid task id", utils.CodeInvalidInput)
		return
	}

	deleted, err := h.svc.Delete(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondInternalError(c, "failed to delete task", err)
		return
	}
	if !deleted {
		utils.RespondError(c, http.StatusNotFound, "task not found", utils.CodeNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "task deleted"})
}

// ToggleComplete 切换任务完成状态
// @Summary      切换任务的完成/未完成状态
// @Description  将任务在已完成和未完成之间切换。如果是重复任务且标记为完成，会自动创建下一次任务
// @Tags         tasks
// @Produce      json
// @Param        id path int true "任务 ID"
// @Success      200  {object} utils.SuccessResponse{data=models.Task}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      404  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/tasks/{id}/complete [patch]
func (h *TaskHandler) ToggleComplete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid task id", utils.CodeInvalidInput)
		return
	}

	task, err := h.svc.ToggleComplete(c.Request.Context(), userID, id)
	if err != nil {
		utils.RespondInternalError(c, "failed to toggle task", err)
		return
	}
	if task == nil {
		utils.RespondError(c, http.StatusNotFound, "task not found", utils.CodeNotFound)
		return
	}

	utils.RespondSuccess(c, views.TaskView(task, timezone.Get()))
}

// HealthCheck 健康检查
// @Summary      健康检查
// @Description  检查服务和数据库连接状态，用于 Docker 健康检查和负载均衡探测
// @Tags         health
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      503  {object} map[string]interface{}
// @Router       /api/v1/health [get]
func HealthCheck(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			logging.SetResponseLogMeta(c, "DATABASE_UNHEALTHY", "database connection failed")
			logging.LoggerFromContext(c).Error("health check failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "status": "unhealthy", "error": "database connection failed"})
			return
		}
		logging.SetResponseLogMeta(c, "OK", "healthy")
		c.JSON(http.StatusOK, gin.H{"success": true, "status": "healthy"})
	}
}
