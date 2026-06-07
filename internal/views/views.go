// Package views 提供"DB 模型 → API 出参"的浅拷贝转换函数。
//
// 设计:
//   - 每个函数返回与输入相同的 model 类型(不是 *View 新类型),
//     保证 JSON 序列化的字段名/嵌套结构零变化,前端/Swagger 不受影响。
//   - 浅拷贝原对象后,只重写时间字段,通过 utils.FormatOutputTime 转成目标时区。
//   - pointer 字段用 utils.FormatOutputTimePtr 重新分配新指针,避免共享底层 string。
//   - view 函数为 pure function,只读不改原对象。
//
// 调用方应在 handler / MCP tool 出口处显式调用,把全局或 per-request 的 *time.Location 传进来。
package views

import (
	"time"

	"todo/internal/models"
	"todo/internal/utils"
)

// ---------- Task ----------

// TaskView 返回新的 *Task,所有时间字段已按 loc 重写。
// nil 输入返回 nil(便于 handler 不需要额外的 nil 检查)。
func TaskView(t *models.Task, loc *time.Location) *models.Task {
	if t == nil {
		return nil
	}
	out := *t
	out.DueAt = utils.FormatOutputTimePtr(t.DueAt, loc)
	out.RemindAt = utils.FormatOutputTimePtr(t.RemindAt, loc)
	out.RepeatEndDate = utils.FormatOutputTimePtr(t.RepeatEndDate, loc)
	out.ReminderSentAt = utils.FormatOutputTimePtr(t.ReminderSentAt, loc)
	out.CreatedAt = utils.FormatOutputTime(t.CreatedAt, loc)
	out.UpdatedAt = utils.FormatOutputTime(t.UpdatedAt, loc)
	return &out
}

// TasksView 批量转换 Task 列表。nil 输入返回 []Task{}(空切片,与 handler 现有契约保持一致)。
func TasksView(ts []models.Task, loc *time.Location) []models.Task {
	if ts == nil {
		return []models.Task{}
	}
	out := make([]models.Task, len(ts))
	for i := range ts {
		v := TaskView(&ts[i], loc)
		out[i] = *v
	}
	return out
}

// ---------- User ----------

// UserView 返回新的 *User,时间字段已按 loc 重写。nil 输入返回 nil。
func UserView(u *models.User, loc *time.Location) *models.User {
	if u == nil {
		return nil
	}
	out := *u
	out.CreatedAt = utils.FormatOutputTime(u.CreatedAt, loc)
	out.UpdatedAt = utils.FormatOutputTime(u.UpdatedAt, loc)
	return &out
}

// UserResponseView 返回新的 UserResponse(值类型),时间字段已按 loc 重写。
func UserResponseView(r models.UserResponse, loc *time.Location) models.UserResponse {
	r.CreatedAt = utils.FormatOutputTime(r.CreatedAt, loc)
	return r
}

// ---------- APIKey ----------

// APIKeyView 返回新的 *APIKey,时间字段已按 loc 重写。nil 输入返回 nil。
func APIKeyView(k *models.APIKey, loc *time.Location) *models.APIKey {
	if k == nil {
		return nil
	}
	out := *k
	out.LastUsedAt = utils.FormatOutputTimePtr(k.LastUsedAt, loc)
	out.CreatedAt = utils.FormatOutputTime(k.CreatedAt, loc)
	return &out
}

// APIKeyInfoView 返回新的 APIKeyInfo(值类型),时间字段已按 loc 重写。
func APIKeyInfoView(k models.APIKeyInfo, loc *time.Location) models.APIKeyInfo {
	k.LastUsedAt = utils.FormatOutputTimePtr(k.LastUsedAt, loc)
	k.CreatedAt = utils.FormatOutputTime(k.CreatedAt, loc)
	return k
}

// APIKeyInfosView 批量转换 APIKeyInfo 列表。nil 输入返回 []APIKeyInfo{}。
func APIKeyInfosView(ks []models.APIKeyInfo, loc *time.Location) []models.APIKeyInfo {
	if ks == nil {
		return []models.APIKeyInfo{}
	}
	out := make([]models.APIKeyInfo, len(ks))
	for i := range ks {
		out[i] = APIKeyInfoView(ks[i], loc)
	}
	return out
}

// ---------- UserReminderConfig ----------

// UserReminderConfigView 返回新的 *UserReminderConfig,时间字段已按 loc 重写。nil 输入返回 nil。
func UserReminderConfigView(c *models.UserReminderConfig, loc *time.Location) *models.UserReminderConfig {
	if c == nil {
		return nil
	}
	out := *c
	out.CreatedAt = utils.FormatOutputTime(c.CreatedAt, loc)
	out.UpdatedAt = utils.FormatOutputTime(c.UpdatedAt, loc)
	return &out
}

// UserReminderConfigsView 批量转换 UserReminderConfig 列表。nil 输入返回 []UserReminderConfig{}。
func UserReminderConfigsView(cs []models.UserReminderConfig, loc *time.Location) []models.UserReminderConfig {
	if cs == nil {
		return []models.UserReminderConfig{}
	}
	out := make([]models.UserReminderConfig, len(cs))
	for i := range cs {
		v := UserReminderConfigView(&cs[i], loc)
		out[i] = *v
	}
	return out
}

// ---------- ReminderLog ----------

// ReminderLogView 返回新的 ReminderLog(值类型),时间字段已按 loc 重写。
func ReminderLogView(l models.ReminderLog, loc *time.Location) models.ReminderLog {
	l.CreatedAt = utils.FormatOutputTime(l.CreatedAt, loc)
	return l
}

// ReminderLogsView 批量转换 ReminderLog 列表。nil 输入返回 []ReminderLog{}。
func ReminderLogsView(ls []models.ReminderLog, loc *time.Location) []models.ReminderLog {
	if ls == nil {
		return []models.ReminderLog{}
	}
	out := make([]models.ReminderLog, len(ls))
	for i := range ls {
		out[i] = ReminderLogView(ls[i], loc)
	}
	return out
}

// ---------- UserTag ----------

// UserTagView 返回新的 *UserTag,时间字段已按 loc 重写。nil 输入返回 nil。
func UserTagView(t *models.UserTag, loc *time.Location) *models.UserTag {
	if t == nil {
		return nil
	}
	out := *t
	out.CreatedAt = utils.FormatOutputTime(t.CreatedAt, loc)
	out.UpdatedAt = utils.FormatOutputTime(t.UpdatedAt, loc)
	return &out
}

// UserTagsView 批量转换 UserTag 列表。nil 输入返回 []UserTag{}。
func UserTagsView(ts []models.UserTag, loc *time.Location) []models.UserTag {
	if ts == nil {
		return []models.UserTag{}
	}
	out := make([]models.UserTag, len(ts))
	for i := range ts {
		v := UserTagView(&ts[i], loc)
		out[i] = *v
	}
	return out
}

// ---------- OAuthAccount ----------

// OAuthAccountView 返回新的 OAuthAccountResponse(值类型),时间字段已按 loc 重写。
func OAuthAccountView(a models.OAuthAccount, loc *time.Location) models.OAuthAccountResponse {
	resp := a.ToResponse()
	resp.LinkedAt = utils.FormatOutputTime(resp.LinkedAt, loc)
	return resp
}

// OAuthAccountsView 批量转换 OAuthAccount 列表。nil 输入返回 []OAuthAccountResponse{}。
func OAuthAccountsView(as []models.OAuthAccount, loc *time.Location) []models.OAuthAccountResponse {
	if as == nil {
		return []models.OAuthAccountResponse{}
	}
	out := make([]models.OAuthAccountResponse, len(as))
	for i := range as {
		out[i] = OAuthAccountView(as[i], loc)
	}
	return out
}
