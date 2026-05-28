package i18n

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// 测试用的错误码常量（避免循环导入）
const (
	testCodeNotFound      = "NOT_FOUND"
	testCodeUnauthorized  = "UNAUTHORIZED"
	testCodeForbidden     = "FORBIDDEN"
	testCodeRateLimited   = "RATE_LIMITED"
	testCodeInvalidInput  = "INVALID_INPUT"
	testCodeInternalError = "INTERNAL_ERROR"
)

func TestAcceptLanguage(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"empty header", "", "zh-CN"},
		{"english only", "en", "en"},
		{"chinese simple", "zh", "zh-CN"},
		{"chinese traditional", "zh-CN", "zh-CN"},
		{"english with quality", "en-US,en;q=0.9,zh-CN;q=0.8", "en"},
		{"chinese first with quality", "zh-CN,zh;q=0.8,en;q=0.6", "zh-CN"},
		{"unsupported language", "fr-FR", "zh-CN"},
		{"mixed case", "EN", "en"},
		{"mixed case chinese", "ZH", "zh-CN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AcceptLanguage(tt.header)
			if result != tt.expected {
				t.Errorf("AcceptLanguage(%q) = %q, want %q", tt.header, result, tt.expected)
			}
		})
	}
}

func TestT(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		key      string
		args     []interface{}
		expected string
	}{
		{"english simple", "en", "unauthorized", nil, "Unauthorized"},
		{"chinese simple", "zh-CN", "unauthorized", nil, "未授权"},
		{"english task not found", "en", "task.not_found", nil, "Task not found"},
		{"chinese task not found", "zh-CN", "task.not_found", nil, "未找到任务"},
		{"english with arg", "en", "validation_error", []interface{}{"field"}, "Validation failed: field"},
		{"chinese with arg", "zh-CN", "validation_error", []interface{}{"field"}, "参数校验失败: field"},
		{"unknown language fallback", "fr", "task.not_found", nil, "Task not found"},
		{"unknown key returns key", "en", "unknown.key", nil, "unknown.key"},
		{"unknown key with args returns key", "en", "unknown.key", []interface{}{"test"}, "unknown.key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := T(tt.lang, tt.key, tt.args...)
			if result != tt.expected {
				t.Errorf("T(%q, %q, %v) = %q, want %q", tt.lang, tt.key, tt.args, result, tt.expected)
			}
		})
	}
}

func TestTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		lang     string
		key      string
		expected string
	}{
		{"with chinese context", "zh-CN", "unauthorized", "未授权"},
		{"with english context", "en", "unauthorized", "Unauthorized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			c.Set(langContextKey, tt.lang)
			result := TL(c, tt.key)
			if result != tt.expected {
				t.Errorf("TL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSetLangAndLangFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(nil)

	// 初始状态应返回默认语言
	lang := LangFromContext(c)
	if lang != defaultLang {
		t.Errorf("LangFromContext() = %q, want %q", lang, defaultLang)
	}

	// 设置语言后应返回设置的语言
	SetLang(c, "en")
	lang = LangFromContext(c)
	if lang != "en" {
		t.Errorf("LangFromContext() after SetLang = %q, want %q", lang, "en")
	}
}

func TestSetDefaultLang(t *testing.T) {
	original := defaultLang
	defer func() { defaultLang = original }()

	SetDefaultLang("en")
	if defaultLang != "en" {
		t.Errorf("SetDefaultLang() failed, defaultLang = %q, want %q", defaultLang, "en")
	}

	SetDefaultLang("zh-CN")
	if defaultLang != "zh-CN" {
		t.Errorf("SetDefaultLang() failed, defaultLang = %q, want %q", defaultLang, "zh-CN")
	}
}

// 这些测试在utils包中，避免循环导入
