package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"todo/internal/middleware"
	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

type ProfileHandler struct {
	authSvc  service.AuthServiceInterface
	oauthSvc service.OAuthServiceInterface // 可为 nil（OAuth 未启用时）
}

func NewProfileHandler(authSvc service.AuthServiceInterface, oauthSvc service.OAuthServiceInterface) *ProfileHandler {
	return &ProfileHandler{authSvc: authSvc, oauthSvc: oauthSvc}
}

// GetFullProfile 返回用户完整资料（含 OAuth 绑定和 hasPassword）。
func (h *ProfileHandler) GetFullProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "auth.unauthorized", "")
		return
	}

	user, err := h.authSvc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "profile.get", err)
		return
	}
	if user == nil {
		utils.RespondError(c, http.StatusNotFound, "profile.not_found", "")
		return
	}

	loc := timezone.Get()
	resp := models.ProfileResponse{
		User:          views.UserResponseView(user.ToResponse(), loc),
		OAuthAccounts: []models.OAuthAccountResponse{},
		HasPassword:   user.PasswordHash != "",
	}

	if h.oauthSvc != nil {
		accounts, err := h.oauthSvc.ListUserAccounts(c.Request.Context(), userID)
		if err != nil {
			utils.RespondLocalizedInternalError(c, "profile.get", err)
			return
		}
		resp.OAuthAccounts = views.OAuthAccountsView(accounts, loc)
	}

	utils.RespondSuccess(c, resp)
}
