package theme

import (
	"fmt"
	"html"
	"regexp"
	"strings"
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

// openTagRegex 匹配开始标签: [name params] 或 [name]
var openTagRegex = regexp.MustCompile(`\[([a-z_][a-z0-9_]*)([^\]]*)\]`)

// Parse 解析内容中的短代码
// 支持两种格式:
//   - 自闭合: [name key=value]
//   - 环绕:   [name key=value]content[/name]
//   - 嵌套:   [outer][inner]...[/inner]...[/outer]
func (r *ShortcodeRegistry) Parse(content string) string {
	return r.parseShortcodes(content)
}

// findCloseTag 在 rest 中查找 name 对应的闭合标签，正确处理嵌套同名标签。
// 返回闭合标签 [/name] 在 rest 中的起始位置（-1 表示未找到）。
func findCloseTag(rest string, name string) int {
	openPrefix := "[" + name
	closeTag := "[/" + name + "]"
	depth := 1
	searchFrom := 0

	for depth > 0 && searchFrom < len(rest) {
		// 优先查找下一个开始标签或闭合标签
		nextOpen := strings.Index(rest[searchFrom:], openPrefix)
		nextClose := strings.Index(rest[searchFrom:], closeTag)

		if nextClose == -1 {
			return -1 // 没有闭合标签了
		}

		absClose := searchFrom + nextClose

		if nextOpen != -1 && nextOpen < nextClose {
			absOpen := searchFrom + nextOpen
			// 确认是完整的开始标签（后面跟着 ] 或空格/字母）
			afterOpen := absOpen + len(openPrefix)
			if afterOpen < len(rest) && (rest[afterOpen] == ']' || rest[afterOpen] == ' ' || (rest[afterOpen] >= 'a' && rest[afterOpen] <= 'z') || (rest[afterOpen] >= '0' && rest[afterOpen] <= '9') || rest[afterOpen] == '_') {
				depth++
				searchFrom = afterOpen
				continue
			}
		}

		depth--
		if depth == 0 {
			return absClose
		}
		searchFrom = absClose + len(closeTag)
	}

	return -1
}

// parseShortcodes 递归解析短代码
func (r *ShortcodeRegistry) parseShortcodes(content string) string {
	result := &strings.Builder{}
	i := 0

	for i < len(content) {
		// 查找下一个开始标签
		loc := openTagRegex.FindStringIndex(content[i:])
		if loc == nil {
			result.WriteString(content[i:])
			break
		}

		// 写入开始标签之前的文本
		result.WriteString(content[i : i+loc[0]])

		// 提取短代码名称和参数
		match := content[i+loc[0] : i+loc[1]]
		parts := openTagRegex.FindStringSubmatch(match)
		name := parts[1]
		paramsStr := parts[2]

		// 查找对应的闭合标签（正确处理嵌套同名标签）
		rest := content[i+loc[1]:]
		closePos := findCloseTag(rest, name)

		if closePos >= 0 {
			// 找到闭合标签 → 环绕短代码
			closeTag := "[/" + name + "]"
			innerContent := rest[:closePos]

			handler, ok := r.handlers[name]
			if ok {
				// 递归处理内部（支持嵌套）
				innerParsed := r.parseShortcodes(innerContent)
				params := parseParams(paramsStr)
				result.WriteString(handler.Render(params, innerParsed))
			} else {
				// 未注册，保留原文（含内部内容），内部也递归处理
				result.WriteString(match)
				result.WriteString(r.parseShortcodes(innerContent))
				result.WriteString(closeTag)
			}

			i = i + loc[1] + closePos + len(closeTag)
		} else {
			// 没有闭合标签 → 自闭合短代码
			handler, ok := r.handlers[name]
			if ok {
				params := parseParams(paramsStr)
				result.WriteString(handler.Render(params, ""))
			} else {
				// 未注册的孤立标签保留原文
				result.WriteString(match)
			}
			i = i + loc[1]
		}
	}

	return result.String()
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

	safeID := html.EscapeString(id)
	return fmt.Sprintf(`<div class="gallery" data-id="%s">
	<div class="gallery-grid">
		<!-- Gallery images would be loaded here -->
		<p>Gallery ID: %s</p>
	</div>
</div>`, safeID, safeID)
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

	safeURL := html.EscapeString(url)
	return fmt.Sprintf(`<div class="video-container">
	<video controls>
		<source src="%s" type="video/mp4">
		Your browser does not support the video tag.
	</video>
</div>`, safeURL)
}

// ButtonShortcode 按钮短代码
// 支持自闭合: [button text="点击" url="/link" style="primary"]
// 支持环绕:   [button url="/link" style="success"]自定义按钮文字[/button]
type ButtonShortcode struct{}

func (b *ButtonShortcode) Name() string {
	return "button"
}

func (b *ButtonShortcode) Render(params map[string]string, content string) string {
	// 环绕模式优先使用 content，否则回退到 text 参数
	text := content
	textFromParam := false
	if text == "" {
		text = params["text"]
		textFromParam = true
	}
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
		"primary":   true,
		"secondary": true,
		"success":   true,
		"danger":    true,
		"warning":   true,
		"info":      true,
	}
	if !validStyles[style] {
		style = "primary"
	}

	safeURL := html.EscapeString(url)
	safeStyle := html.EscapeString(style)
	// content 来自标签体（已递归解析），不转义；params["text"] 来自用户输入，需转义
	if textFromParam {
		text = html.EscapeString(text)
	}
	return fmt.Sprintf(`<a href="%s" class="btn btn-%s">%s</a>`, safeURL, safeStyle, text)
}

// AlertShortcode 提示框短代码
// 支持自闭合: [alert type="warning" text="注意"]
// 支持环绕:   [alert type="warning"]自定义提示内容[/alert]
type AlertShortcode struct{}

func (a *AlertShortcode) Name() string {
	return "alert"
}

func (a *AlertShortcode) Render(params map[string]string, content string) string {
	alertType := params["type"]

	if alertType == "" {
		alertType = "info"
	}

	// 允许的类型
	validTypes := map[string]bool{
		"info":    true,
		"success": true,
		"warning": true,
		"danger":  true,
	}
	if !validTypes[alertType] {
		alertType = "info"
	}

	// 环绕模式优先使用 content，否则回退到 text 参数
	text := content
	textFromParam := false
	if text == "" {
		text = params["text"]
		textFromParam = true
	}
	if text == "" {
		return "<!-- alert: missing content or text parameter -->"
	}

	safeType := html.EscapeString(alertType)
	// content 来自标签体（已递归解析），不转义；params["text"] 来自用户输入，需转义
	if textFromParam {
		text = html.EscapeString(text)
	}
	return fmt.Sprintf(`<div class="alert alert-%s" role="alert">
	%s
</div>`, safeType, text)
}

// RegisterDefaultShortcodes 注册默认短代码
func RegisterDefaultShortcodes(registry *ShortcodeRegistry) {
	registry.Register(&GalleryShortcode{})
	registry.Register(&VideoShortcode{})
	registry.Register(&ButtonShortcode{})
	registry.Register(&AlertShortcode{})
	registry.Register(&HighlightShortcode{})
}

// HighlightShortcode 高亮短代码（环绕型）
// 用法: [highlight color="yellow"]要高亮的文本[/highlight]
type HighlightShortcode struct{}

func (h *HighlightShortcode) Name() string {
	return "highlight"
}

func (h *HighlightShortcode) Render(params map[string]string, content string) string {
	color := params["color"]
	if color == "" {
		color = "yellow"
	}

	// 允许的颜色
	validColors := map[string]bool{
		"yellow": true,
		"red":    true,
		"green":  true,
		"blue":   true,
		"gray":   true,
	}
	if !validColors[color] {
		color = "yellow"
	}

	if content == "" {
		return "<!-- highlight: missing content -->"
	}

	safeColor := html.EscapeString(color)
	// content 来自标签体（已递归解析），不转义
	return fmt.Sprintf(`<mark class="highlight-%s">%s</mark>`, safeColor, content)
}
