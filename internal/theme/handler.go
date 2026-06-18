package theme

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/content"
	"github.com/yourorg/gopress/internal/taxonomy"
	"github.com/yourorg/gopress/internal/user"
)

// Handler 主题处理器
type Handler struct {
	svc        *service
	contentSvc content.Service
	userRepo   user.Repository
	termRepo   taxonomy.Repository
	themesPath string
}

// NewHandler 创建主题处理器
func NewHandler(svc *service, contentSvc content.Service, userRepo user.Repository, termRepo taxonomy.Repository, themesPath string) *Handler {
	return &Handler{
		svc:        svc,
		contentSvc: contentSvc,
		userRepo:   userRepo,
		termRepo:   termRepo,
		themesPath: themesPath,
	}
}

// RegisterPublicRoutes 注册公开路由
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.Home)
	rg.GET("/post/:slug", h.Post)
	rg.GET("/page/:slug", h.Page)
	rg.GET("/theme/assets/*filepath", h.ThemeAssets)
}

// RegisterAdminRoutes 注册管理员路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/themes", h.ListThemes)
	rg.GET("/themes/:name", h.GetTheme)
	rg.POST("/themes/:name/activate", h.ActivateTheme)
}

// Home 首页处理
func (h *Handler) Home(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	filter := content.ListFilter{
		Type:     string(content.TypePost),
		Status:   string(content.StatusPublished),
		Page:     page,
		PageSize: pageSize,
	}

	posts, total, err := h.contentSvc.List(c.Request.Context(), filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load posts")
		return
	}

	summaries := make([]PostSummary, 0, len(posts))
	for _, p := range posts {
		summary := PostSummary{
			ID:         p.ID,
			Title:      p.Title,
			Slug:       p.Slug,
			Excerpt:    p.Excerpt,
			AuthorName: p.AuthorName,
			ViewCount:  p.ViewCount,
		}
		if p.PublishedAt != nil {
			summary.PublishedAt = *p.PublishedAt
		}
		summaries = append(summaries, summary)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	html, err := h.svc.RenderHome(summaries, page, totalPages)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render page: "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// Post 文章页处理
func (h *Handler) Post(c *gin.Context) {
	slug := c.Param("slug")

	posts, _, err := h.contentSvc.List(c.Request.Context(), content.ListFilter{
		Type:     string(content.TypePost),
		Status:   string(content.StatusPublished),
		Slug:     &slug,
		Page:     1,
		PageSize: 1,
	})
	if err != nil || len(posts) == 0 {
		c.String(http.StatusNotFound, "Post not found")
		return
	}

	p := posts[0]

	postDetail := PostDetail{
		ID:             p.ID,
		Title:          p.Title,
		Slug:           p.Slug,
		Content:        p.Content,
		RawContent:     p.Content,
		Excerpt:        p.Excerpt,
		AuthorID:       p.AuthorID,
		AuthorName:     p.AuthorName,
		ViewCount:      p.ViewCount,
		CommentAllowed: p.CommentAllowed,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
	if p.PublishedAt != nil {
		postDetail.PublishedAt = *p.PublishedAt
	}

	categories := make([]Taxonomy, 0)
	tags := make([]Taxonomy, 0)
	for _, term := range p.Terms {
		t := Taxonomy{ID: term.ID, Name: term.Name, Slug: term.Slug}
		if term.Taxonomy == "category" {
			categories = append(categories, t)
		} else if term.Taxonomy == "tag" {
			tags = append(tags, t)
		}
	}

	html, err := h.svc.RenderPost(&postDetail, categories, tags)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render post: "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// Page 页面处理
func (h *Handler) Page(c *gin.Context) {
	slug := c.Param("slug")

	pages, _, err := h.contentSvc.List(c.Request.Context(), content.ListFilter{
		Type:     string(content.TypePage),
		Status:   string(content.StatusPublished),
		Slug:     &slug,
		Page:     1,
		PageSize: 1,
	})
	if err != nil || len(pages) == 0 {
		c.String(http.StatusNotFound, "Page not found")
		return
	}

	p := pages[0]

	pageDetail := PageDetail{
		ID:         p.ID,
		Title:      p.Title,
		Slug:       p.Slug,
		Content:    p.Content,
		RawContent: p.Content,
		AuthorID:   p.AuthorID,
		AuthorName: p.AuthorName,
		ViewCount:  p.ViewCount,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	if p.PublishedAt != nil {
		pageDetail.PublishedAt = *p.PublishedAt
	}

	html, err := h.svc.RenderPage(&pageDetail)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render page: "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ThemeAssets 主题静态资源处理
func (h *Handler) ThemeAssets(c *gin.Context) {
	fp := c.Param("filepath")
	fp = strings.TrimPrefix(fp, "/")

	// 获取当前激活主题（默认 default）
	themeName := "default"

	// 构建预期的基础路径
	baseDir := filepath.Join(h.themesPath, themeName, "assets")

	// 清理路径并验证最终路径仍在 baseDir 内
	cleanPath := filepath.Join(baseDir, fp)
	if !strings.HasPrefix(cleanPath, baseDir) {
		c.String(http.StatusForbidden, "forbidden")
		return
	}

	c.File(cleanPath)
}

// ListThemes godoc
// @Summary      获取主题列表
// @Description  获取所有已安装的主题列表（管理员）
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ThemeListResponse
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/admin/themes [get]
func (h *Handler) ListThemes(c *gin.Context) {
	themes, err := h.svc.ListThemes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list themes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"themes": themes,
	})
}

// GetTheme godoc
// @Summary      获取主题详情
// @Description  获取指定主题的详细信息（管理员）
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "主题名称"
// @Success      200  {object}  Theme
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/admin/themes/{name} [get]
func (h *Handler) GetTheme(c *gin.Context) {
	name := c.Param("name")
	themes, err := h.svc.ListThemes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get theme"})
		return
	}
	for _, theme := range themes {
		if theme.Name == name || theme.Slug == name {
			c.JSON(http.StatusOK, theme)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "theme not found"})
}

// ActivateTheme godoc
// @Summary      激活主题
// @Description  激活指定的主题（管理员）
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "主题名称"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/admin/themes/{name}/activate [post]
func (h *Handler) ActivateTheme(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.SetActiveTheme(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "theme activated", "theme": name})
}

// 确保编译期引用 time 包（用于 PostDetail 等结构体）
var _ = time.Time{}
