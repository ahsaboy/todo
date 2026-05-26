package models

import "regexp"

// UserTag 用户的标签字典项,per-user 隔离。
type UserTag struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateTagRequest struct {
	Name      string  `json:"name" binding:"required,min=1,max=32"`
	Color     *string `json:"color" binding:"omitempty"`
	Icon      *string `json:"icon" binding:"omitempty"`
	SortOrder *int    `json:"sort_order"`
}

type UpdateTagRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=32"`
	Color     *string `json:"color" binding:"omitempty"`
	Icon      *string `json:"icon" binding:"omitempty"`
	SortOrder *int    `json:"sort_order"`
}

const DefaultTagColor = "#3b82f6"

var hexColorRegexp = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// IsValidHexColor 校验是否是 #RRGGBB 形式。
func IsValidHexColor(s string) bool {
	return hexColorRegexp.MatchString(s)
}

// CuratedIcons 是后端持有的图标白名单,前端通过 shared/icons/curated.ts 共享相同的 key 集合。
// 仅 ~60 个精选 Lucide 图标 key,空字符串表示"无图标"。
var CuratedIcons = map[string]struct{}{
	// 工作 / 项目
	"briefcase": {}, "building": {}, "building2": {}, "factory": {}, "presentation": {},
	"file-text": {}, "folder": {}, "clipboard-list": {}, "calendar": {}, "calendar-clock": {},
	// 学习 / 阅读
	"book": {}, "book-open": {}, "graduation-cap": {}, "library": {}, "pencil": {},
	"notebook": {}, "lightbulb": {}, "brain": {},
	// 生活 / 家庭
	"home": {}, "heart": {}, "users": {}, "user": {}, "baby": {},
	"shopping-cart": {}, "shopping-bag": {}, "utensils": {}, "coffee": {}, "pizza": {},
	// 健康 / 运动
	"dumbbell": {}, "bike": {}, "activity": {}, "pill": {}, "stethoscope": {},
	// 旅行 / 外出
	"plane": {}, "car": {}, "map": {}, "map-pin": {}, "compass": {},
	// 状态 / 标记
	"star": {}, "flag": {}, "bookmark": {}, "alert-triangle": {}, "alert-circle": {},
	"check-circle": {}, "clock": {}, "zap": {}, "flame": {}, "target": {},
	// 创作 / 技术
	"code": {}, "terminal": {}, "git-branch": {}, "bug": {}, "wrench": {},
	"palette": {}, "music": {}, "camera": {}, "film": {}, "gamepad-2": {},
	// 其他
	"gift": {}, "tag": {}, "hash": {}, "package": {},
}

// IsValidIcon 空字符串表示"无图标",合法;否则必须在精选词典内。
func IsValidIcon(icon string) bool {
	if icon == "" {
		return true
	}
	_, ok := CuratedIcons[icon]
	return ok
}
