package theme

import (
	"html/template"
	"time"
)

// Theme 主题数据库模型
type Theme struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
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

// TableName 指定表名
func (Theme) TableName() string {
	return "themes"
}

// ThemeConfig 主题YAML配置
type ThemeConfig struct {
	Name        string `yaml:"name"`
	Slug        string `yaml:"slug"`
	Version     string `yaml:"version"`
	Author      string `yaml:"author"`
	Description string `yaml:"description"`
	Screenshot  string `yaml:"screenshot"`
}

// ThemeInfo 主题运行时信息（已加载到内存的主题）
type ThemeInfo struct {
	Config      ThemeConfig
	RootPath    string
	TemplateDir string
	AssetsDir   string
	Templates   map[string]*template.Template
}

// BaseContext 通用模板上下文
type BaseContext struct {
	SiteName    string `json:"site_name"`
	SiteURL     string `json:"site_url"`
	CurrentYear int    `json:"current_year"`
	ThemePath   string `json:"theme_path"`
	AssetsPath  string `json:"assets_path"`
}

// HomeContext 首页模板上下文
type HomeContext struct {
	BaseContext
	Posts      []PostSummary `json:"posts"`
	Pagination Pagination    `json:"pagination"`
}

// PostContext 文章详情模板上下文
type PostContext struct {
	BaseContext
	Post       PostDetail `json:"post"`
	Categories []Taxonomy `json:"categories"`
	Tags       []Taxonomy `json:"tags"`
}

// PageContext 页面模板上下文
type PageContext struct {
	BaseContext
	Page PageDetail `json:"page"`
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
	Content        string    `json:"content"`
	RawContent     string    `json:"raw_content"`
	Excerpt        string    `json:"excerpt"`
	AuthorID       uint      `json:"author_id"`
	AuthorName     string    `json:"author_name"`
	ViewCount      int       `json:"view_count"`
	CommentAllowed bool      `json:"comment_allowed"`
	PublishedAt    time.Time `json:"published_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasPrev    bool `json:"has_prev"`
	HasNext    bool `json:"has_next"`
}

// Taxonomy 分类/标签
type Taxonomy struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
