package theme

import (
	"fmt"
	"regexp"
)

// ShortcodeHandler 短代码处理器接口
type ShortcodeHandler interface {
	Name() string
	Render(params map[string]string, content string) string
}

// ShortcodeParser 短代码解析器接口
type ShortcodeParser interface {
	Parse(content string) string
}

// ShortcodeRegistry 短代码注册表
type ShortcodeRegistry struct {
	handlers map[string]ShortcodeHandler
}

// NewShortcodeRegistry 创建短代码注册表
func NewShortcodeRegistry() *ShortcodeRegistry {
	return &ShortcodeRegistry{
		handlers: make(map[string]ShortcodeHandler),
	}
}

// Register 注册短代码处理器
func (r *ShortcodeRegistry) Register(handler ShortcodeHandler) {
	r.handlers[handler.Name()] = handler
}

// Parse 解析内容中的短代码
func (r *ShortcodeRegistry) Parse(content string) string {
	// 正则表达式匹配短代码: [name key=value key="string value"]
	re := regexp.MustCompile(`\[([a-z_][a-z0-9_]*)([^\]]*)\]`)

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		// 提取短代码名称和参数部分
		parts := re.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		name := parts[1]
		paramsStr := parts[2]

		// 查找处理器
		handler, ok := r.handlers[name]
		if !ok {
			// 未找到处理器，返回原始内容
			return match
		}

		// 解析参数
		params := parseParams(paramsStr)

		// 渲染短代码
		return handler.Render(params, "")
	})

	return result
}

// parseParams 解析参数字符串
// 支持格式: key=value 或 key="value with spaces"
func parseParams(paramsStr string) map[string]string {
	params := make(map[string]string)

	// 正则表达式匹配 key=value 或 key="value"
	re := regexp.MustCompile(`(\w+)=("[^"]*"|'[^']*'|[\w-]+)`)

	matches := re.FindAllStringSubmatch(paramsStr, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		key := match[1]
		value := match[2]

		// 去除引号
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') {
			value = value[1 : len(value)-1]
		}

		params[key] = value
	}

	return params
}

// ---- 内置短代码实现 ----

// GalleryShortcode 图库短代码
type GalleryShortcode struct{}

func (g *GalleryShortcode) Name() string {
	return "gallery"
}

func (g *GalleryShortcode) Render(params map[string]string, content string) string {
	id := params["id"]
	if id == "" {
		return "<!-- gallery: missing id parameter -->"
	}

	return fmt.Sprintf(`<div class="gallery" data-id="%s">
	<div class="gallery-grid">
		<!-- Gallery images would be loaded here -->
		<p>Gallery ID: %s</p>
	</div>
</div>`, id, id)
}

// VideoShortcode 视频短代码
type VideoShortcode struct{}

func (v *VideoShortcode) Name() string {
	return "video"
}

func (v *VideoShortcode) Render(params map[string]string, content string) string {
	url := params["url"]
	if url == "" {
		return "<!-- video: missing url parameter -->"
	}

	return fmt.Sprintf(`<div class="video-container">
	<video controls>
		<source src="%s" type="video/mp4">
		Your browser does not support the video tag.
	</video>
</div>`, url)
}

// ButtonShortcode 按钮短代码
type ButtonShortcode struct{}

func (b *ButtonShortcode) Name() string {
	return "button"
}

func (b *ButtonShortcode) Render(params map[string]string, content string) string {
	text := params["text"]
	url := params["url"]
	style := params["style"]

	if text == "" {
		text = "Click Here"
	}
	if url == "" {
		url = "#"
	}
	if style == "" {
		style = "primary"
	}

	// 允许的样式
	validStyles := map[string]bool{
		"primary": true,
		"secondary": true,
		"success": true,
		"danger": true,
		"warning": true,
		"info": true,
	}
	if !validStyles[style] {
		style = "primary"
	}

	return fmt.Sprintf(`<a href="%s" class="btn btn-%s">%s</a>`, url, style, text)
}

// AlertShortcode 提示框短代码
type AlertShortcode struct{}

func (a *AlertShortcode) Name() string {
	return "alert"
}

func (a *AlertShortcode) Render(params map[string]string, content string) string {
	alertType := params["type"]
	text := params["text"]

	if text == "" {
		return "<!-- alert: missing text parameter -->"
	}
	if alertType == "" {
		alertType = "info"
	}

	// 允许的类型
	validTypes := map[string]bool{
		"info": true,
		"success": true,
		"warning": true,
		"danger": true,
	}
	if !validTypes[alertType] {
		alertType = "info"
	}

	return fmt.Sprintf(`<div class="alert alert-%s" role="alert">
	%s
</div>`, alertType, text)
}

// RegisterDefaultShortcodes 注册默认短代码
func RegisterDefaultShortcodes(registry *ShortcodeRegistry) {
	registry.Register(&GalleryShortcode{})
	registry.Register(&VideoShortcode{})
	registry.Register(&ButtonShortcode{})
	registry.Register(&AlertShortcode{})
}
