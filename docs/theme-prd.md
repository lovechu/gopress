# GoPress 主题系统产品需求文档 (PRD)

**文档版本**: v1.0
**创建日期**: 2026-05-05
**产品经理**: 许清楚 (xu)
**状态**: 初稿

---

## 1. 产品目标

为 GoPress CMS 提供一套灵活、可扩展的主题系统，使开发者能够通过创建主题目录来自定义网站的视觉呈现和页面结构，无需修改核心代码即可实现站点的个性化定制。

**核心价值**：
- **模板覆盖**：支持自定义全套页面模板（首页、文章页、页面页等）
- **资源管理**：内置静态资源（CSS/JS/图片）支持
- **内容扩展**：通过短代码在文章内容中嵌入动态组件
- **开箱即用**：提供默认主题，确保系统零配置可运行

---

## 2. 用户故事

| 角色 | 用户故事 |
|------|----------|
| **站点管理员** | 作为站点管理员，我希望选择一个主题来改变网站外观，无需编写代码 |
| **主题开发者** | 作为主题开发者，我希望按照规范创建主题目录，上传后即可被系统识别和使用 |
| **模板开发者** | 作为模板开发者，我希望使用 Go 原生 html/template 语法编写模板，访问到完整的页面数据 |
| **内容编辑** | 作为内容编辑，我希望在文章中插入 `[gallery id=1]` 这样的短代码来展示图片画廊 |

---

## 3. 需求池

### P0 - 必须实现（MVP）

| ID | 需求 | 描述 | 验收标准 |
|----|------|------|----------|
| P0-1 | 主题加载引擎 | 扫描 themes/ 目录，加载主题元数据（theme.yaml） | 系统启动时加载主题列表，支持获取当前激活主题信息 |
| P0-2 | 基础模板渲染 | 使用 Go html/template 实现服务端渲染 | 可渲染 .gohtml 模板文件，支持模板继承（block/define） |
| P0-3 | 文章/页面路由 | 实现公开访问路由，通过 slug 访问内容 | `/post/{slug}` 渲染文章页，`/{slug}` 渲染页面页 |
| P0-4 | 默认主题 | 内置一个名为 "Default" 的默认主题 | 包含首页列表、文章详情、页面详情三个模板 |

### P1 - 重要功能

| ID | 需求 | 描述 | 验收标准 |
|----|------|------|----------|
| P1-1 | 短代码解析 | 解析文章内容中的短代码标签 | `[gallery id=1]` 可被解析为画廊组件输出 |
| P1-2 | 静态资源服务 | 自动处理主题 assets 目录下的文件 | 访问 `/assets/css/style.css` 返回对应文件 |
| P1-3 | 主题切换 API | 提供切换激活主题的接口 | 可通过配置文件或 API 切换当前主题 |

### P2 - 可选功能

| ID | 需求 | 描述 | 备注 |
|----|------|------|------|
| P2-1 | 主题预览 | 在后台预览未激活主题 | 可按主题分类管理 |
| P2-2 | 主题市场 | 提供主题上传/下载功能 | 远期规划 |

---

## 4. 主题结构规范

### 4.1 目录结构

```
themes/
└── {theme-name}/           # 主题目录名（英文小写，中划线分隔）
    ├── theme.yaml           # 主题元数据文件（必须）
    ├── templates/           # 模板文件目录（必须）
    │   ├── layout/         # 布局模板
    │   │   └── base.gohtml # 基础布局
    │   ├── home.gohtml     # 首页模板
    │   ├── post.gohtml     # 文章详情模板
    │   ├── page.gohtml     # 页面详情模板
    │   └── partials/       # 局部模板
    │       ├── header.gohtml
    │       └── footer.gohtml
    └── assets/             # 静态资源目录
        ├── css/
        │   └── style.css
        ├── js/
        │   └── main.js
        └── images/
            └── logo.png
```

### 4.2 theme.yaml 字段说明

```yaml
name: "Default"              # 主题显示名称
slug: "default"               # 主题标识（目录名）
version: "1.0.0"              # 主题版本
author: "GoPress Team"       # 作者
description: "默认主题"        # 主题描述
screenshot: "screenshot.png"  # 预览图（可选，相对于主题根目录）
```

### 4.3 主题查找顺序

1. 优先使用激活主题目录 `themes/{active-theme}/templates/`
2. 若文件不存在，回退到内置默认模板 `internal/templates/`

---

## 5. 模板规范

### 5.1 支持的模板文件

| 模板文件 | 说明 | 路由 |
|----------|------|------|
| `home.gohtml` | 首页模板 | `/` |
| `post.gohtml` | 文章详情模板 | `/post/{slug}` |
| `page.gohtml` | 页面模板 | `/{slug}` |
| `layout/base.gohtml` | 基础布局（可选，用于模板继承） | - |
| `partials/header.gohtml` | 页头局部模板 | - |
| `partials/footer.gohtml` | 页脚局部模板 | - |

### 5.2 模板变量

**通用变量（所有模板可用）**：

```go
type BaseContext struct {
    SiteName    string            // 站点名称
    SiteURL     string            // 站点 URL
    CurrentYear int               // 当前年份
    ThemePath   string            // 当前主题路径（如 /themes/default）
    AssetsPath  string            // 静态资源路径（如 /themes/default/assets）
}
```

**首页模板变量 (HomeContext)**：

```go
type HomeContext struct {
    BaseContext
    Posts      []PostSummary      // 文章列表
    Pagination Pagination          // 分页信息
}
```

**文章详情模板变量 (PostContext)**：

```go
type PostContext struct {
    BaseContext
    Post       Post               // 文章详情
    Categories []Taxonomy          // 关联分类
    Tags       []Taxonomy          // 关联标签
}
```

**页面模板变量 (PageContext)**：

```go
type PageContext struct {
    BaseContext
    Page       Page               // 页面详情
}
```

### 5.3 模板示例

**layout/base.gohtml（基础布局）**：
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{ block "title" . }}Default{{ end }}</title>
    <link rel="stylesheet" href="{{ .AssetsPath }}/css/style.css">
</head>
<body>
    {{ template "partials/header.gohtml" . }}
    <main>
        {{ block "content" . }}{{ end }}
    </main>
    {{ template "partials/footer.gohtml" . }}
</body>
</html>
```

**post.gohtml（文章页）**：
```html
{{ define "post.gohtml" }}
{{ template "layout/base.gohtml" . }}
{{ define "content" }}
<article class="post">
    <h1>{{ .Post.Title }}</h1>
    <div class="meta">
        <span>作者: {{ .Post.Author }}</span>
        <span>发布时间: {{ .Post.PublishedAt }}</span>
    </div>
    <div class="content">
        {{ .Post.Content }}
    </div>
</article>
{{ end }}
{{ end }}
```

---

## 6. 短代码规范

### 6.1 短代码格式

```html
[shortcode-name key=value key="string value"]
```

- 短代码由方括号包裹
- 第一个词为短代码名称（小写英文）
- 后续为键值对参数，参数名用等号连接
- 字符串值可使用双引号包裹

### 6.2 内置短代码

| 短代码 | 参数 | 说明 | 示例 |
|--------|------|------|------|
| `gallery` | `id`（必需） | 渲染媒体库中的图片画廊 | `[gallery id=1]` |
| `video` | `id`, `type` | 嵌入视频播放器 | `[video id=123 type="mp4"]` |
| `button` | `text`, `url`, `style` | 渲染按钮组件 | `[button text="了解更多" url="/about"]` |
| `alert` | `type`, `message` | 渲染提示框 | `[alert type="info" message="欢迎访问"]` |

### 6.3 短代码解析流程

1. 内容存储时保留原始短代码标签
2. 模板渲染前，调用短代码解析器扫描内容
3. 解析器根据短代码名称查找对应处理器
4. 处理器根据参数生成 HTML 片段替换原标签
5. 返回处理后的内容给模板渲染

### 6.4 自定义短代码扩展

开发者可在 `internal/theme/shortcodes/` 目录下注册新的短代码：

```go
// internal/theme/shortcodes/gallery.go
package shortcodes

func init() {
    Register("gallery", RenderGallery)
}

func RenderGallery(params map[string]string) string {
    id := params["id"]
    // 返回 HTML 片段
    return fmt.Sprintf(`<div class="gallery" data-id="%s">...</div>`, id)
}
```

---

## 7. 待确认问题

### 技术细节

| # | 问题 | 影响 | 优先级 |
|---|------|------|--------|
| 1 | **模板缓存策略**：是否启用模板缓存？缓存失效机制是什么？ | 影响性能和数据一致性 | 中 |
| 2 | **主题隔离级别**：主题是否可以访问核心模块的内部 API？有无安全限制？ | 影响系统安全 | 高 |
| 3 | **多语言支持**：主题模板是否需要内置 i18n 支持？ | 影响国际化场景 | 低 |
| 4 | **SEO 优化**：是否需要在主题系统层面提供 SEO 元标签（meta description、og:image）的支持？ | 影响搜索引擎收录 | 中 |
| 5 | **静态资源版本控制**：是否需要支持资源文件名哈希（如 style.a1b2c3.css）用于缓存？ | 影响前端性能 | 低 |

### 产品方向

| # | 问题 | 影响 | 优先级 |
|---|------|------|--------|
| 6 | **主题配置化**：主题是否支持在 theme.yaml 中定义可配置项（如颜色、Logo），供管理员在后台修改？ | 影响主题灵活性 | 中 |
| 7 | **主题依赖管理**：未来是否可能存在主题依赖（如主题 A 需要插件 B）？是否需要预留扩展点？ | 影响架构演进 | 低 |

---

## 8. 附录

### 8.1 相关文件路径

```
gopress/
├── internal/
│   └── theme/
│       ├── engine.go        # 主题引擎
│       ├── loader.go        # 主题加载器
│       ├── renderer.go      # 模板渲染器
│       ├── router.go        # 公开路由
│       ├── shortcodes/      # 短代码实现
│       │   └── registry.go  # 短代码注册表
│       └── assets.go        # 静态资源服务
├── themes/                  # 用户主题目录
│   └── default/             # 默认主题
│       ├── theme.yaml
│       ├── templates/
│       └── assets/
└── docs/
    └── theme-prd.md         # 本文档
```

### 8.2 后续计划

本 PRD 聚焦 P0/P1 功能，P2 功能（主题预览、主题市场）将在后续迭代中规划。
