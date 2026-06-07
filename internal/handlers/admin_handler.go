package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"todo/internal/config"
	"todo/internal/logging"
	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type AdminHandler struct {
	db                 *sql.DB
	userRepo           repository.UserRepository
	apiKeyRepo         repository.APIKeyRepository
	taskRepo           repository.TaskRepository
	reminderConfigRepo repository.ReminderConfigRepository
	reminderLogRepo    repository.ReminderLogRepository
	auditLogRepo       repository.AuditLogRepository
	cfg                *config.Config
	appConfigSvc       *service.AppConfigService
	reminderSvc        *service.ReminderService
	emailSvc           service.EmailServiceInterface
	lockedKeys         map[string]bool
	applyI18n          func(string)
	reloadOAuth        func()
}

// gracePeriodSetter 是支持运行时更新宽限期的 TaskRepo 子集接口。
type gracePeriodSetter interface {
	SetGracePeriod(d time.Duration)
}

func NewAdminHandler(
	db *sql.DB,
	userRepo repository.UserRepository,
	apiKeyRepo repository.APIKeyRepository,
	taskRepo repository.TaskRepository,
	reminderConfigRepo repository.ReminderConfigRepository,
	reminderLogRepo repository.ReminderLogRepository,
	auditLogRepo repository.AuditLogRepository,
	cfg *config.Config,
	appConfigSvc *service.AppConfigService,
	reminderSvc *service.ReminderService,
	emailSvc service.EmailServiceInterface,
	lockedKeys map[string]bool,
	applyI18n func(string),
	reloadOAuth func(),
) *AdminHandler {
	return &AdminHandler{
		db:                 db,
		userRepo:           userRepo,
		apiKeyRepo:         apiKeyRepo,
		taskRepo:           taskRepo,
		reminderConfigRepo: reminderConfigRepo,
		reminderLogRepo:    reminderLogRepo,
		auditLogRepo:       auditLogRepo,
		cfg:                cfg,
		appConfigSvc:       appConfigSvc,
		reminderSvc:        reminderSvc,
		emailSvc:           emailSvc,
		lockedKeys:         lockedKeys,
		applyI18n:          applyI18n,
		reloadOAuth:        reloadOAuth,
	}
}

type adminLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type adminLoginResponse struct {
	User    models.UserResponse `json:"user"`
	APIKey  string              `json:"api_key"`
	IsAdmin bool                `json:"is_admin"`
}

func (h *AdminHandler) AdminLogin(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	user, err := h.userRepo.GetByUsername(c.Request.Context(), req.Account)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "auth.authenticate", err)
		return
	}
	if user == nil {
		user, err = h.userRepo.GetByEmail(c.Request.Context(), req.Account)
		if err != nil {
			utils.RespondLocalizedInternalError(c, "auth.authenticate", err)
			return
		}
		if user == nil {
			utils.RespondLocalizedError(c, http.StatusUnauthorized, "auth.invalid_credentials")
			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "auth.invalid_credentials")
		return
	}

	if !user.IsAdmin {
		utils.RespondLocalizedError(c, http.StatusForbidden, "admin.access_required")
		return
	}

	loginKey, err := h.generateAdminLoginKey(c.Request.Context(), user.ID)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "auth.generate_session", err)
		return
	}

	loc := timezone.Get()
	resp := adminLoginResponse{
		User:    views.UserResponseView(user.ToResponse(), loc),
		APIKey:  loginKey,
		IsAdmin: true,
	}
	utils.RespondSuccess(c, resp)
}

func (h *AdminHandler) generateAdminLoginKey(ctx context.Context, userID int64) (string, error) {
	_, _ = h.apiKeyRepo.CleanupExpiredLoginKeys(ctx, userID)

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	apiKey := base64.URLEncoding.EncodeToString(b)
	keyHash := middleware.HashAPIKey(apiKey)

	if _, err := h.apiKeyRepo.Create(ctx, userID, keyHash, "admin-login"); err != nil {
		return "", err
	}
	return apiKey, nil
}

type adminStats struct {
	TotalUsers            int64   `json:"total_users"`
	TotalTasks            int64   `json:"total_tasks"`
	CompletedTasks        int64   `json:"completed_tasks"`
	TotalReminderConfigs  int64   `json:"total_reminder_configs"`
	TotalReminderLogs     int64   `json:"total_reminder_logs"`
	TodayNewTasks         int64   `json:"today_new_tasks"`
	TodayNewUsers         int64   `json:"today_new_users"`
	ActiveUsers7d         int64   `json:"active_users_7d"`
	PriorityDist          []adminPriorityCount `json:"priority_dist"`
	CompletionTrend       []adminDayCount      `json:"completion_trend"`
	TopTags               []adminTagCount      `json:"top_tags"`
}

type adminPriorityCount struct {
	Priority int   `json:"priority"`
	Count    int64 `json:"count"`
}

type adminTagCount struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()
	var stats adminStats

	const statsQuery = `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM tasks) AS total_tasks,
			(SELECT COUNT(*) FROM tasks WHERE completed = 1) AS completed_tasks,
			(SELECT COUNT(*) FROM user_reminder_configs) AS total_reminder_configs,
			(SELECT COUNT(*) FROM reminder_logs) AS total_reminder_logs,
			(SELECT COUNT(*) FROM tasks WHERE DATE(created_at) = DATE('now')) AS today_new_tasks,
			(SELECT COUNT(*) FROM users WHERE DATE(created_at) = DATE('now')) AS today_new_users,
			(SELECT COUNT(DISTINCT user_id) FROM tasks WHERE created_at >= DATE('now', '-7 days')) AS active_users_7d
	`
	if err := h.db.QueryRowContext(ctx, statsQuery).Scan(
		&stats.TotalUsers,
		&stats.TotalTasks,
		&stats.CompletedTasks,
		&stats.TotalReminderConfigs,
		&stats.TotalReminderLogs,
		&stats.TodayNewTasks,
		&stats.TodayNewUsers,
		&stats.ActiveUsers7d,
	); err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}

	// 任务优先级分布
	rows1, err := h.db.QueryContext(ctx, `SELECT COALESCE(priority, 0) as p, COUNT(*) FROM tasks GROUP BY p ORDER BY p`)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}
	defer rows1.Close()
	for rows1.Next() {
		var pc adminPriorityCount
		if err := rows1.Scan(&pc.Priority, &pc.Count); err == nil {
			stats.PriorityDist = append(stats.PriorityDist, pc)
		}
	}
	if err := rows1.Err(); err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}

	// 近7天每日完成任务数
	rows2, err := h.db.QueryContext(ctx, `
		SELECT DATE(updated_at) as day, COUNT(*) as cnt
		FROM tasks
		WHERE completed = 1 AND updated_at >= DATE('now', '-7 days')
		GROUP BY day ORDER BY day`)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var dc adminDayCount
		if err := rows2.Scan(&dc.Date, &dc.Count); err == nil {
			stats.CompletionTrend = append(stats.CompletionTrend, dc)
		}
	}
	if err := rows2.Err(); err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}

	// 热门标签 Top 8
	rows3, err := h.db.QueryContext(ctx, `
		SELECT tag, COUNT(*) as cnt
		FROM (
			SELECT json_each.value as tag
			FROM tasks, json_each(tasks.tags)
			WHERE tasks.tags IS NOT NULL AND tasks.tags != '[]'
		)
		GROUP BY tag ORDER BY cnt DESC LIMIT 8`)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}
	defer rows3.Close()
	for rows3.Next() {
		var tc adminTagCount
		if err := rows3.Scan(&tc.Tag, &tc.Count); err == nil {
			stats.TopTags = append(stats.TopTags, tc)
		}
	}
	if err := rows3.Err(); err != nil {
		utils.RespondLocalizedInternalError(c, "system.stats", err)
		return
	}

	utils.RespondSuccess(c, stats)
}

type adminTrends struct {
	TasksPerDay        []adminDayCount `json:"tasks_per_day"`
	UsersPerDay        []adminDayCount `json:"users_per_day"`
	ReminderStatusDist map[string]int  `json:"reminder_status_dist"`
}

type adminDayCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (h *AdminHandler) GetTrends(c *gin.Context) {
	ctx := c.Request.Context()
	trends := adminTrends{
		ReminderStatusDist: make(map[string]int),
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT DATE(created_at) as day, COUNT(*) as cnt
		FROM tasks
		WHERE created_at >= DATE('now', '-30 days')
		GROUP BY day ORDER BY day`)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.task_trends", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var dc adminDayCount
		if err := rows.Scan(&dc.Date, &dc.Count); err == nil {
			trends.TasksPerDay = append(trends.TasksPerDay, dc)
		}
	}
	if err := rows.Err(); err != nil {
		utils.RespondLocalizedInternalError(c, "system.task_trends", err)
		return
	}

	rows2, err := h.db.QueryContext(ctx, `
		SELECT DATE(created_at) as day, COUNT(*) as cnt
		FROM users
		WHERE created_at >= DATE('now', '-30 days')
		GROUP BY day ORDER BY day`)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "user.list", err)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var dc adminDayCount
		if err := rows2.Scan(&dc.Date, &dc.Count); err == nil {
			trends.UsersPerDay = append(trends.UsersPerDay, dc)
		}
	}
	if err := rows2.Err(); err != nil {
		utils.RespondLocalizedInternalError(c, "user.list", err)
		return
	}

	rows3, err := h.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM reminder_logs GROUP BY status`)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.reminder_status", err)
		return
	}
	defer rows3.Close()
	for rows3.Next() {
		var status string
		var count int
		if err := rows3.Scan(&status, &count); err == nil {
			trends.ReminderStatusDist[status] = count
		}
	}
	if err := rows3.Err(); err != nil {
		utils.RespondLocalizedInternalError(c, "system.reminder_status", err)
		return
	}

	utils.RespondSuccess(c, trends)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	search := c.Query("search")

	users, total, err := h.userRepo.ListAll(c.Request.Context(), page, limit, search)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "user.list", err)
		return
	}
	loc := timezone.Get()
	type adminUserListItem struct {
		models.UserResponse
		IsAdmin bool `json:"is_admin"`
	}
	result := make([]adminUserListItem, 0, len(users))
	for _, u := range users {
		result = append(result, adminUserListItem{
			UserResponse: views.UserResponseView(u.ToResponse(), loc),
			IsAdmin:      u.IsAdmin,
		})
	}
	utils.RespondPaginated(c, result, page, limit, total)
}

type adminUserDetail struct {
	models.UserResponse
	IsAdmin     bool  `json:"is_admin"`
	TaskCount   int64 `json:"task_count"`
	APIKeyCount int64 `json:"api_key_count"`
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "user.invalid_id")
		return
	}
	ctx := c.Request.Context()
	u, err := h.userRepo.GetByID(ctx, id)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "user.get", err)
		return
	}
	if u == nil {
		utils.RespondLocalizedError(c, http.StatusNotFound, "user.not_found")
		return
	}

	var taskCount, keyCount int64
	if err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = ?", id).Scan(&taskCount); err != nil {
		utils.RespondLocalizedInternalError(c, "user.count_tasks", err)
		return
	}
	if err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_api_keys WHERE user_id = ?", id).Scan(&keyCount); err != nil {
		utils.RespondLocalizedInternalError(c, "user.count_api_keys", err)
		return
	}

	loc := timezone.Get()
	detail := adminUserDetail{
		UserResponse: views.UserResponseView(u.ToResponse(), loc),
		IsAdmin:      u.IsAdmin,
		TaskCount:    taskCount,
		APIKeyCount:  keyCount,
	}
	utils.RespondSuccess(c, detail)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "user.invalid_id")
		return
	}
	if err := h.userRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.RespondLocalizedError(c, http.StatusNotFound, "user.not_found")
		} else {
			utils.RespondLocalizedInternalError(c, "user.delete", err)
		}
		return
	}
	h.addAuditLog(c, "delete_user", "user", &id, "")
	c.Status(http.StatusNoContent)
}

type adminResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

func (h *AdminHandler) ForceResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "user.invalid_id")
		return
	}
	var req adminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "user.hash_password", err)
		return
	}
	if err := h.userRepo.ForceResetPassword(c.Request.Context(), id, string(hash)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.RespondLocalizedError(c, http.StatusNotFound, "user.not_found")
		} else {
			utils.RespondLocalizedInternalError(c, "user.reset_password", err)
		}
		return
	}
	h.addAuditLog(c, "reset_password", "user", &id, "")
	utils.RespondSuccess(c, gin.H{"ok": true})
}

type adminToggleAdminRequest struct {
	IsAdmin bool `json:"is_admin"`
}

func (h *AdminHandler) ToggleAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "user.invalid_id")
		return
	}
	var req adminToggleAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}
	// Prevent admin from demoting themselves
	if !req.IsAdmin && c.GetInt64("user_id") == id {
		utils.RespondLocalizedError(c, http.StatusForbidden, "admin.cannot_remove_own_privilege")
		return
	}
	if err := h.userRepo.SetIsAdmin(c.Request.Context(), id, req.IsAdmin); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.RespondLocalizedError(c, http.StatusNotFound, "user.not_found")
		} else {
			utils.RespondLocalizedInternalError(c, "user.update_admin", err)
		}
		return
	}
	action := "promote_admin"
	if !req.IsAdmin {
		action = "demote_admin"
	}
	h.addAuditLog(c, action, "user", &id, "")
	utils.RespondSuccess(c, gin.H{"ok": true})
}

func (h *AdminHandler) ListAllTasks(c *gin.Context) {
	page, limit := parsePaginationParams(c)

	var userID int64
	if raw := c.Query("user_id"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			userID = v
		}
	}

	filters := models.TaskFilters{
		Status:   c.Query("status"),
		Priority: 0,
	}
	if raw := c.Query("priority"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			filters.Priority = v
		}
	}

	tasks, total, err := h.taskRepo.ListAll(c.Request.Context(), userID, filters, page, limit)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "task.list", err)
		return
	}
	ids := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.UserID)
	}
	usernames := h.lookupUsernames(c.Request.Context(), uniqueUserIDs(ids))
	loc := timezone.Get()
	_ = loc // timestamps are already formatted by ListAll
	utils.RespondPaginated(c, adminTasksView(tasks, usernames), page, limit, total)
}

func (h *AdminHandler) AdminToggleComplete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "task.invalid_id")
		return
	}
	task, err := h.taskRepo.AdminToggleComplete(c.Request.Context(), id)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "task.toggle", err)
		return
	}
	if task == nil {
		utils.RespondLocalizedError(c, http.StatusNotFound, "task.not_found")
		return
	}
	h.addAuditLog(c, "toggle_task_complete", "task", &id, fmt.Sprintf("completed=%v", task.Completed))
	usernames := h.lookupUsernames(c.Request.Context(), []int64{task.UserID})
	utils.RespondSuccess(c, singleAdminTaskView(task, usernames))
}

func (h *AdminHandler) AdminUpdateTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "task.invalid_id")
		return
	}
	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}
	task, err := h.taskRepo.AdminUpdate(c.Request.Context(), id, req)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "task.update", err)
		return
	}
	if task == nil {
		utils.RespondLocalizedError(c, http.StatusNotFound, "task.not_found")
		return
	}
	h.addAuditLog(c, "update_task", "task", &id, "")
	usernames := h.lookupUsernames(c.Request.Context(), []int64{task.UserID})
	utils.RespondSuccess(c, singleAdminTaskView(task, usernames))
}

func (h *AdminHandler) AdminDeleteTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "task.invalid_id")
		return
	}
	deleted, err := h.taskRepo.AdminDelete(c.Request.Context(), id)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "task.delete", err)
		return
	}
	if !deleted {
		utils.RespondLocalizedError(c, http.StatusNotFound, "task.not_found")
		return
	}
	h.addAuditLog(c, "delete_task", "task", &id, "")
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) ListAllReminderConfigs(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	configs, total, err := h.reminderConfigRepo.ListAll(c.Request.Context(), page, limit)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "reminder_config.list", err)
		return
	}
	ids := make([]int64, 0, len(configs))
	for _, c := range configs {
		ids = append(ids, c.UserID)
	}
	usernames := h.lookupUsernames(c.Request.Context(), uniqueUserIDs(ids))
	utils.RespondPaginated(c, adminReminderConfigsView(configs, usernames), page, limit, total)
}

func (h *AdminHandler) AdminToggleReminderConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "config.invalid_id")
		return
	}
	found, err := h.reminderConfigRepo.AdminToggleEnabled(c.Request.Context(), id)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "reminder_config.toggle", err)
		return
	}
	if !found {
		utils.RespondLocalizedError(c, http.StatusNotFound, "config.not_found")
		return
	}
	h.addAuditLog(c, "toggle_reminder_config", "reminder_config", &id, "")
	utils.RespondSuccess(c, gin.H{"ok": true})
}

func (h *AdminHandler) AdminDeleteReminderConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "config.invalid_id")
		return
	}
	deleted, err := h.reminderConfigRepo.AdminDelete(c.Request.Context(), id)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "reminder_config.delete", err)
		return
	}
	if !deleted {
		utils.RespondLocalizedError(c, http.StatusNotFound, "config.not_found")
		return
	}
	h.addAuditLog(c, "delete_reminder_config", "reminder_config", &id, "")
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) ListAllReminderLogs(c *gin.Context) {
	page, limit := parsePaginationParams(c)

	var userID int64
	if raw := c.Query("user_id"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			userID = v
		}
	}
	status := c.Query("status")
	if status != "" && status != "success" && status != "failed" {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "system.invalid_status_filter")
		return
	}

	logs, total, err := h.reminderLogRepo.AdminListFiltered(c.Request.Context(), page, limit, userID, status)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "reminder_log.list", err)
		return
	}
	ids := make([]int64, 0, len(logs))
	for _, l := range logs {
		ids = append(ids, l.UserID)
	}
	usernames := h.lookupUsernames(c.Request.Context(), uniqueUserIDs(ids))
	utils.RespondPaginated(c, adminReminderLogsView(logs, usernames), page, limit, total)
}

type configFieldView struct {
	Key       string   `json:"key"`
	Label     string   `json:"label"`
	Type      string   `json:"type"`
	Enum      []string `json:"enum,omitempty"`
	Editable  bool     `json:"editable"`
	HotReload bool     `json:"hot_reload"`
	Value     any      `json:"value"`
	Source    string   `json:"source"` // cli | db | config
}

type configGroupView struct {
	Group  string            `json:"group"`
	Fields []configFieldView `json:"fields"`
}

func (h *AdminHandler) GetConfig(c *gin.Context) {
	dbValues, err := h.appConfigSvc.LoadAll(c.Request.Context())
	if err != nil {
		dbValues = map[string]string{}
	}
	utils.RespondSuccess(c, gin.H{"groups": h.buildConfigView(dbValues)})
}

// GetOAuthCallbackURLs 返回各 OAuth provider 的回调地址，供管理员复制到 GitHub/Google/LinuxDo 后台填写。
func (h *AdminHandler) GetOAuthCallbackURLs(c *gin.Context) {
	if !h.cfg.OAuth.Enabled {
		utils.RespondSuccess(c, gin.H{"base_url": "", "callbacks": []any{}})
		return
	}

	// 优先从请求中获取实际访问地址（处理反向代理场景）
	host := c.Request.Host

	// 检测 scheme：反向代理通常通过 X-Forwarded-Proto 传递
	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Request.TLS != nil {
		scheme = "https"
	} else if h.cfg.Server.Mode == "release" {
		scheme = "https"
	}

	base := fmt.Sprintf("%s://%s", scheme, host)

	// 从 OAuth registry 获取已启用的 provider 信息（含 icon）
	type callbackItem struct {
		Provider    string `json:"provider"`
		Label       string `json:"label"`
		Icon        string `json:"icon"`
		CallbackURL string `json:"callback_url"`
	}
	var callbacks []callbackItem
	icons := map[string]string{"github": "github", "google": "chrome", "linuxdo": "terminal"}
	for _, name := range []string{"github", "google", "linuxdo"} {
		callbacks = append(callbacks, callbackItem{
			Provider:    name,
			Label:       map[string]string{"github": "GitHub", "google": "Google", "linuxdo": "LinuxDo"}[name],
			Icon:        icons[name],
			CallbackURL: fmt.Sprintf("%s/api/v1/auth/oauth/%s/callback", base, name),
		})
	}

	utils.RespondSuccess(c, gin.H{"base_url": base, "callbacks": callbacks})
}

// buildConfigView 遍历注册表按分组聚合;有 DB 覆盖且未被 CLI 锁定时显示 DB 待生效值,密码脱敏。
func (h *AdminHandler) buildConfigView(dbValues map[string]string) []configGroupView {
	groups := []configGroupView{}
	idx := map[string]int{}
	for i := range config.Registry {
		spec := &config.Registry[i]
		value := spec.Get(h.cfg)
		_, inDB := dbValues[spec.Key]
		if inDB && spec.Editable && !h.lockedKeys[spec.Key] {
			var dbv any
			if json.Unmarshal([]byte(dbValues[spec.Key]), &dbv) == nil {
				value = dbv
			}
		}
		if spec.Key == "admin.password" || spec.Key == "email.smtp_password" {
			value = "***"
		}
		source := "config"
		if h.lockedKeys[spec.Key] {
			source = "cli"
		} else if inDB && spec.Editable {
			source = "db"
		}
		field := configFieldView{
			Key:       spec.Key,
			Label:     spec.Label,
			Type:      string(spec.Type),
			Enum:      spec.Enum,
			Editable:  spec.Editable,
			HotReload: spec.HotReload,
			Value:     value,
			Source:    source,
		}
		gi, ok := idx[spec.Group]
		if !ok {
			groups = append(groups, configGroupView{Group: spec.Group})
			gi = len(groups) - 1
			idx[spec.Group] = gi
		}
		groups[gi].Fields = append(groups[gi].Fields, field)
	}
	return groups
}

type updateConfigRequest struct {
	Updates []struct {
		Key   string `json:"key" binding:"required"`
		Value any    `json:"value"`
	} `json:"updates" binding:"required"`
}

func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req updateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}
	adminID := c.GetInt64("user_id")
	byKey := config.RegistryByKey()
	updated := make([]string, 0, len(req.Updates))
	restartRequired := false
	for _, u := range req.Updates {
		if err := h.appConfigSvc.Set(c.Request.Context(), u.Key, u.Value, adminID); err != nil {
			switch {
			case errors.Is(err, service.ErrConfigKeyUnknown):
				utils.RespondError(c, http.StatusBadRequest, "未知配置项: "+u.Key, utils.CodeInvalidInput)
			case errors.Is(err, service.ErrConfigKeyLocked):
				utils.RespondError(c, http.StatusBadRequest, "配置项只读: "+u.Key, utils.CodeInvalidInput)
			case errors.Is(err, service.ErrConfigValueInvalid):
				utils.RespondError(c, http.StatusBadRequest, "配置值无效: "+u.Key, utils.CodeInvalidInput)
			default:
				utils.RespondInternalError(c, "更新配置失败", err)
			}
			return
		}
		updated = append(updated, u.Key)
		if spec := byKey[u.Key]; spec != nil && spec.HotReload {
			h.applyHotReload(u.Key, u.Value)
		} else {
			restartRequired = true
		}
	}
	if len(updated) > 0 {
		h.addAuditLog(c, "config.update", "app_config", nil, strings.Join(updated, ","))
	}
	utils.RespondSuccess(c, gin.H{"restart_required": restartRequired, "updated": updated})
}

func (h *AdminHandler) ResetConfig(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" {
		utils.RespondError(c, http.StatusBadRequest, "缺少配置 key", utils.CodeInvalidInput)
		return
	}
	adminID := c.GetInt64("user_id")
	if err := h.appConfigSvc.Reset(c.Request.Context(), key, adminID); err != nil {
		switch {
		case errors.Is(err, service.ErrConfigKeyUnknown):
			utils.RespondError(c, http.StatusBadRequest, "未知配置项: "+key, utils.CodeInvalidInput)
		case errors.Is(err, service.ErrConfigKeyLocked):
			utils.RespondError(c, http.StatusBadRequest, "配置项只读: "+key, utils.CodeInvalidInput)
		default:
			utils.RespondInternalError(c, "重置配置失败", err)
		}
		return
	}
	h.addAuditLog(c, "config.reset", "app_config", nil, key)
	utils.RespondSuccess(c, gin.H{"restart_required": true})
}

// TestEmail 测试 SMTP 邮箱连接。
// 注意：h.cfg 的字段在并发请求下可能存在短暂竞态（与 reminder.enabled 等同），
// 但 admin 端点仅限 localhost 低频访问，可接受。
func (h *AdminHandler) TestEmail(c *gin.Context) {
	if h.emailSvc == nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.not_configured")
		return
	}
	if err := h.emailSvc.TestConnection(c.Request.Context()); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), "EMAIL_TEST_FAILED")
		return
	}
	h.addAuditLog(c, "config.test_email", "app_config", nil, "email")
	utils.RespondSuccess(c, gin.H{"ok": true, "message": "SMTP 连接测试成功"})
}

// applyHotReload 对热生效项立即应用到运行中的系统,并同步内存 cfg 供 GetConfig 显示。
func (h *AdminHandler) applyHotReload(key string, value any) {
	switch key {
	case "i18n.default_lang":
		if s, ok := value.(string); ok && h.applyI18n != nil {
			h.applyI18n(s)
			h.cfg.I18n.DefaultLang = s
		}
	case "reminder.enabled":
		if b, ok := value.(bool); ok && h.reminderSvc != nil {
			h.reminderSvc.SetEnabled(b)
			h.cfg.Reminder.Enabled = b
		}
	case "email.enabled":
		if b, ok := value.(bool); ok {
			h.cfg.Email.Enabled = b
			if h.emailSvc != nil {
				h.emailSvc.SetEnabled(b)
			}
		}
	case "email.smtp_host":
		if s, ok := value.(string); ok {
			h.cfg.Email.SMTPHost = s
		}
	case "email.smtp_port":
		if n, ok := value.(float64); ok {
			h.cfg.Email.SMTPPort = int(n)
		}
	case "email.smtp_username":
		if s, ok := value.(string); ok {
			h.cfg.Email.SMTPUsername = s
		}
	case "email.smtp_password":
		if s, ok := value.(string); ok {
			h.cfg.Email.SMTPPassword = s
		}
	case "email.from_address":
		if s, ok := value.(string); ok {
			h.cfg.Email.FromAddress = s
		}
	case "email.from_name":
		if s, ok := value.(string); ok {
			h.cfg.Email.FromName = s
		}
	case "server.timezone":
		if s, ok := value.(string); ok {
			h.cfg.Server.Timezone = s
			loc, err := utils.ResolveTimezone(s)
			if err == nil {
				timezone.Init(loc)
			}
		}
	case "reminder.worker_count":
		if n, ok := value.(float64); ok && h.reminderSvc != nil {
			h.cfg.Reminder.WorkerCount = int(n)
			h.reminderSvc.SetWorkerCount(int(n))
		}
	case "reminder.grace_period_minutes":
		if n, ok := value.(float64); ok {
			h.cfg.Reminder.GracePeriodMinutes = int(n)
			if setter, ok := h.taskRepo.(gracePeriodSetter); ok {
				setter.SetGracePeriod(time.Duration(int(n)) * time.Minute)
			}
		}
	case "reminder.webhook_body_template":
		if s, ok := value.(string); ok && h.reminderSvc != nil {
			h.cfg.Reminder.WebhookBodyTemplate = s
			h.reminderSvc.UpdateTemplate(s)
		}
	// —— OAuth 热更新 ——
	case "oauth.enabled":
		if b, ok := value.(bool); ok {
			h.cfg.OAuth.Enabled = b
		}
	case "oauth.github.enabled":
		if b, ok := value.(bool); ok {
			h.cfg.OAuth.GitHub.Enabled = b
		}
	case "oauth.github.client_id":
		if s, ok := value.(string); ok {
			h.cfg.OAuth.GitHub.ClientID = s
		}
	case "oauth.github.client_secret":
		if s, ok := value.(string); ok {
			h.cfg.OAuth.GitHub.ClientSecret = s
		}
	case "oauth.google.enabled":
		if b, ok := value.(bool); ok {
			h.cfg.OAuth.Google.Enabled = b
		}
	case "oauth.google.client_id":
		if s, ok := value.(string); ok {
			h.cfg.OAuth.Google.ClientID = s
		}
	case "oauth.google.client_secret":
		if s, ok := value.(string); ok {
			h.cfg.OAuth.Google.ClientSecret = s
		}
	case "oauth.linuxdo.enabled":
		if b, ok := value.(bool); ok {
			h.cfg.OAuth.LinuxDo.Enabled = b
		}
	case "oauth.linuxdo.client_id":
		if s, ok := value.(string); ok {
			h.cfg.OAuth.LinuxDo.ClientID = s
		}
	case "oauth.linuxdo.client_secret":
		if s, ok := value.(string); ok {
			h.cfg.OAuth.LinuxDo.ClientSecret = s
		}
	}
	// 所有 OAuth key 变更后统一重建 provider 实例
	if strings.HasPrefix(key, "oauth.") && h.reloadOAuth != nil {
		h.reloadOAuth()
	}
}

// lookupUsernames returns a userID → username map for the given IDs in one query.
func (h *AdminHandler) lookupUsernames(ctx context.Context, ids []int64) map[int64]string {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "SELECT id, username FROM users WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	m := make(map[int64]string, len(ids))
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			m[id] = name
		}
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return m
}

// --- Admin response structs (carry username alongside the original data) ---

type adminTaskResponse struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	Username       string  `json:"username"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Completed      bool    `json:"completed"`
	Priority       int     `json:"priority"`
	DueAt          *string `json:"due_at"`
	RemindAt       *string `json:"remind_at"`
	RepeatType     string  `json:"repeat_type"`
	RepeatInterval int     `json:"repeat_interval"`
	RepeatEndDate  *string `json:"repeat_end_date"`
	ReminderSent   bool    `json:"reminder_sent"`
	ReminderSentAt *string `json:"reminder_sent_at"`
	FocusDuration  *int    `json:"focus_duration"`
	Tags           []string `json:"tags"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func adminTasksView(tasks []models.Task, usernames map[int64]string) []adminTaskResponse {
	out := make([]adminTaskResponse, len(tasks))
	for i, t := range tasks {
		out[i] = adminTaskResponse{
			ID: t.ID, UserID: t.UserID, Username: usernames[t.UserID],
			Title: t.Title, Description: t.Description, Completed: t.Completed,
			Priority: t.Priority, DueAt: t.DueAt, RemindAt: t.RemindAt,
			RepeatType: t.RepeatType, RepeatInterval: t.RepeatInterval,
			RepeatEndDate: t.RepeatEndDate, ReminderSent: t.ReminderSent,
			ReminderSentAt: t.ReminderSentAt, FocusDuration: t.FocusDuration,
			Tags: t.Tags, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		}
	}
	return out
}

func singleAdminTaskView(t *models.Task, usernames map[int64]string) adminTaskResponse {
	return adminTaskResponse{
		ID: t.ID, UserID: t.UserID, Username: usernames[t.UserID],
		Title: t.Title, Description: t.Description, Completed: t.Completed,
		Priority: t.Priority, DueAt: t.DueAt, RemindAt: t.RemindAt,
		RepeatType: t.RepeatType, RepeatInterval: t.RepeatInterval,
		RepeatEndDate: t.RepeatEndDate, ReminderSent: t.ReminderSent,
		ReminderSentAt: t.ReminderSentAt, FocusDuration: t.FocusDuration,
		Tags: t.Tags, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

type adminReminderConfigResponse struct {
	ID                  int64             `json:"id"`
	UserID              int64             `json:"user_id"`
	Username            string            `json:"username"`
	Name                string            `json:"name"`
	ChannelType         string            `json:"channel_type"`
	WebhookURL          string            `json:"webhook_url"`
	WebhookMethod       string            `json:"webhook_method"`
	WebhookHeaders      map[string]string `json:"webhook_headers"`
	WebhookBodyTemplate string            `json:"webhook_body_template"`
	MaxRetries          int               `json:"max_retries"`
	RetryDelaySeconds   int               `json:"retry_delay_seconds"`
	Enabled             bool              `json:"enabled"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
}

func adminReminderConfigsView(configs []models.UserReminderConfig, usernames map[int64]string) []adminReminderConfigResponse {
	out := make([]adminReminderConfigResponse, len(configs))
	for i, c := range configs {
		out[i] = adminReminderConfigResponse{
			ID: c.ID, UserID: c.UserID, Username: usernames[c.UserID],
			Name: c.Name, ChannelType: c.ChannelType, WebhookURL: c.WebhookURL,
			WebhookMethod: c.WebhookMethod, WebhookHeaders: c.WebhookHeaders,
			WebhookBodyTemplate: c.WebhookBodyTemplate, MaxRetries: c.MaxRetries,
			RetryDelaySeconds: c.RetryDelaySeconds, Enabled: c.Enabled,
			CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
		}
	}
	return out
}

type adminReminderLogResponse struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	Username         string `json:"username"`
	TaskID           int64  `json:"task_id"`
	TaskTitle        string `json:"task_title"`
	ReminderConfigID *int64 `json:"reminder_config_id"`
	ChannelName      string `json:"channel_name"`
	ChannelType      string `json:"channel_type"`
	Status           string `json:"status"`
	Attempts         int    `json:"attempts"`
	ErrorMessage     string `json:"error_message"`
	CreatedAt        string `json:"created_at"`
}

func adminReminderLogsView(logs []models.ReminderLog, usernames map[int64]string) []adminReminderLogResponse {
	out := make([]adminReminderLogResponse, len(logs))
	for i, l := range logs {
		out[i] = adminReminderLogResponse{
			ID: l.ID, UserID: l.UserID, Username: usernames[l.UserID],
			TaskID: l.TaskID, TaskTitle: l.TaskTitle, ReminderConfigID: l.ReminderConfigID,
			ChannelName: l.ChannelName, ChannelType: l.ChannelType,
			Status: l.Status, Attempts: l.Attempts, ErrorMessage: l.ErrorMessage,
			CreatedAt: l.CreatedAt,
		}
	}
	return out
}

func uniqueUserIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func parsePaginationParams(c *gin.Context) (page, limit int) {
	page = 1
	limit = 20
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 && v <= 100 {
		limit = v
	}
	return
}

func (h *AdminHandler) addAuditLog(c *gin.Context, action, targetType string, targetID *int64, detail string) {
	if h.auditLogRepo == nil {
		return
	}
	adminUserID := c.GetInt64("user_id")
	if adminUserID == 0 {
		return
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), adminUserID, action, targetType, targetID, detail); err != nil {
		logging.Logger(c.Request.Context(), nil).Sugar().Warnw("failed to write audit log",
			"action", action, "target_type", targetType, "error", err)
	}
}

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	page, limit := parsePaginationParams(c)
	logs, total, err := h.auditLogRepo.ListAll(c.Request.Context(), page, limit)
	if err != nil {
		utils.RespondInternalError(c, "failed to list audit logs", err)
		return
	}
	utils.RespondPaginated(c, logs, page, limit, total)
}
