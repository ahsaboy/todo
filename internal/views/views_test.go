package views

import (
	"testing"
	"time"

	"todo/internal/models"
)

func mustShanghai(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	return loc
}

func strPtr(s string) *string { return &s }

func TestTaskView_ConvertsAllTimeFields(t *testing.T) {
	loc := mustShanghai(t)
	due := "2026-05-10T10:30:00Z"
	remind := "2026-05-10T11:00:00Z"
	repeatEnd := "2026-06-10T10:30:00Z"
	sent := "2026-05-10T11:01:00Z"

	in := &models.Task{
		ID:             7,
		Title:          "demo",
		DueAt:          &due,
		RemindAt:       &remind,
		RepeatEndDate:  &repeatEnd,
		ReminderSentAt: &sent,
		CreatedAt:      "2026-05-09T08:00:00Z",
		UpdatedAt:      "2026-05-09T09:00:00Z",
	}
	out := TaskView(in, loc)
	if out == nil {
		t.Fatal("expected non-nil")
	}
	wantPtr := func(s string) string { return s }
	cases := []struct {
		name, got, want string
	}{
		{"DueAt", *out.DueAt, wantPtr("2026-05-10T18:30:00+08:00")},
		{"RemindAt", *out.RemindAt, wantPtr("2026-05-10T19:00:00+08:00")},
		{"RepeatEndDate", *out.RepeatEndDate, wantPtr("2026-06-10T18:30:00+08:00")},
		{"ReminderSentAt", *out.ReminderSentAt, wantPtr("2026-05-10T19:01:00+08:00")},
		{"CreatedAt", out.CreatedAt, "2026-05-09T16:00:00+08:00"},
		{"UpdatedAt", out.UpdatedAt, "2026-05-09T17:00:00+08:00"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}

	// 原对象未被修改
	if in.CreatedAt != "2026-05-09T08:00:00Z" {
		t.Errorf("original CreatedAt mutated: %q", in.CreatedAt)
	}
	if *in.DueAt != "2026-05-10T10:30:00Z" {
		t.Errorf("original DueAt mutated: %q", *in.DueAt)
	}
	// 新指针(不共享内存)
	if in.DueAt == out.DueAt {
		t.Error("expected new pointer for DueAt")
	}
}

func TestTaskView_Nil(t *testing.T) {
	if got := TaskView(nil, time.UTC); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestTaskView_EmptyPointerStringRoundTrip(t *testing.T) {
	loc := mustShanghai(t)
	empty := ""
	in := &models.Task{ID: 1, DueAt: &empty}
	out := TaskView(in, loc)
	if out.DueAt == nil {
		t.Fatal("DueAt should remain non-nil")
	}
	if *out.DueAt != "" {
		t.Errorf("DueAt = %q, want empty", *out.DueAt)
	}
}

func TestTasksView_NilReturnsEmptySlice(t *testing.T) {
	got := TasksView(nil, time.UTC)
	if got == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestTasksView_Batch(t *testing.T) {
	loc := mustShanghai(t)
	ts := []models.Task{
		{ID: 1, CreatedAt: "2026-05-10T10:00:00Z"},
		{ID: 2, CreatedAt: "2026-05-11T10:00:00Z"},
	}
	out := TasksView(ts, loc)
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].CreatedAt != "2026-05-10T18:00:00+08:00" {
		t.Errorf("out[0].CreatedAt = %q", out[0].CreatedAt)
	}
	if out[1].CreatedAt != "2026-05-11T18:00:00+08:00" {
		t.Errorf("out[1].CreatedAt = %q", out[1].CreatedAt)
	}
	// 原 slice 未被修改
	if ts[0].CreatedAt != "2026-05-10T10:00:00Z" {
		t.Errorf("original mutated: %q", ts[0].CreatedAt)
	}
}

func TestUserView_AndUserResponseView(t *testing.T) {
	loc := mustShanghai(t)

	u := &models.User{ID: 1, CreatedAt: "2026-05-10T10:30:00Z", UpdatedAt: "2026-05-10T11:30:00Z"}
	uo := UserView(u, loc)
	if uo.CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("User.CreatedAt = %q", uo.CreatedAt)
	}
	if uo.UpdatedAt != "2026-05-10T19:30:00+08:00" {
		t.Errorf("User.UpdatedAt = %q", uo.UpdatedAt)
	}

	if UserView(nil, loc) != nil {
		t.Error("UserView(nil) should return nil")
	}

	r := models.UserResponse{ID: 2, CreatedAt: "2026-05-10T10:30:00Z"}
	ro := UserResponseView(r, loc)
	if ro.CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("UserResponse.CreatedAt = %q", ro.CreatedAt)
	}
	// 原对象未变(值类型,本身就是拷贝;断言以防意外)
	if r.CreatedAt != "2026-05-10T10:30:00Z" {
		t.Errorf("original UserResponse mutated: %q", r.CreatedAt)
	}
}

func TestAPIKeyView_AndAPIKeyInfoView(t *testing.T) {
	loc := mustShanghai(t)
	last := "2026-05-10T10:30:00Z"
	k := &models.APIKey{ID: 1, LastUsedAt: &last, CreatedAt: "2026-05-09T08:00:00Z"}
	ko := APIKeyView(k, loc)
	if *ko.LastUsedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("APIKey.LastUsedAt = %q", *ko.LastUsedAt)
	}
	if ko.CreatedAt != "2026-05-09T16:00:00+08:00" {
		t.Errorf("APIKey.CreatedAt = %q", ko.CreatedAt)
	}
	if APIKeyView(nil, loc) != nil {
		t.Error("APIKeyView(nil) should return nil")
	}

	info := models.APIKeyInfo{ID: 1, LastUsedAt: strPtr("2026-05-10T10:30:00Z"), CreatedAt: "2026-05-09T08:00:00Z"}
	io := APIKeyInfoView(info, loc)
	if *io.LastUsedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("APIKeyInfo.LastUsedAt = %q", *io.LastUsedAt)
	}
	if io.CreatedAt != "2026-05-09T16:00:00+08:00" {
		t.Errorf("APIKeyInfo.CreatedAt = %q", io.CreatedAt)
	}

	infos := []models.APIKeyInfo{info}
	outs := APIKeyInfosView(infos, loc)
	if len(outs) != 1 || outs[0].CreatedAt != "2026-05-09T16:00:00+08:00" {
		t.Errorf("APIKeyInfosView batch failed: %+v", outs)
	}
	// nil → 空切片
	if got := APIKeyInfosView(nil, loc); got == nil || len(got) != 0 {
		t.Errorf("APIKeyInfosView(nil) should return empty slice, got %+v", got)
	}
}

func TestUserReminderConfigView(t *testing.T) {
	loc := mustShanghai(t)
	c := &models.UserReminderConfig{ID: 1, CreatedAt: "2026-05-10T10:30:00Z", UpdatedAt: "2026-05-10T11:30:00Z"}
	co := UserReminderConfigView(c, loc)
	if co.CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("CreatedAt = %q", co.CreatedAt)
	}
	if co.UpdatedAt != "2026-05-10T19:30:00+08:00" {
		t.Errorf("UpdatedAt = %q", co.UpdatedAt)
	}
	if UserReminderConfigView(nil, loc) != nil {
		t.Error("nil should return nil")
	}

	cs := []models.UserReminderConfig{*c}
	cos := UserReminderConfigsView(cs, loc)
	if len(cos) != 1 || cos[0].CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("batch failed: %+v", cos)
	}
	if got := UserReminderConfigsView(nil, loc); got == nil || len(got) != 0 {
		t.Errorf("nil should return empty slice, got %+v", got)
	}
}

func TestReminderLogView(t *testing.T) {
	loc := mustShanghai(t)
	l := models.ReminderLog{ID: 1, CreatedAt: "2026-05-10T10:30:00Z"}
	lo := ReminderLogView(l, loc)
	if lo.CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("CreatedAt = %q", lo.CreatedAt)
	}

	ls := []models.ReminderLog{l}
	los := ReminderLogsView(ls, loc)
	if len(los) != 1 || los[0].CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("batch failed: %+v", los)
	}
	if got := ReminderLogsView(nil, loc); got == nil || len(got) != 0 {
		t.Errorf("nil should return empty slice, got %+v", got)
	}
}

func TestTaskView_LegacyDBFormatCompatibility(t *testing.T) {
	loc := mustShanghai(t)
	in := &models.Task{ID: 1, CreatedAt: "2026-05-10 10:30:00"}
	out := TaskView(in, loc)
	if out.CreatedAt != "2026-05-10T18:30:00+08:00" {
		t.Errorf("legacy format conversion failed: %q", out.CreatedAt)
	}
}
