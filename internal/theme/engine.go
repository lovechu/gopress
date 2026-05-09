package theme

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// 错误定义
var (
	ErrThemeNotFound    = errors.New("theme not found")
	ErrInvalidTheme     = errors.New("invalid theme configuration")
	ErrThemeNotActive   = errors.New("theme is not active")
	ErrTemplateNotFound = errors.New("template not found")
)

// TemplatePathResolver 模板路径解析器
type TemplatePathResolver interface {
	ResolveTemplate(themeName, templateName string) (string, error)
}

// ThemeEngine 主题引擎接口
type ThemeEngine interface {
	GetActiveTheme() (*ThemeInfo, error)
	LoadTheme(name string) error
	RenderHome(ctx context.Context, data HomeContext) (string, error)
	RenderPost(ctx context.Context, data PostContext) (string, error)
	RenderPage(ctx context.Context, data PageContext) (string, error)
}

// ThemeMetrics holds template loading metrics
type ThemeMetrics struct {
	LoadTime      time.Duration
	TemplateCount int
	LoadedAt      time.Time
}

// TemplateStats holds template statistics
type TemplateStats struct {
	TotalTemplates   int
	TotalThemes      int
	ActiveTheme      string
	TotalLoadTime    time.Duration
	LastWarmTime     time.Time
	PerThemeMetrics  map[string]ThemeMetrics
}

// engine 主题引擎实现
type engine struct {
	themesPath      string
	activeTheme     string
	themes          map[string]*ThemeInfo
	shortcodeParser ShortcodeParser
	mu              sync.RWMutex
	metrics         map[string]ThemeMetrics
	lastWarmTime    time.Time
	logger          *zap.Logger
}

// NewEngine 创建主题引擎
func NewEngine(themesPath string, shortcodeParser ShortcodeParser, logger *zap.Logger) *engine {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &engine{
		themesPath:      themesPath,
		themes:          make(map[string]*ThemeInfo),
		shortcodeParser: shortcodeParser,
		metrics:         make(map[string]ThemeMetrics),
		logger:          logger,
	}
}

// LoadTheme 加载指定主题
func (e *engine) LoadTheme(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()
	themePath := filepath.Join(e.themesPath, name)

	// 检查目录是否存在
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return fmt.Errorf("load theme: %w", ErrThemeNotFound)
	}

	// 读取 theme.yaml
	configPath := filepath.Join(themePath, "theme.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read theme config: %w", err)
	}

	var config ThemeConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("parse theme config: %w", err)
	}

	// 验证必填字段
	if config.Name == "" || config.Slug == "" {
		return fmt.Errorf("load theme: %w", ErrInvalidTheme)
	}

	// 构建主题信息
	info := &ThemeInfo{
		Config:      config,
		RootPath:    themePath,
		TemplateDir: filepath.Join(themePath, "templates"),
		AssetsDir:   filepath.Join(themePath, "assets"),
	}

	// 加载模板
	templates, err := e.loadTemplates(info.TemplateDir)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}
	info.Templates = templates

	// 缓存主题
	e.themes[name] = info

	// Track metrics
	loadTime := time.Since(startTime)
	e.metrics[name] = ThemeMetrics{
		LoadTime:      loadTime,
		TemplateCount: len(templates),
		LoadedAt:      time.Now(),
	}

	e.logger.Info("theme loaded",
		zap.String("name", name),
		zap.Duration("load_time", loadTime),
		zap.Int("template_count", len(templates)),
	)

	return nil
}

// loadTemplates 加载模板文件
func (e *engine) loadTemplates(templateDir string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	// 定义模板文件映射（map key是模板名称，用于ExecuteTemplate）
	templateFiles := map[string]string{
		"index":  "index.html",
		"single": "single.html",
		"page":   "page.html",
	}

	// 注册模板函数
	funcMap := template.FuncMap{
		"dateformat": dateFormat,
		"truncate":   truncate,
		"safeHTML":   safeHTML,
		"add":        add,
		"sub":        sub,
		"len":        myLen,
	}

	for name, file := range templateFiles {
		templatePath := filepath.Join(templateDir, file)
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			continue // 模板文件不存在，跳过
		}

		// 创建模板并注册函数
		tmpl := template.New(name).Funcs(funcMap)

		// 解析模板文件（文件中应包含 {{ define "index" }} 等）
		_, err := tmpl.ParseFiles(templatePath)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", file, err)
		}

		templates[name] = tmpl
	}

	return templates, nil
}

// GetActiveTheme 获取当前激活主题
func (e *engine) GetActiveTheme() (*ThemeInfo, error) {
	if e.activeTheme == "" {
		return nil, ErrThemeNotActive
	}

	info, ok := e.themes[e.activeTheme]
	if !ok {
		return nil, ErrThemeNotFound
	}

	return info, nil
}

// SetActiveTheme 设置激活主题
func (e *engine) SetActiveTheme(name string) error {
	if _, ok := e.themes[name]; !ok {
		// 尝试加载主题
		if err := e.LoadTheme(name); err != nil {
			return err
		}
	}

	e.activeTheme = name
	return nil
}

// RenderHome 渲染首页
func (e *engine) RenderHome(ctx context.Context, data HomeContext) (string, error) {
	info, err := e.GetActiveTheme()
	if err != nil {
		return "", err
	}

	tmpl, ok := info.Templates["index"]
	if !ok {
		return "", ErrTemplateNotFound
	}

	// 解析内容中的短代码
	if e.shortcodeParser != nil {
		for i := range data.Posts {
			data.Posts[i].Excerpt = e.shortcodeParser.Parse(data.Posts[i].Excerpt)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "index", data); err != nil {
		return "", fmt.Errorf("render home: %w", err)
	}

	return buf.String(), nil
}

// RenderPost 渲染文章页
func (e *engine) RenderPost(ctx context.Context, data PostContext) (string, error) {
	info, err := e.GetActiveTheme()
	if err != nil {
		return "", err
	}

	tmpl, ok := info.Templates["single"]
	if !ok {
		return "", ErrTemplateNotFound
	}

	// 解析内容中的短代码
	if e.shortcodeParser != nil {
		data.Post.Content = e.shortcodeParser.Parse(data.Post.Content)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "single", data); err != nil {
		return "", fmt.Errorf("render post: %w", err)
	}

	return buf.String(), nil
}

// RenderPage 渲染页面
func (e *engine) RenderPage(ctx context.Context, data PageContext) (string, error) {
	info, err := e.GetActiveTheme()
	if err != nil {
		return "", err
	}

	tmpl, ok := info.Templates["page"]
	if !ok {
		return "", ErrTemplateNotFound
	}

	// 解析内容中的短代码
	if e.shortcodeParser != nil {
		data.Page.Content = e.shortcodeParser.Parse(data.Page.Content)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "page", data); err != nil {
		return "", fmt.Errorf("render page: %w", err)
	}

	return buf.String(), nil
}

// LoadAllThemes 加载所有主题
func (e *engine) LoadAllThemes() error {
	entries, err := os.ReadDir(e.themesPath)
	if err != nil {
		return fmt.Errorf("read themes directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if err := e.LoadTheme(name); err != nil {
			// 记录错误但继续加载其他主题
			fmt.Printf("failed to load theme %s: %v\n", name, err)
		}
	}

	return nil
}

// ListThemes 列出所有已加载的主题
func (e *engine) ListThemes() []*Theme {
	themes := make([]*Theme, 0, len(e.themes))
	for _, info := range e.themes {
		theme := &Theme{
			Name:        info.Config.Name,
			Slug:        info.Config.Slug,
			Version:     info.Config.Version,
			Author:      info.Config.Author,
			Description: info.Config.Description,
			Screenshot:  info.Config.Screenshot,
			IsActive:    info.Config.Slug == e.activeTheme,
		}
		themes = append(themes, theme)
	}
	return themes
}

// Prewarm loads all templates in advance for faster first request
func (e *engine) Prewarm(ctx context.Context) error {
	startTime := time.Now()
	e.logger.Info("starting template prewarming")

	if err := e.LoadAllThemes(); err != nil {
		return fmt.Errorf("prewarm themes: %w", err)
	}

	// Preload active theme templates into memory
	if _, err := e.GetActiveTheme(); err != nil {
		e.logger.Warn("no active theme to preload")
	}

	e.lastWarmTime = time.Now()
	e.logger.Info("template prewarming completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("themes_loaded", len(e.themes)),
	)

	return nil
}

// GetStats returns template loading statistics
func (e *engine) GetStats() TemplateStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var totalLoadTime time.Duration
	perThemeMetrics := make(map[string]ThemeMetrics)

	for name, m := range e.metrics {
		totalLoadTime += m.LoadTime
		perThemeMetrics[name] = m
	}

	return TemplateStats{
		TotalTemplates:   0,
		TotalThemes:      len(e.themes),
		ActiveTheme:      e.activeTheme,
		TotalLoadTime:    totalLoadTime,
		LastWarmTime:     e.lastWarmTime,
		PerThemeMetrics:  perThemeMetrics,
	}
}

// GetTemplateLoadTime returns the load time for a specific theme
func (e *engine) GetTemplateLoadTime(name string) time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if m, ok := e.metrics[name]; ok {
		return m.LoadTime
	}
	return 0
}

// IsWarm returns true if templates have been prewarmed
func (e *engine) IsWarm() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return !e.lastWarmTime.IsZero()
}

// ---- 模板函数 ----

// dateFormat 日期格式化
func dateFormat(t time.Time, layout string) string {
	return t.Format(layout)
}

// truncate 截断字符串
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// safeHTML 标记字符串为安全的HTML
func safeHTML(s string) template.HTML {
	return template.HTML(s)
}

// add 加法
func add(a, b int) int {
	return a + b
}

// sub 减法
func sub(a, b int) int {
	return a - b
}

// myLen 获取长度
func myLen(v interface{}) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []interface{}:
		return len(val)
	default:
		return 0
	}
}

// ---- 短代码正则表达式 ----

var shortcodePattern = regexp.MustCompile(`\[([a-z][a-z0-9_]*)([^\]]*)\]`)

// ParseShortcodes 解析短代码（公开方法）
func (e *engine) ParseShortcodes(content string) string {
	if e.shortcodeParser != nil {
		return e.shortcodeParser.Parse(content)
	}
	return content
}
