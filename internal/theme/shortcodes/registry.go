package shortcodes

import (
	"strings"
	"sync"

	"github.com/yourorg/gopress/internal/theme"
)

// ==================== 短代码注册表（全局单例） ====================

var (
	// 全局短代码注册表实例
	globalRegistry *theme.ShortcodeRegistry
	once           sync.Once
)

// GetGlobalParser 获取全局短代码注册表（单例模式）
func GetGlobalParser() *theme.ShortcodeRegistry {
	once.Do(func() {
		globalRegistry = theme.NewShortcodeRegistry()
		theme.RegisterDefaultShortcodes(globalRegistry)
	})
	return globalRegistry
}

// RegisterGlobalShortcode 注册全局短代码
func RegisterGlobalShortcode(handler theme.ShortcodeHandler) {
	parser := GetGlobalParser()
	parser.Register(handler)
}

// ParseGlobalShortcodes 使用全局解析器解析短代码
func ParseGlobalShortcodes(content string) string {
	parser := GetGlobalParser()
	return parser.Parse(content)
}

// ==================== 短代码工具函数 ====================

// EscapeHTML 转义HTML特殊字符
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// Attr 生成HTML属性字符串
func Attr(attrs map[string]string) string {
	var parts []string
	for k, v := range attrs {
		parts = append(parts, k+`="`+EscapeHTML(v)+`"`)
	}
	return strings.Join(parts, " ")
}

// GetParam 从参数map中获取值，如果不存在则返回默认值
func GetParam(params map[string]string, key, defaultValue string) string {
	if v, ok := params[key]; ok {
		return v
	}
	return defaultValue
}

// HasParam 检查参数是否存在
func HasParam(params map[string]string, key string) bool {
	_, ok := params[key]
	return ok
}
