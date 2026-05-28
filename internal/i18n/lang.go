package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const langContextKey = "i18n_lang"

// AcceptLanguage 解析 Accept-Language header 并返回最佳匹配的语言代码。
// 支持 "zh-CN", "zh", "en" 格式。如果无匹配，返回 defaultLang。
func AcceptLanguage(header string) string {
	if header == "" {
		return defaultLang
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		// 移除质量值 (q=...)
		lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if lang == "" {
			continue
		}

		// 取基础语言（去掉区域后缀，如 en-US -> en）
		base := strings.ToLower(strings.SplitN(lang, "-", 2)[0])

		switch base {
		case "zh":
			return "zh-CN"
		case "en":
			return "en"
		}
	}

	return defaultLang
}

// SetLang 将解析出的语言存入 gin.Context（每个请求只调用一次）
func SetLang(c *gin.Context, lang string) {
	c.Set(langContextKey, lang)
}

// LangFromContext 从 gin.Context 提取语言；如果未设置，返回 defaultLang
func LangFromContext(c *gin.Context) string {
	if lang, ok := c.Get(langContextKey); ok {
		if s, ok := lang.(string); ok && s != "" {
			return s
		}
	}
	return defaultLang
}
