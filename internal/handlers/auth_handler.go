package handlers

import (
	"errors"
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

type AuthHandler struct {
	svc     service.AuthServiceInterface
	emailSvc service.EmailServiceInterface
}

func NewAuthHandler(svc service.AuthServiceInterface, emailSvc service.EmailServiceInterface) *AuthHandler {
	return &AuthHandler{svc: svc, emailSvc: emailSvc}
}

// Register 用户注册
// @Summary      用户注册
// @Description  创建新用户并返回 API Key
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body models.RegisterRequest true "注册信息"
// @Success      201  {object} utils.SuccessResponse{data=object{user=models.UserResponse, api_key=string}}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      409  {object} utils.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	// 邮箱服务启用时，强制要求邮箱和验证码
	if h.emailSvc != nil && h.emailSvc.IsEnabled() {
		if req.Email == "" {
			utils.RespondLocalizedError(c, http.StatusBadRequest, "email.required")
			return
		}
		if req.Code == "" {
			utils.RespondLocalizedError(c, http.StatusBadRequest, "email.code_required")
			return
		}
		if err := h.emailSvc.VerifyCode(c.Request.Context(), req.Email, req.Code, "register"); err != nil {
			h.handleCodeError(c, err)
			return
		}
	}

	user, apiKey, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			utils.RespondLocalizedError(c, http.StatusConflict, "auth.username_taken")
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyTaken) {
			utils.RespondLocalizedError(c, http.StatusConflict, "auth.email_taken")
			return
		}
		utils.RespondLocalizedInternalError(c, "auth.register", err)
		return
	}

	userView := views.UserResponseView(*user, timezone.Get())
	utils.RespondCreated(c, gin.H{
		"user":    userView,
		"api_key": apiKey,
	})
}

// Login 用户登录
// @Summary      用户登录
// @Description  验证用户名或邮箱密码并返回新的 API Key
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body models.LoginRequest true "登录信息"
// @Success      200  {object} utils.SuccessResponse{data=object{user=models.UserResponse, api_key=string}}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      401  {object} utils.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	user, apiKey, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "auth.invalid_credentials")
		return
	}

	userView := views.UserResponseView(*user, timezone.Get())
	utils.RespondSuccess(c, gin.H{
		"user":    userView,
		"api_key": apiKey,
	})
}

// GenerateAPIKey 生成新 API Key
// @Summary      生成新 API Key
// @Description  为当前用户生成一个新的 API Key
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        body body models.CreateKeyRequest true "Key 信息"
// @Success      200  {object} utils.SuccessResponse{data=models.APIKeyPlainResponse}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/keys [post]
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	apiKey, err := h.svc.GenerateAPIKey(c.Request.Context(), userID, req.Name)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "auth.generate_key", err)
		return
	}

	utils.RespondSuccess(c, models.APIKeyPlainResponse{
		APIKey: apiKey,
		Key:    apiKey,
	})
}

// RevokeAPIKey 撤销 API Key
// @Summary      撤销 API Key
// @Description  删除指定的 API Key
// @Tags         user
// @Produce      json
// @Param        id path int true "API Key ID"
// @Success      200  {object} map[string]interface{}
// @Failure      404  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/keys/{id} [delete]
func (h *AuthHandler) RevokeAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "apikey.invalid_id")
		return
	}

	deleted, err := h.svc.RevokeAPIKey(c.Request.Context(), id, userID)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "auth.revoke_key", err)
		return
	}
	if !deleted {
		utils.RespondLocalizedError(c, http.StatusNotFound, "apikey.not_found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "api key revoked"})
}

// ListAPIKeys 列出所有 API Keys
// @Summary      列出所有 API Keys
// @Description  获取当前用户的所有 API Key
// @Tags         user
// @Produce      json
// @Success      200  {object} utils.SuccessResponse{data=[]models.APIKeyInfo}
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/keys [get]
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	keys, err := h.svc.ListAPIKeys(c.Request.Context(), userID)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "auth.list_keys", err)
		return
	}

	infos := make([]models.APIKeyInfo, 0, len(keys))
	for _, k := range keys {
		infos = append(infos, models.APIKeyInfo{
			ID:         k.ID,
			Name:       k.Name,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		})
	}

	utils.RespondSuccess(c, views.APIKeyInfosView(infos, timezone.Get()))
}

// GetProfile 获取用户信息
// @Summary      获取用户信息
// @Description  获取当前登录用户的个人信息
// @Tags         user
// @Produce      json
// @Success      200  {object} utils.SuccessResponse{data=models.UserResponse}
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		utils.RespondLocalizedError(c, http.StatusNotFound, "user.not_found")
		return
	}

	utils.RespondSuccess(c, views.UserResponseView(user.ToResponse(), timezone.Get()))
}

// UpdateProfile 更新用户信息
// @Summary      更新用户信息
// @Description  更新当前用户的邮箱等信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        body body models.UpdateProfileRequest true "要更新的信息"
// @Success      200  {object} utils.SuccessResponse{data=models.UserResponse}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if req.Email == nil {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "profile.no_fields")
		return
	}

	// 邮件服务启用时，更换邮箱需要验证码
	if h.emailSvc != nil && h.emailSvc.IsEnabled() {
		if req.Code == nil || *req.Code == "" {
			utils.RespondLocalizedError(c, http.StatusBadRequest, "profile.email_code_required")
			return
		}
		if err := h.emailSvc.VerifyCode(c.Request.Context(), *req.Email, *req.Code, "change_email"); err != nil {
			h.handleCodeError(c, err)
			return
		}
		// 检查邮箱是否已被其他用户使用
		existing, _ := h.svc.GetUserByEmail(c.Request.Context(), *req.Email)
		if existing != nil && existing.ID != userID {
			utils.RespondLocalizedError(c, http.StatusConflict, "profile.email_taken")
			return
		}
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), userID, *req.Email); err != nil {
		utils.RespondLocalizedInternalError(c, "user.update_profile", err)
		return
	}

	user, _ := h.svc.GetUserByID(c.Request.Context(), userID)
	utils.RespondSuccess(c, views.UserResponseView(user.ToResponse(), timezone.Get()))
}

// ChangePassword 修改密码
// @Summary      修改密码
// @Description  验证旧密码后设置新密码
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        body body models.ChangePasswordRequest true "密码信息"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      401  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidOldPassword) {
			utils.RespondLocalizedError(c, http.StatusUnauthorized, "auth.invalid_old_password")
			return
		}
		utils.RespondLocalizedInternalError(c, "user.change_password", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password changed"})
}

// SetPassword 为无密码用户（OAuth 用户）设置初始密码
func (h *AuthHandler) SetPassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondLocalizedError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	hasPwd, err := h.svc.HasPassword(c.Request.Context(), userID)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "user.change_password", err)
		return
	}
	if hasPwd {
		utils.RespondLocalizedError(c, http.StatusConflict, "profile.password_already_set")
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), userID, req.NewPassword); err != nil {
		utils.RespondLocalizedInternalError(c, "user.change_password", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password set"})
}

// EmailStatus 返回邮箱服务是否可用
func (h *AuthHandler) EmailStatus(c *gin.Context) {
	available := h.emailSvc != nil && h.emailSvc.IsEnabled()
	utils.RespondSuccess(c, gin.H{"available": available})
}

// SendCode 发送邮箱验证码
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req models.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if h.emailSvc == nil || !h.emailSvc.IsEnabled() {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.not_configured")
		return
	}

	// 注册时检查邮箱是否已被占用
	if req.Purpose == "register" {
		existing, err := h.svc.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			utils.RespondLocalizedInternalError(c, "email.send_code", err)
			return
		}
		if existing != nil {
			utils.RespondLocalizedError(c, http.StatusConflict, "auth.email_taken")
			return
		}
	}

	if err := h.emailSvc.SendVerificationCode(c.Request.Context(), req.Email, req.Purpose); err != nil {
		if errors.Is(err, service.ErrCodeCooldown) {
			utils.RespondLocalizedError(c, http.StatusTooManyRequests, "email.code_cooldown")
			return
		}
		if errors.Is(err, service.ErrEmailNotConfigured) {
			utils.RespondLocalizedError(c, http.StatusBadRequest, "email.not_configured")
			return
		}
		utils.RespondLocalizedInternalError(c, "email.send_code", err)
		return
	}

	utils.RespondSuccess(c, gin.H{"sent": true})
}

// VerifyCode 验证邮箱验证码
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req models.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if h.emailSvc == nil || !h.emailSvc.IsEnabled() {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.not_configured")
		return
	}

	if err := h.emailSvc.VerifyCode(c.Request.Context(), req.Email, req.Code, req.Purpose); err != nil {
		h.handleCodeError(c, err)
		return
	}

	utils.RespondSuccess(c, gin.H{"verified": true})
}

// ResetPassword 通过邮箱验证码重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if h.emailSvc == nil || !h.emailSvc.IsEnabled() {
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.not_configured")
		return
	}

	// 验证码校验
	if err := h.emailSvc.VerifyCode(c.Request.Context(), req.Email, req.Code, "reset_password"); err != nil {
		h.handleCodeError(c, err)
		return
	}

	// 查找用户（防止枚举，这里查不到也返回成功）
	user, err := h.svc.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "user.reset_password", err)
		return
	}
	if user == nil {
		utils.RespondSuccess(c, gin.H{"reset": true})
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), user.ID, req.Password); err != nil {
		utils.RespondLocalizedInternalError(c, "user.reset_password", err)
		return
	}

	utils.RespondSuccess(c, gin.H{"reset": true})
}

// handleCodeError 将验证码相关错误映射为本地化 HTTP 响应
func (h *AuthHandler) handleCodeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCodeNotFound):
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.code_not_found")
	case errors.Is(err, service.ErrCodeExpired):
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.code_expired")
	case errors.Is(err, service.ErrCodeAttemptsExceeded):
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.code_attempts_exceeded")
	case errors.Is(err, service.ErrCodeInvalid):
		utils.RespondLocalizedError(c, http.StatusBadRequest, "email.code_invalid")
	default:
		utils.RespondLocalizedInternalError(c, "email.verify_code", err)
	}
}
