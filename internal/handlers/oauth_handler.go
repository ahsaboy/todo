package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"todo/internal/config"
	"todo/internal/middleware"
	"todo/internal/oauth"
	"todo/internal/service"
	"todo/internal/timezone"
	"todo/internal/utils"
	"todo/internal/views"
)

const (
	oauthStateCookie = "oauth_state"
	oauthStateMaxAge = 600 // 10 分钟
)

// OAuthHandler 处理用户端和管理后台的 OAuth 流程。
type OAuthHandler struct {
	oauthSvc service.OAuthServiceInterface
	oauthReg *oauth.Registry
	cfg      *config.Config
}

func NewOAuthHandler(
	oauthSvc service.OAuthServiceInterface,
	oauthReg *oauth.Registry,
	cfg *config.Config,
) *OAuthHandler {
	return &OAuthHandler{
		oauthSvc: oauthSvc,
		oauthReg: oauthReg,
		cfg:      cfg,
	}
}

// GetProviders 返回已启用的 OAuth provider 列表。
func (h *OAuthHandler) GetProviders(c *gin.Context) {
	providers := h.oauthSvc.GetAvailableProviders()
	utils.RespondSuccess(c, providers)
}

// InitiateOAuth 发起用户端 OAuth 流程。
// 支持 ?redirect_uri=http://frontend-origin 参数，由前端传入当前页面 origin。
func (h *OAuthHandler) InitiateOAuth(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = h.cfg.OAuth.FrontendURL
	}
	h.initiate(c, "user", redirectURI)
}

// HandleCallback 处理用户端 OAuth 回调。
func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	h.callback(c)
}

// AdminInitiateOAuth 发起管理后台 OAuth 流程。
func (h *OAuthHandler) AdminInitiateOAuth(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = h.adminRedirectBase()
	}
	h.initiate(c, "admin", redirectURI)
}

// AdminHandleCallback 处理管理后台 OAuth 回调。
func (h *OAuthHandler) AdminHandleCallback(c *gin.Context) {
	h.callback(c)
}

func (h *OAuthHandler) initiate(c *gin.Context, scope, redirectURI string) {
	providerName := c.Param("provider")
	if providerName == "" {
		utils.RespondError(c, http.StatusBadRequest, "provider is required", utils.CodeInvalidInput)
		return
	}

	// 验证 redirect_uri 合法（必须是 http/https origin，不能是任意 URL）
	if !isValidOrigin(redirectURI) {
		utils.RespondError(c, http.StatusBadRequest, "invalid redirect_uri", utils.CodeInvalidInput)
		return
	}

	// 验证 provider 可用
	providers := h.oauthSvc.GetAvailableProviders()
	found := false
	for _, p := range providers {
		if p.Name == providerName {
			found = true
			break
		}
	}
	if !found {
		utils.RespondError(c, http.StatusBadRequest, "provider not available", utils.CodeInvalidInput)
		return
	}

	// 生成 state: scope|provider|timestamp|nonce|redirect_uri_b64[|userID]
	nonce := make([]byte, 16)
	rand.Read(nonce)
	ts := time.Now().Unix()
	redirectB64 := base64.RawURLEncoding.EncodeToString([]byte(redirectURI))
	payload := fmt.Sprintf("%s|%s|%d|%s|%s", scope, providerName, ts, hex.EncodeToString(nonce), redirectB64)
	if scope == "link" {
		userID, ok := middleware.GetUserID(c)
		if !ok {
			utils.RespondError(c, http.StatusUnauthorized, "auth.unauthorized", "")
			return
		}
		payload = fmt.Sprintf("%s|%d", payload, userID)
	}
	sig := h.signState(payload)

	// 写入 cookie（统一路径，callback 通过 scope 字段区分登录/绑定）
	secure := h.cfg.Server.Mode == "release"
	cookiePath := "/api/v1/auth/oauth"
	if scope == "admin" {
		cookiePath = "/admin/api/auth/oauth"
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, payload+"|"+sig, oauthStateMaxAge, cookiePath, "", secure, true)

	// 获取 OAuth2 配置并跳转
	provider, ok := h.oauthReg.Get(providerName)
	if !ok {
		utils.RespondInternalError(c, "provider not found", nil)
		return
	}
	oauth2Cfg := provider.OAuth2Config()
	if oauth2Cfg == nil {
		utils.RespondInternalError(c, "provider config error", nil)
		return
	}

	authURL := oauth2Cfg.AuthCodeURL(sig)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *OAuthHandler) callback(c *gin.Context) {
	providerName := c.Param("provider")
	code := c.Query("code")
	stateParam := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		h.redirectWithError(c, h.cfg.OAuth.FrontendURL, "user", errorParam)
		return
	}
	if code == "" {
		h.redirectWithError(c, h.cfg.OAuth.FrontendURL, "user", "missing_code")
		return
	}

	// 读取并清除 state cookie
	cookieVal, err := c.Cookie(oauthStateCookie)
	if err != nil || cookieVal == "" {
		h.redirectWithError(c, h.cfg.OAuth.FrontendURL, "user", "invalid_state")
		return
	}
	// 先读取 scope 以确定清除路径
	firstPipe := strings.Index(cookieVal, "|")
	if firstPipe < 0 {
		h.redirectWithError(c, h.cfg.OAuth.FrontendURL, "user", "invalid_state")
		return
	}
	cookieScope := cookieVal[:firstPipe]
	clearPath := "/api/v1/auth/oauth"
	if cookieScope == "admin" {
		clearPath = "/admin/api/auth/oauth"
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, "", -1, clearPath, "", false, true)

	if cookieScope == "link" {
		h.handleLinkCallback(c, providerName, code, stateParam, cookieVal)
		return
	}

	// ── 登录回调 ──
	// 解析 cookie: scope|provider|timestamp|nonce|redirect_uri_b64|signature
	parts := strings.SplitN(cookieVal, "|", 6)
	if len(parts) != 6 {
		h.redirectWithError(c, h.cfg.OAuth.FrontendURL, "user", "invalid_state")
		return
	}
	cookieProvider, tsStr, _, redirectB64, cookieSig := parts[1], parts[2], parts[3], parts[4], parts[5]

	// 解码 redirect_uri
	redirectBytes, err := base64.RawURLEncoding.DecodeString(redirectB64)
	if err != nil {
		h.redirectWithError(c, h.cfg.OAuth.FrontendURL, "user", "invalid_state")
		return
	}
	redirectURI := string(redirectBytes)

	// 验证 provider
	if cookieProvider != providerName {
		h.redirectWithError(c, redirectURI, "user", "provider_mismatch")
		return
	}
	// 验证 HMAC（签名覆盖前 5 个字段，不含 signature 本身）
	payload := strings.Join(parts[:5], "|")
	if !h.verifyState(payload, cookieSig) {
		h.redirectWithError(c, redirectURI, "user", "invalid_signature")
		return
	}
	// 验证过期
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Since(time.Unix(ts, 0)) > 10*time.Minute {
		h.redirectWithError(c, redirectURI, "user", "state_expired")
		return
	}
	// 验证 state 参数匹配
	if stateParam != cookieSig {
		h.redirectWithError(c, redirectURI, "user", "state_mismatch")
		return
	}

	// 调用 service
	_, apiKey, err := h.oauthSvc.HandleCallback(c.Request.Context(), providerName, code)
	if err != nil {
		h.redirectWithError(c, redirectURI, "user", "callback_failed")
		return
	}

	redirectURL := fmt.Sprintf("%s#/oauth/callback?key=%s", redirectURI, url.QueryEscape(apiKey))
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// handleLinkCallback 处理 OAuth 绑定回调（scope=link）。
func (h *OAuthHandler) handleLinkCallback(c *gin.Context, providerName, code, stateParam, cookieVal string) {
	frontendBase := h.cfg.OAuth.FrontendURL

	// 解析: link|provider|timestamp|nonce|redirect_uri_b64|userID|signature
	parts := strings.SplitN(cookieVal, "|", 8)
	if len(parts) != 7 {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=invalid_state", frontendBase))
		return
	}
	cookieProvider, tsStr, _, redirectB64, userIDStr, cookieSig := parts[1], parts[2], parts[3], parts[4], parts[5], parts[6]

	redirectBytes, err := base64.RawURLEncoding.DecodeString(redirectB64)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=invalid_state", frontendBase))
		return
	}
	redirectURI := string(redirectBytes)

	if cookieProvider != providerName {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=provider_mismatch", redirectURI))
		return
	}

	payload := strings.Join(parts[:6], "|")
	if !h.verifyState(payload, cookieSig) {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=invalid_signature", redirectURI))
		return
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Since(time.Unix(ts, 0)) > 10*time.Minute {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=state_expired", redirectURI))
		return
	}
	if stateParam != cookieSig {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=state_mismatch", redirectURI))
		return
	}

	// 解析 userID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=invalid_state", redirectURI))
		return
	}

	// 执行绑定
	if err := h.oauthSvc.LinkAccount(c.Request.Context(), userID, providerName, code); err != nil {
		errMsg := "link_failed"
		switch {
		case errors.Is(err, service.ErrOAuthAlreadyLinked):
			errMsg = "already_linked"
		case errors.Is(err, service.ErrOAuthLinkedToOther):
			errMsg = "linked_to_other"
		}
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_link_error=%s", redirectURI, errMsg))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#/profile?oauth_linked=%s", redirectURI, providerName))
}

func (h *OAuthHandler) adminRedirectBase() string {
	if h.cfg.OAuth.AdminURL != "" {
		return h.cfg.OAuth.AdminURL
	}
	return h.cfg.OAuth.FrontendURL
}

func (h *OAuthHandler) redirectWithError(c *gin.Context, base, scope, errMsg string) {
	loginPath := "/login"
	if scope == "admin" {
		loginPath = "/admin/login"
	}
	redirectURL := fmt.Sprintf("%s#%s?error=%s", base, loginPath, url.QueryEscape(errMsg))
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) signState(payload string) string {
	mac := hmac.New(sha256.New, []byte(h.cfg.OAuth.StateSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *OAuthHandler) verifyState(payload, signature string) bool {
	expected := h.signState(payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// isValidOrigin 验证是否是合法的 HTTP(S) origin（scheme + host，路径必须为空或 "/"）。
func isValidOrigin(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && (u.Path == "" || u.Path == "/")
}

// ── OAuth 账号管理（已认证用户） ──

// GetProfileAccounts 返回当前用户已绑定的 OAuth 账号列表。
func (h *OAuthHandler) GetProfileAccounts(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "auth.unauthorized", "")
		return
	}

	accounts, err := h.oauthSvc.ListUserAccounts(c.Request.Context(), userID)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "oauth.list", err)
		return
	}

	utils.RespondSuccess(c, views.OAuthAccountsView(accounts, timezone.Get()))
}

// UnlinkOAuthAccount 解除当前用户的指定 OAuth 绑定。
func (h *OAuthHandler) UnlinkOAuthAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "auth.unauthorized", "")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid id", utils.CodeInvalidInput)
		return
	}

	if err := h.oauthSvc.UnlinkAccount(c.Request.Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, service.ErrLastAuthMethod):
			utils.RespondError(c, http.StatusConflict, "oauth.unlink_last_method", utils.CodeInvalidInput)
		case errors.Is(err, service.ErrOAuthAccountNotFound):
			utils.RespondError(c, http.StatusNotFound, "oauth.account_not_found", utils.CodeNotFound)
		default:
			utils.RespondLocalizedInternalError(c, "oauth.unlink", err)
		}
		return
	}

	utils.RespondSuccess(c, nil)
}

// InitiateLinkOAuth 发起 OAuth 绑定流程（已认证用户将 OAuth 账号绑定到当前账户）。
func (h *OAuthHandler) InitiateLinkOAuth(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = h.cfg.OAuth.FrontendURL
	}
	h.initiate(c, "link", redirectURI)
}

