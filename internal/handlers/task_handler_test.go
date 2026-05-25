package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/testutil"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// injectUserID 是测试辅助中间件，向 context 注入固定 user_id。
func injectUserID(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// ---- helpers ----

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

func assertStatus(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Errorf("status: want %d, got %d", want, got)
	}
}

func assertJSONField(t *testing.T, body []byte, field string, want any) {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, ok := m[field]
	if !ok {
		t.Errorf("field %q not found in response", field)
		return
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("field %q: want %v, got %v", field, want, got)
	}
}

// ---- TaskHandler tests ----

func newTaskRouter(svc service.TaskServiceInterface) *gin.Engine {
	r := gin.New()
	h := handlers.NewTaskHandler(svc)
	r.Use(injectUserID(1))
	r.POST("/tasks", h.Create)
	r.GET("/tasks/:id", h.GetByID)
	r.GET("/tasks", h.List)
	r.PUT("/tasks/:id", h.Update)
	r.DELETE("/tasks/:id", h.Delete)
	r.PATCH("/tasks/:id/complete", h.ToggleComplete)
	return r
}

func TestTaskHandler_Create_Success(t *testing.T) {
	svc := &testutil.MockTaskService{
		CreateFn: func(_ context.Context, userID int64, req models.CreateTaskRequest) (*models.Task, error) {
			return &models.Task{ID: 1, UserID: userID, Title: req.Title}, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"title": "buy milk"})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusCreated, w.Code)
	assertJSONField(t, w.Body.Bytes(), "success", true)
}

func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	svc := &testutil.MockTaskService{}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Create_ReminderChannelMissing(t *testing.T) {
	svc := &testutil.MockTaskService{
		CreateFn: func(_ context.Context, _ int64, _ models.CreateTaskRequest) (*models.Task, error) {
			return nil, service.ErrReminderChannelMissing
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"title": "remind", "remind_at": "2099-01-01T00:00:00Z"})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_GetByID_Success(t *testing.T) {
	svc := &testutil.MockTaskService{
		GetByIDFn: func(_ context.Context, _, id int64) (*models.Task, error) {
			return &models.Task{ID: id, UserID: 1, Title: "my task"}, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tasks/1", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusOK, w.Code)
	assertJSONField(t, w.Body.Bytes(), "success", true)
}

func TestTaskHandler_GetByID_NotFound(t *testing.T) {
	svc := &testutil.MockTaskService{
		GetByIDFn: func(_ context.Context, _, _ int64) (*models.Task, error) {
			return nil, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tasks/99", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusNotFound, w.Code)
}

func TestTaskHandler_GetByID_InvalidID(t *testing.T) {
	svc := &testutil.MockTaskService{}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tasks/abc", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Delete_NotFound(t *testing.T) {
	svc := &testutil.MockTaskService{
		DeleteFn: func(_ context.Context, _, _ int64) (bool, error) {
			return false, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/tasks/99", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusNotFound, w.Code)
}

func TestTaskHandler_Delete_Success(t *testing.T) {
	svc := &testutil.MockTaskService{
		DeleteFn: func(_ context.Context, _, _ int64) (bool, error) {
			return true, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/tasks/1", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusOK, w.Code)
}

func TestTaskHandler_ToggleComplete_NotFound(t *testing.T) {
	svc := &testutil.MockTaskService{
		ToggleCompleteFn: func(_ context.Context, _, _ int64, _ *int) (*models.Task, error) {
			return nil, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/tasks/99/complete", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusNotFound, w.Code)
}

func TestTaskHandler_ToggleComplete_Success(t *testing.T) {
	svc := &testutil.MockTaskService{
		ToggleCompleteFn: func(_ context.Context, _, id int64, _ *int) (*models.Task, error) {
			return &models.Task{ID: id, UserID: 1, Completed: true}, nil
		},
	}
	r := newTaskRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, "/tasks/1/complete", nil)
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusOK, w.Code)
}

func TestTaskHandler_NoUserID_Unauthorized(t *testing.T) {
	svc := &testutil.MockTaskService{}

	r := gin.New()
	h := handlers.NewTaskHandler(svc)
	r.POST("/tasks", h.Create) // no user_id middleware

	w := httptest.NewRecorder()
	body := toJSON(t, map[string]any{"title": "buy milk"})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assertStatus(t, http.StatusUnauthorized, w.Code)
}
