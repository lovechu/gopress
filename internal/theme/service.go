package theme

import (
	"context"
)

// Service 主题服务接口
type Service interface {
	GetActiveTheme() (*Theme, error)
	ListThemes() ([]*Theme, error)
	SetActiveTheme(name string) error
	RenderHome(posts []PostSummary, page, totalPages int) (string, error)
	RenderPost(post *PostDetail, categories []Taxonomy, tags []Taxonomy) (string, error)
	RenderPage(page *PageDetail) (string, error)
	InitThemes() error
}

type service struct {
	engine *engine
	repo   Repository
}

// NewService 创建主题服务
func NewService(engine *engine, repo Repository) *service {
	return &service{
		engine: engine,
		repo:   repo,
	}
}

// GetActiveTheme 获取激活主题
func (s *service) GetActiveTheme() (*Theme, error) {
	info, err := s.engine.GetActiveTheme()
	if err != nil {
		return nil, err
	}

	return &Theme{
		Name:        info.Config.Name,
		Slug:        info.Config.Slug,
		Version:     info.Config.Version,
		Author:      info.Config.Author,
		Description: info.Config.Description,
		Screenshot:  info.Config.Screenshot,
		IsActive:    true,
	}, nil
}

// ListThemes 列出所有主题
func (s *service) ListThemes() ([]*Theme, error) {
	// 从引擎获取已加载的主题
	themes := s.engine.ListThemes()
	return themes, nil
}

// SetActiveTheme 设置激活主题
func (s *service) SetActiveTheme(name string) error {
	// 先尝试设置，如果主题未加载会自动加载
	if err := s.engine.SetActiveTheme(name); err != nil {
		return err
	}

	// 如果有 repository，更新数据库
	if s.repo != nil {
		// 先取消所有主题的激活状态
		// 然后设置新的激活主题
		// 这里需要 repo 的具体实现
	}

	return nil
}

// RenderHome 渲染首页
func (s *service) RenderHome(posts []PostSummary, page, totalPages int) (string, error) {
	// 构建上下文
	ctx := context.Background()

	// 计算分页信息
	pagination := Pagination{
		Page:       page,
		PageSize:   10, // 默认每页10条
		Total:      page * 10,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	}

	data := HomeContext{
		BaseContext: BaseContext{
			CurrentYear: 2026, // 应该从配置或时间获取
			ThemePath:   "/themes/default",
			AssetsPath:  "/themes/default/assets",
		},
		Posts:      posts,
		Pagination: pagination,
	}

	return s.engine.RenderHome(ctx, data)
}

// RenderPost 渲染文章页
func (s *service) RenderPost(post *PostDetail, categories []Taxonomy, tags []Taxonomy) (string, error) {
	ctx := context.Background()

	data := PostContext{
		BaseContext: BaseContext{
			CurrentYear: 2026,
			ThemePath:   "/themes/default",
			AssetsPath:  "/themes/default/assets",
		},
		Post:       *post,
		Categories: categories,
		Tags:       tags,
	}

	return s.engine.RenderPost(ctx, data)
}

// RenderPage 渲染页面
func (s *service) RenderPage(page *PageDetail) (string, error) {
	ctx := context.Background()

	data := PageContext{
		BaseContext: BaseContext{
			CurrentYear: 2026,
			ThemePath:   "/themes/default",
			AssetsPath:  "/themes/default/assets",
		},
		Page: *page,
	}

	return s.engine.RenderPage(ctx, data)
}

// InitThemes 初始化主题（启动时加载所有主题）
func (s *service) InitThemes() error {
	return s.engine.LoadAllThemes()
}

// Repository 主题数据访问接口（数据库操作）
type Repository interface {
	FindBySlug(ctx context.Context, slug string) (*Theme, error)
	FindActive(ctx context.Context) (*Theme, error)
	List(ctx context.Context) ([]*Theme, error)
	SetActive(ctx context.Context, id uint) error
	Create(ctx context.Context, theme *Theme) error
	Delete(ctx context.Context, id uint) error
}
