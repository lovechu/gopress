package shortcodes

import "html"

// ==================== Alert 提示框短代码 ====================

// AlertShortcode 提示框短代码（改进版）
type AlertShortcode struct{}

// Name 返回短代码名称
func (s *AlertShortcode) Name() string { return "alert" }

// Render 渲染提示框
// 参数：
//   - type: 类型（info/success/warning/danger，默认：info）
//   - title: 标题（可选）
//   - dismissible: 可关闭（true/false，默认：false）
func (s *AlertShortcode) Render(params map[string]string, content string) string {
	alertType := GetParam(params, "type", "info")
	title := GetParam(params, "title", "")
	dismissible := GetParam(params, "dismissible", "false")

	// 允许的类型
	validTypes := map[string]bool{
		"info": true, "success": true, "warning": true, "danger": true,
	}
	if !validTypes[alertType] {
		alertType = "info"
	}

	// 如果没有内容，使用默认文本
	if content == "" {
		switch alertType {
		case "info":
			content = "This is an info message."
		case "success":
			content = "Operation completed successfully!"
		case "warning":
			content = "Please note this warning."
		case "danger":
			content = "An error has occurred."
		default:
			content = "This is an alert message."
		}
	}

	// 构建CSS类名
	class := "alert alert-" + html.EscapeString(alertType)
	if dismissible == "true" {
		class += " alert-dismissible"
	}

	// 生成HTML
	result := `<div class="` + class + `" role="alert">`

	// 添加标题
	if title != "" {
		result += `<h4 class="alert-title">` + html.EscapeString(title) + `</h4>`
	}

	// 添加内容
	result += `<div class="alert-content">` + html.EscapeString(content) + `</div>`

	// 添加关闭按钮（如果可关闭）
	if dismissible == "true" {
		result += `<button type="button" class="close" onclick="this.parentElement.style.display='none'">&times;</button>`
	}

	result += `</div>`

	// 添加CSS
	result += `<style>
	.alert {
		padding: 12px 20px;
		margin-bottom: 16px;
		border: 1px solid transparent;
		border-radius: 4px;
	}
	.alert-info {
		color: #0c5460;
		background-color: #d1ecf1;
		border-color: #bee5eb;
	}
	.alert-success {
		color: #155724;
		background-color: #d4edda;
		border-color: #c3e6cb;
	}
	.alert-warning {
		color: #856404;
		background-color: #fff3cd;
		border-color: #ffeaa7;
	}
	.alert-danger {
		color: #721c24;
		background-color: #f8d7da;
		border-color: #f5c6cb;
	}
	.alert-title {
		font-size: 16px;
		font-weight: 600;
		margin-bottom: 8px;
	}
	.alert-content {
		font-size: 14px;
	}
	.alert-dismissible {
		position: relative;
		padding-right: 40px;
	}
	.alert-dismissible .close {
		position: absolute;
		top: 8px;
		right: 12px;
		background: none;
		border: none;
		font-size: 20px;
		cursor: pointer;
	}
	</style>`

	return result
}
