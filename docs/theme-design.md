# GoPress 主题系统架构设计文档

**文档版本**: v1.0
**创建日期**: 2026-05-05
**架构师**: 高见远 (gao)
**状态**: 初稿

---

## 1. 实现方案与框架选型

### 1.1 模板引擎选型

**选型**: Go 原生 `html/template` 包

**理由**:
- 标准库，无需额外依赖
- 支持模板继承（block/define）
- 内置 HTML 转义，安全性好
- 与 Gin 框架天然集成

**备选方案对比**:
| 方案 | 优点 | 缺点 |
|------|------|------|
| html/template | 标准库、安全、无依赖 | 功能较基础 |
| pongo2 | 功能丰富 | 需要额外学习语法 |
| ace | 简洁 | 社区活跃度低 |

### 1.2 主题加载策略

**策略**: 文件系统扫描 + 内存缓存

```
启动时加载:
themes/ 目录扫描 → 解析 theme.yaml → 缓存 Theme 对象

访问时:
缓存命中 → 直接使用
缓存未命中 → 重新加载并解析
```

**缓存失效机制**:
- 启动时全量加载
- 主题切换时清空并重新加载
- 开发模式下可配置禁用缓存

### 1.3 短代码解析方案

**方案**: 正则表达式匹配 + 注册表模式

```go
// 正则模式
pattern := `\[([a-z][a-z0-9_]*)([^\]]*)\]`

// 示例匹配:
// [gallery id=1]        → name=gallery, params="id=1"
// [button text="提交" url="/"]  → name=button, params=`text="提交" url="/"`
```

**解析流程**:
1. 使用正则扫描内容中的 `[xxx]` 标签
2. 提取短代码名称和参数字符串
3. 解析参数为 key=value map
4. 从注册表查找处理器
5. 执行处理器生成 HTML
6. 替换原标签

---

## 2. 文件列表及相对路径

```
gopress/
├── internal/
│   └── theme/
│       ├── model.go           # 主题模型定义
│       ├── config.go          # 主题配置结构
│       ├── engine.go          # 主题引擎（核心渲染逻辑）
│       ├── loader.go          # 主题加载器
│       ├── shortcode.go       # 短代码解析器
│       ├── shortcodes/        # 内置短代码实现
│       │   ├── registry.go    # 短代码注册表
│       │   ├── gallery.go     # 图库短代码
│       │   ├── video.go       # 视频短代码
│       │   ├── button.go      # 按钮短代码
│       │   └── alert.go       # 提示框短代码
│       ├── repository.go      # 主题数据访问（数据库）
│       ├── service.go         # 主题服务层
│       ├── handler.go         # HTTP 处理层
│       └── assets.go          # 静态资源服务
├── themes/                    # 用户主题目录
│   └── default/               # 默认主题
│       ├── theme.yaml
│       ├── templates/
│       │   ├── layout/
│       │   │   └── base.gohtml
│       │   ├── home.gohtml
│       │   ├── post.gohtml
│       │   ├── page.gohtml
│       │   └── partials/
│       │       ├── header.gohtml
│       │       └── footer.gohtml
│       └── assets/
│           ├── css/
│           │   └── style.css
│           └── js/
│               └── main.js
├── internal/templates/         # 内置回退模板
│   ├── layout/
│   │   └── base.gohtml
│   ├── home.gohtml
│   ├── post.gohtml
│   └── page.gohtml
└── docs/
    └── theme-design.md        # 本文档
```

---

## 3. 数据结构定义

### 3.1 Theme 主题模型

```go
// Theme 主题实体
type Theme struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `gorm:"type:varchar(100);not null" json:"name"`
    Slug        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
    Version     string    `gorm:"type:varchar(20)" json:"version"`
    Author      string    `gorm:"type:varchar(100)" json:"author"`
    Description string    `gorm:"type:text" json:"description"`
    Screenshot  string    `gorm:"type:varchar(255)" json:"screenshot"`
    IsActive    bool      `gorm:"not null;default:false" json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (Theme) TableName() string { return "themes" }
```

### 3.2 ThemeConfig 主题配置

```go
// ThemeConfig theme.yaml 解析结构
type ThemeConfig struct {
    Name        string `yaml:"name"`
    Slug        string `yaml:"slug"`
    Version     string `yaml:"version"`
    Author      string `yaml:"author"`
    Description string `yaml:"description"`
    Screenshot  string `yaml:"screenshot"`
}

// ThemeInfo 运行时主题信息（合并数据库和文件系统）
type ThemeInfo struct {
    Config      ThemeConfig
    RootPath    string              // 主题根目录
    TemplateDir string              // 模板目录
    AssetsDir   string              // 资源目录
    Templates   map[string]string   // 模板文件缓存
}
```

### 3.3 模板变量结构

```go
// BaseContext 通用上下文（所有模板可用）
type BaseContext struct {
    SiteName    string            `json:"site_name"`
    SiteURL     string            `json:"site_url"`
    CurrentYear int               `json:"current_year"`
    ThemePath   string            `json:"theme_path"`    // /themes/default
    AssetsPath  string            `json:"assets_path"`   // /themes/default/assets
}

// HomeContext 首页上下文
type HomeContext struct {
    BaseContext
    Posts      []PostSummary `json:"posts"`
    Pagination Pagination     `json:"pagination"`
}

// PostContext 文章详情上下文
type PostContext struct {
    BaseContext
    Post       *PostDetail `json:"post"`
    Categories []Taxonomy   `json:"categories"`
    Tags       []Taxonomy   `json:"tags"`
}

// PageContext 页面详情上下文
type PageContext struct {
    BaseContext
    Page *PageDetail `json:"page"`
}

// PostSummary 文章摘要（用于列表）
type PostSummary struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Slug        string    `json:"slug"`
    Excerpt     string    `json:"excerpt"`
    AuthorName  string    `json:"author_name"`
    ViewCount   int       `json:"view_count"`
    PublishedAt time.Time `json:"published_at"`
}

// PostDetail 文章详情
type PostDetail struct {
    ID             uint      `json:"id"`
    Title          string    `json:"title"`
    Slug           string    `json:"slug"`
    Content        string    `json:"content"`       // 已解析短代码的内容
    RawContent     string    `json:"raw_content"`  // 原始内容
    Excerpt        string    `json:"excerpt"`
    AuthorID       uint      `json:"author_id"`
    AuthorName     string    `json:"author_name"`
    ViewCount      int       `json:"view_count"`
    CommentAllowed bool      `json:"comment_allowed"`
    PublishedAt    time.Time `json:"published_at"`
    CreatedAt      time.Time `json:"created_at"`
}

// PageDetail 页面详情
type PageDetail struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Slug        string    `json:"slug"`
    Content     string    `json:"content"`
    RawContent  string    `json:"raw_content"`
    AuthorID    uint      `json:"author_id"`
    AuthorName  string    `json:"author_name"`
    ViewCount   int       `json:"view_count"`
    PublishedAt time.Time `json:"published_at"`
}

// Pagination 分页信息
type Pagination struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
    HasPrev    bool `json:"has_prev"`
    HasNext    bool `json:"has_next"`
}

// Taxonomy 分类/标签
type Taxonomy struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
    Slug string `json:"slug"`
}
```

---

## 4. 接口设计

### 4.1 ThemeEngine 接口

```go
// ThemeEngine 主题渲染引擎接口
type ThemeEngine interface {
    // GetTheme 获取当前激活主题
    GetTheme() (*ThemeInfo, error)

    // ListThemes 列出所有可用主题
    ListThemes() ([]*Theme, error)

    // SetActiveTheme 切换激活主题
    SetActiveTheme(ctx context.Context, slug string) error

    // RenderHome 渲染首页
    RenderHome(ctx context.Context, w io.Writer, data HomeContext) error

    // RenderPost 渲染文章页
    RenderPost(ctx context.Context, w io.Writer, data PostContext) error

    // RenderPage 渲染页面
    RenderPage(ctx context.Context, w io.Writer, data PageContext) error
}

// engine 主题引擎实现
type engine struct {
    repo          Repository
    tmpl          *template.Template
    activeTheme   *ThemeInfo
    themeCache    map[string]*ThemeInfo
    baseDir       string        // themes 目录
    builtinDir    string        // 内置模板目录
    shortcodeParser *ShortcodeParser
}
```

### 4.2 ShortcodeHandler 接口

```go
// ShortcodeHandler 短代码处理器接口
type ShortcodeHandler interface {
    // Name 返回短代码名称
    Name() string

    // Render 渲染短代码
    // params 解析后的参数
    // content 短代码包裹的内容（用于 [shortcode]content[/shortcode] 格式）
    Render(params map[string]string, content string) string
}

// ShortcodeParser 短代码解析器
type ShortcodeParser interface {
    // Register 注册短代码处理器
    Register(handler ShortcodeHandler)

    // Parse 解析内容中的短代码
    Parse(content string) (string, error)
}

// ShortcodeFunc 短代码函数类型（简化版）
type ShortcodeFunc func(params map[string]string) string
```

### 4.3 Repository 接口

```go
// Repository 主题数据访问接口
type Repository interface {
    // FindBySlug 根据slug查找主题
    FindBySlug(ctx context.Context, slug string) (*Theme, error)

    // FindActive 获取激活主题
    FindActive(ctx context.Context) (*Theme, error)

    // List 列出所有主题
    List(ctx context.Context) ([]*Theme, error)

    // SetActive 设置激活主题
    SetActive(ctx context.Context, id uint) error

    // Create 创建主题记录
    Create(ctx context.Context, theme *Theme) error

    // Delete 删除主题
    Delete(ctx context.Context, id uint) error
}
```

### 4.4 Service 接口

```go
// Service 主题服务接口
type Service interface {
    // GetActiveTheme 获取激活主题
    GetActiveTheme(ctx context.Context) (*ThemeInfo, error)

    // ListThemes 列出所有主题
    ListThemes(ctx context.Context) ([]*Theme, error)

    // GetThemeInfo 获取主题详情
    GetThemeInfo(ctx context.Context, slug string) (*ThemeInfo, error)

    // SetActiveTheme 切换激活主题
    SetActiveTheme(ctx context.Context, slug string) error

    // ReloadTheme 重新加载主题（开发模式）
    ReloadTheme(ctx context.Context, slug string) error
}
```

### 4.5 Handler 接口

```go
// Handler HTTP 处理层
type Handler struct {
    svc Service
}

// RegisterPublicRoutes 注册公开路由
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup)

// RegisterAuthRoutes 注册认证路由
func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup)

// RegisterAdminRoutes 注册管理路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup)
```

---

## 5. 程序调用流程

### 5.1 页面访问流程（时序图）

```
┌─────────┐     ┌──────────┐     ┌─────────┐     ┌────────┐     ┌─────────┐
│  用户   │     │   Gin    │     │ Handler │     │Service │     │ Engine  │
└────┬────┘     └────┬─────┘     └────┬────┘     └───┬────┘     └───┬─────┘
     │               │               │               │             │
     │ GET /post/hello               │               │             │
     │───────────>│                  │               │             │
     │               │               │               │             │
     │               │ GetPostBySlug │               │             │
     │               │──────────────>│               │             │
     │               │               │               │             │
     │               │               │ List(ctx,filter)            │
     │               │               │──────────────>│             │
     │               │               │               │             │
     │               │               │   PostDTO      │             │
     │               │               │<───────────────│             │
     │               │               │               │             │
     │               │   PostContext │               │             │
     │               │<──────────────│               │             │
     │               │               │               │             │
     │               │    RenderPost │               │             │
     │               │──────────────>│               │             │
     │               │               │               │             │
     │               │               │ ParseShortcodes             │
     │               │               │──────────────>│             │
     │               │               │               │             │
     │               │               │  Render(template, data)     │
     │               │               │──────────────>│             │
     │               │               │               │             │
     │               │               │    HTML       │             │
     │               │               │<───────────────│             │
     │   200 OK     │               │               │             │
     │<──────────────│               │               │             │
     │   HTML        │               │               │             │
     │               │               │               │             │
```

### 5.2 短代码解析流程

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│    原始内容   │───>│    正则扫描   │───>│   提取参数   │───>│  查找处理器  │
│ [gallery id=1]│    │  匹配 [xxx]  │    │  key=value  │    │  registry   │
└──────────────┘    └──────────────┘    └──────────────┘    └──────┬───────┘
                                                                     │
                    ┌──────────────┐    ┌──────────────┐    ┌────────▼───────┐
                    │   HTML 输出   │<───│   替换标签   │<───│   执行处理器   │
                    │ <div class=...│    │  替换原标签  │    │ Render(params)│
                    └──────────────┘    └──────────────┘    └────────────────┘
```

### 5.3 主题加载流程

```
┌─────────────────────────────────────────────────────────────────┐
│                        应用启动                                  │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. 扫描 themes/ 目录                                           │
│     - 查找所有子目录                                             │
│     - 验证 theme.yaml 存在                                       │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. 解析 theme.yaml                                              │
│     - 读取 YAML 配置                                             │
│     - 验证必填字段（name, slug）                                │
│     - 构建 ThemeInfo 结构                                        │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. 加载模板文件                                                 │
│     - 扫描 templates/ 目录                                        │
│     - 解析 .gohtml 文件                                          │
│     - 缓存到 memory                                              │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│  4. 缓存主题                                                     │
│     - 存入 themeCache map                                        │
│     - 设置 activeTheme                                           │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│  5. 注册路由                                                    │
│     - GET /                    → RenderHome                     │
│     - GET /post/:slug          → RenderPost                      │
│     - GET /:slug               → RenderPage                      │
│     - GET /themes/*filepath    → 静态资源                        │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. 任务列表（有序、含依赖）

### 任务 1：创建主题目录结构和模型

**依赖**: 无
**输出**:
- `internal/theme/model.go`
- `internal/theme/config.go`
- 创建 `themes/default/` 目录

**内容**:
- Theme 模型定义
- ThemeConfig 结构（YAML 解析）
- ThemeInfo 运行时信息结构
- 模板变量结构（BaseContext、PostContext 等）

### 任务 2：实现主题引擎（核心渲染逻辑）

**依赖**: 任务 1
**输出**: `internal/theme/engine.go`
**内容**:
- ThemeEngine 接口
- engine 结构体实现
- 模板加载和缓存
- RenderHome/RenderPost/RenderPage 方法
- 模板函数注册（dateformat、truncate 等）

### 任务 3：实现短代码解析器

**依赖**: 任务 1
**输出**:
- `internal/theme/shortcode.go`
- `internal/theme/shortcodes/registry.go`
- `internal/theme/shortcodes/gallery.go`
- `internal/theme/shortcodes/video.go`
- `internal/theme/shortcodes/button.go`
- `internal/theme/shortcodes/alert.go`

**内容**:
- ShortcodeHandler 接口
- ShortcodeParser 实现
- 正则表达式匹配逻辑
- 参数解析逻辑
- 内置短代码实现
- init() 自动注册

### 任务 4：实现 Service 层

**依赖**: 任务 1、2、3
**输出**: `internal/theme/service.go`
**内容**:
- Service 接口
- service 结构体实现
- GetActiveTheme
- ListThemes
- SetActiveTheme
- 缓存管理

### 任务 5：实现 Handler 层和路由

**依赖**: 任务 4
**输出**: `internal/theme/handler.go`
**内容**:
- Handler 结构体
- RegisterPublicRoutes（首页、文章页、页面路由）
- RegisterAuthRoutes（主题管理 API）
- RegisterAdminRoutes（主题切换 API）
- 页面渲染处理函数

### 任务 6：创建默认主题模板

**依赖**: 任务 1
**输出**: `themes/default/`
**内容**:
- `theme.yaml` 主题配置
- `templates/layout/base.gohtml` 基础布局
- `templates/home.gohtml` 首页模板
- `templates/post.gohtml` 文章模板
- `templates/page.gohtml` 页面模板
- `templates/partials/header.gohtml` 页头
- `templates/partials/footer.gohtml` 页脚
- `assets/css/style.css` 样式文件
- `assets/js/main.js` 脚本文件

### 任务 7：集成到 bootstrap

**依赖**: 任务 4、5
**输出**: 修改 `internal/bootstrap/app.go` 和 `router.go`
**内容**:
- 初始化 Theme 模块（repo、service、handler）
- 注册主题路由到 Gin
- 添加主题配置到 Config
- 处理静态资源路由

### 任务 8：更新 gen.py 代码生成器

**依赖**: 任务 1-7
**输出**: 修改 `scripts/gen.py`
**内容**:
- 添加 theme 模块到生成模板
- 生成 theme 相关文件的命令
- 生成主题脚手架的命令

---

## 7. 依赖包列表

### Go 标准库

| 包 | 用途 |
|----|------|
| `html/template` | 模板渲染 |
| `text/template` | 文本模板 |
| `regexp` | 正则表达式 |
| `path/filepath` | 文件路径处理 |
| `io` | IO 操作 |
| `time` | 时间处理 |

### 外部依赖

| 包 | 版本 | 用途 |
|----|------|------|
| `github.com/spf13/viper` | ^1.18 | YAML 配置解析 |
| `gopkg.in/yaml.v3` | ^3.0 | YAML 解析 |

### 内部依赖

| 包 | 依赖关系 |
|----|----------|
| `github.com/yourorg/gopress/internal/config` | 配置读取 |
| `github.com/yourorg/gopress/internal/content` | 文章数据 |
| `github.com/yourorg/gopress/internal/media` | 媒体数据 |
| `github.com/yourorg/gopress/internal/taxonomy` | 分类标签 |
| `github.com/yourorg/gopress/internal/user` | 用户信息 |
| `github.com/yourorg/gopress/pkg/database` | 数据库连接 |

---

## 8. 共享知识

### 8.1 模板函数约定

以下函数将在模板引擎初始化时注册，所有模板均可调用：

```go
// 字符串处理
func truncate(s string, length int) string
func stripTags(s string) string
func markdown(s string) string

// 日期格式化
func dateformat(t time.Time, layout string) string
func date(t time.Time) string  // 格式: 2006-01-02

// HTML 处理
func safeHTML(s string) template.HTML
func rawHTML(s string) template.HTML

// 列表处理
func first(s []any, n int) []any
func last(s []any, n int) []any
func len(s []any) int

// URL 处理
func absURL(path string) string
func relURL(path string) string

// 条件函数
func eq(a, b any) bool
func ne(a, b any) bool
func lt(a, b any) bool
func gt(a, b any) bool
```

### 8.2 命名规范

**Go 文件**:
- `model.go` - 数据模型
- `config.go` - 配置结构
- `engine.go` - 核心引擎
- `loader.go` - 加载器
- `repository.go` - 数据访问
- `service.go` - 业务逻辑
- `handler.go` - HTTP 处理
- `shortcode.go` - 短代码解析
- `shortcodes/*.go` - 具体短代码实现

**结构体命名**:
- `Theme` - 数据库实体
- `ThemeConfig` - YAML 配置
- `ThemeInfo` - 运行时信息
- `BaseContext` - 通用上下文（以 Context 结尾）
- `HomeContext` - 首页上下文
- `PostContext` - 文章上下文

**接口命名**:
- `Engine` / `Engineer` 结尾
- `Repository` 结尾
- `Service` 结尾
- `Handler` 结尾

**变量命名**:
- `engine` - 引擎实例
- `svc` - 服务实例
- `repo` - 仓库实例
- `tmpl` - 模板实例
- `cfg` - 配置实例

### 8.3 错误处理约定

```go
// 使用 errors.New 定义错误
var (
    ErrThemeNotFound    = errors.New("theme not found")
    ErrInvalidTheme     = errors.New("invalid theme configuration")
    ErrThemeNotActive   = errors.New("theme is not active")
    ErrTemplateNotFound = errors.New("template not found")
    ErrShortcodeNotFound = errors.New("shortcode handler not found")
)

// 错误传播使用 fmt.Errorf
return nil, fmt.Errorf("load theme: %w", err)
```

### 8.4 配置约定

主题相关配置添加到 `config.yaml`:

```yaml
app:
  theme_dir: "./themes"        # 主题目录
  builtin_template_dir: "./internal/templates"  # 内置模板目录
  default_theme: "default"    # 默认主题

theme:
  cache_enabled: true         # 是否启用模板缓存
  shortcode_enabled: true     # 是否启用短代码
```

---

## 9. 安全考虑

### 9.1 模板安全
- `html/template` 默认转义 HTML 特殊字符
- 不安全内容使用 `template.HTML` 显式标记
- 禁止在模板中执行任意代码

### 9.2 路径安全
- 验证主题目录不包含 `..` 路径遍历
- 静态资源路径白名单
- 文件读取限制在主题目录内

### 9.3 短代码安全
- 参数白名单验证
- HTML 输出过滤
- 禁止执行系统命令

---

## 10. 性能考虑

### 10.1 模板缓存
- 生产环境启用模板缓存
- `template.Must(template.New().ParseFiles(...))`
- 开发环境可禁用缓存实时 reload

### 10.2 主题缓存
- 启动时全量加载主题信息
- 主题切换时清空并重建缓存
- 使用 `sync.RWMutex` 保护并发访问

### 10.3 静态资源
- 配置 Etag/Last-Modified 头
- 启用浏览器缓存
- 考虑使用 CDN 加速

---

*文档结束*
