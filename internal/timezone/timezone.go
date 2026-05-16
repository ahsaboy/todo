// Package timezone 提供进程级时区单例,供 view 函数和模板渲染使用。
//
// 用法:
//   - 启动期调用 Init(loc) 注入 utils.ResolveTimezone 的结果
//   - 任意位置调用 Get() 读取当前时区(永远返回非 nil)
//
// 设计要点:
//   - 用 atomic.Pointer 而非 mutex,Get() 无锁,零热路径开销
//   - 未 Init 时 Get() 兜底返回 time.Local,确保下游不需要 nil 检查
package timezone

import (
	"sync/atomic"
	"time"
)

var current atomic.Pointer[time.Location]

// Init 设置进程级时区。loc 为 nil 时回退到 time.Local。
// 应在 main 启动期完成调用一次;后续不可在并发请求中调用。
func Init(loc *time.Location) {
	if loc == nil {
		loc = time.Local
	}
	current.Store(loc)
}

// Get 返回当前进程时区,永远非 nil。
// 未 Init 时返回 time.Local。
func Get() *time.Location {
	if loc := current.Load(); loc != nil {
		return loc
	}
	return time.Local
}
