package i18n

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// defaultLang 是默认语言，可通过 SetDefaultLang() 修改
var defaultLang = "zh-CN"

// SetDefaultLang 在启动时设置默认语言
func SetDefaultLang(lang string) {
	defaultLang = lang
}

// T 根据语言代码翻译消息键，支持 fmt.Sprintf 参数
// 查找顺序: 指定语言 -> 英文 -> key本身（永远不会返回空字符串）
func T(lang, key string, args ...interface{}) string {
	var tmpl string

	// 先查用户可见消息，再查内部消息
	if m, ok := messages[key]; ok {
		tmpl = m[lang]
		if tmpl == "" {
			tmpl = m["en"]
		}
	} else if m, ok := internalMessages[key]; ok {
		tmpl = m[lang]
		if tmpl == "" {
			tmpl = m["en"]
		}
	}

	// 如果没找到，直接返回key
	if tmpl == "" {
		return key
	}

	// 如果有参数，使用fmt.Sprintf格式化
	if len(args) > 0 {
		return fmt.Sprintf(tmpl, args...)
	}
	return tmpl
}

// TL 是 T 的便捷版本，从gin.Context自动提取语言
func TL(c *gin.Context, key string, args ...interface{}) string {
	lang := LangFromContext(c)
	return T(lang, key, args...)
}
