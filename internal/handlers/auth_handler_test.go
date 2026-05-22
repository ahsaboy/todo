package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/testutil"
)

func newAuthRouter(svc service.AuthServiceInterface) *gin.Engine {
	r := gin.New()
	h := handlers.NewAuthHandler(svc)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.Use(injectUserID(1))
	r.PUT("/user/password", h.ChangePassword)
	r.GET("/user/profile", h.GetProfile)
	return r
}

func TestAuthHandler_Register_Success(t *testing.T) {
	svc := &testutil.MockAuthService{
		RegisterFn: func(_ context.Context, req models.RegisterRequest) (*models.UserResponse, string, error) {
			return &models.UserResponse{ID: 1, Username: req.Username}, "test-api-key", nil
		},
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"username": "alice", "password": "password123"})
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusCreated, w.Code)
	assertJSONField(t, w.Body.Bytes(), "success", true)
}

func TestAuthHandler_Register_UsernameTaken(t *testing.T) {
	svc := &testutil.MockAuthService{
		RegisterFn: func(_ context.Context, _ models.RegisterRequest) (*models.UserResponse, string, error) {
			return nil, "", service.ErrUsernameTaken
		},
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"username": "alice", "password": "password123"})
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusConflict, w.Code)
}

func TestAuthHandler_Register_InvalidInput(t *testing.T) {
	svc := &testutil.MockAuthService{}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	// password too short (< 6 chars)
	body := toJSON(t, map[string]any{"username": "ab", "password": "x"})
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := &testutil.MockAuthService{
		LoginFn: func(_ context.Context, req models.LoginRequest) (*models.UserResponse, string, error) {
			return &models.UserResponse{ID: 1, Username: "alice"}, "api-key", nil
		},
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"account": "alice", "password": "pass"})
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusOK, w.Code)
	assertJSONField(t, w.Body.Bytes(), "success", true)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &testutil.MockAuthService{
		LoginFn: func(_ context.Context, _ models.LoginRequest) (*models.UserResponse, string, error) {
			return nil, "", service.ErrInvalidCredentials
		},
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"account": "ghost", "password": "wrong"})
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Login_InvalidInput(t *testing.T) {
	svc := &testutil.MockAuthService{}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	// missing required "account" field
	body := toJSON(t, map[string]any{"password": "pass"})
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	svc := &testutil.MockAuthService{
		ChangePasswordFn: func(_ context.Context, _ int64, _, _ string) error { return nil },
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"old_password": "old", "new_password": "newpass123"})
	req, _ := http.NewRequest(http.MethodPut, "/user/password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ChangePassword_WrongOld(t *testing.T) {
	svc := &testutil.MockAuthService{
		ChangePasswordFn: func(_ context.Context, _ int64, _, _ string) error {
			return service.ErrInvalidOldPassword
		},
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"old_password": "wrong", "new_password": "newpass123"})
	req, _ := http.NewRequest(http.MethodPut, "/user/password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_ChangePassword_InternalError(t *testing.T) {
	svc := &testutil.MockAuthService{
		ChangePasswordFn: func(_ context.Context, _ int64, _, _ string) error {
			return errors.New("db error")
		},
	}
	r := newAuthRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"old_password": "old", "new_password": "newpass123"})
	req, _ := http.NewRequest(http.MethodPut, "/user/password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusInternalServerError, w.Code)
}
