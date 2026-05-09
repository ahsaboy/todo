package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/utils"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
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

	user, apiKey, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			utils.RespondError(c, http.StatusConflict, "username already taken", utils.CodeInvalidInput)
			return
		}
		utils.RespondError(c, http.StatusInternalServerError, "failed to register", utils.CodeInternalError)
		return
	}

	utils.RespondCreated(c, gin.H{
		"user":    user,
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
		utils.RespondError(c, http.StatusUnauthorized, "invalid username or password", utils.CodeUnauthorized)
		return
	}

	utils.RespondSuccess(c, gin.H{
		"user":    user,
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
// @Success      200  {object} utils.SuccessResponse{data=models.APIKeyResponse}
// @Failure      400  {object} utils.ErrorResponse
// @Failure      500  {object} utils.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/v1/user/keys [post]
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	var req models.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	apiKey, err := h.svc.GenerateAPIKey(c.Request.Context(), userID, req.Name)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to generate key", utils.CodeInternalError)
		return
	}

	utils.RespondSuccess(c, gin.H{"api_key": apiKey})
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
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid key id", utils.CodeInvalidInput)
		return
	}

	deleted, err := h.svc.RevokeAPIKey(c.Request.Context(), id, userID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to revoke key", utils.CodeInternalError)
		return
	}
	if !deleted {
		utils.RespondError(c, http.StatusNotFound, "key not found", utils.CodeNotFound)
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
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	keys, err := h.svc.ListAPIKeys(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to list keys", utils.CodeInternalError)
		return
	}

	var infos []models.APIKeyInfo
	for _, k := range keys {
		infos = append(infos, models.APIKeyInfo{
			ID:         k.ID,
			Name:       k.Name,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		})
	}

	utils.RespondSuccess(c, infos)
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
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		utils.RespondError(c, http.StatusNotFound, "user not found", utils.CodeNotFound)
		return
	}

	utils.RespondSuccess(c, user.ToResponse())
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
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if req.Email == nil {
		utils.RespondError(c, http.StatusBadRequest, "no fields to update", utils.CodeInvalidInput)
		return
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), userID, *req.Email); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update profile", utils.CodeInternalError)
		return
	}

	user, _ := h.svc.GetUserByID(c.Request.Context(), userID)
	utils.RespondSuccess(c, user.ToResponse())
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
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized", utils.CodeUnauthorized)
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error(), utils.CodeInvalidInput)
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "invalid old password" {
			utils.RespondError(c, http.StatusUnauthorized, "invalid old password", utils.CodeUnauthorized)
			return
		}
		utils.RespondError(c, http.StatusInternalServerError, "failed to change password", utils.CodeInternalError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password changed"})
}
