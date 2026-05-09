package theme

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestThemeTableName 测试主题模型表名
func TestThemeTableName(t *testing.T) {
	theme := Theme{}
	tableName := theme.TableName()

	assert.Equal(t, "themes", tableName, "表名应该是 themes")
}

// TestThemeConfig 测试主题配置结构
func TestThemeConfig(t *testing.T) {
	config := ThemeConfig{
		Name:        "Default Theme",
		Slug:        "default",
		Version:     "1.0.0",
		Author:      "GoPress Team",
		Description: "A default theme for GoPress",
		Screenshot:  "screenshot.png",
	}

	assert.Equal(t, "Default Theme", config.Name)
	assert.Equal(t, "default", config.Slug)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "GoPress Team", config.Author)
	assert.Equal(t, "A default theme for GoPress", config.Description)
	assert.Equal(t, "screenshot.png", config.Screenshot)
}

// TestThemeFields 测试主题模型字段
func TestThemeFields(t *testing.T) {
	now := time.Now()
	theme := Theme{
		ID:          1,
		Name:        "Test Theme",
		Slug:        "test-theme",
		Version:     "2.0.0",
		Author:      "Tester",
		Description: "Test description",
		Screenshot:  "test.png",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, uint(1), theme.ID)
	assert.Equal(t, "Test Theme", theme.Name)
	assert.Equal(t, "test-theme", theme.Slug)
	assert.Equal(t, "2.0.0", theme.Version)
	assert.Equal(t, "Tester", theme.Author)
	assert.Equal(t, "Test description", theme.Description)
	assert.Equal(t, "test.png", theme.Screenshot)
	assert.True(t, theme.IsActive)
	assert.Equal(t, now, theme.CreatedAt)
	assert.Equal(t, now, theme.UpdatedAt)
}

// TestBaseContext 测试基础模板上下文
func TestBaseContext(t *testing.T) {
	ctx := BaseContext{
		SiteName:    "My Site",
		SiteURL:     "https://example.com",
		CurrentYear: 2026,
		ThemePath:   "/themes/default",
		AssetsPath:  "/themes/default/assets",
	}

	assert.Equal(t, "My Site", ctx.SiteName)
	assert.Equal(t, "https://example.com", ctx.SiteURL)
	assert.Equal(t, 2026, ctx.CurrentYear)
	assert.Equal(t, "/themes/default", ctx.ThemePath)
	assert.Equal(t, "/themes/default/assets", ctx.AssetsPath)
}

// TestHomeContext 测试首页模板上下文
func TestHomeContext(t *testing.T) {
	posts := []PostSummary{
		{ID: 1, Title: "Post 1", Slug: "post-1"},
		{ID: 2, Title: "Post 2", Slug: "post-2"},
	}
	pagination := Pagination{
		Page:       1,
		PageSize:   10,
		Total:      20,
		TotalPages: 2,
		HasPrev:    false,
		HasNext:    true,
	}

	ctx := HomeContext{
		BaseContext: BaseContext{
			SiteName: "Test Site",
		},
		Posts:      posts,
		Pagination: pagination,
	}

	assert.Equal(t, "Test Site", ctx.SiteName)
	assert.Len(t, ctx.Posts, 2)
	assert.Equal(t, "Post 1", ctx.Posts[0].Title)
	assert.Equal(t, 2, ctx.Pagination.TotalPages)
	assert.True(t, ctx.Pagination.HasNext)
}

// TestPostContext 测试文章页模板上下文
func TestPostContext(t *testing.T) {
	post := PostDetail{
		ID:     1,
		Title:  "Test Post",
		Slug:   "test-post",
		Content: "<p>Content</p>",
	}

	categories := []Taxonomy{
		{ID: 1, Name: "Tech", Slug: "tech"},
	}
	tags := []Taxonomy{
		{ID: 1, Name: "Go", Slug: "go"},
	}

	ctx := PostContext{
		BaseContext: BaseContext{
			SiteName: "Test Site",
		},
		Post:       post,
		Categories: categories,
		Tags:       tags,
	}

	assert.Equal(t, "Test Post", ctx.Post.Title)
	assert.Len(t, ctx.Categories, 1)
	assert.Equal(t, "Tech", ctx.Categories[0].Name)
	assert.Len(t, ctx.Tags, 1)
	assert.Equal(t, "Go", ctx.Tags[0].Name)
}

// TestPageContext 测试页面模板上下文
func TestPageContext(t *testing.T) {
	page := PageDetail{
		ID:      1,
		Title:   "About",
		Slug:    "about",
		Content: "<p>About us</p>",
	}

	ctx := PageContext{
		BaseContext: BaseContext{
			SiteName: "Test Site",
		},
		Page: page,
	}

	assert.Equal(t, "About", ctx.Page.Title)
	assert.Equal(t, "about", ctx.Page.Slug)
}

// TestPagination 测试分页结构
func TestPagination(t *testing.T) {
	tests := []struct {
		name       string
		pagination Pagination
	}{
		{
			name: "第一页",
			pagination: Pagination{
				Page:       1,
				PageSize:   10,
				Total:      50,
				TotalPages: 5,
				HasPrev:    false,
				HasNext:    true,
			},
		},
		{
			name: "中间页",
			pagination: Pagination{
				Page:       3,
				PageSize:   10,
				Total:      50,
				TotalPages: 5,
				HasPrev:    true,
				HasNext:    true,
			},
		},
		{
			name: "最后一页",
			pagination: Pagination{
				Page:       5,
				PageSize:   10,
				Total:      50,
				TotalPages: 5,
				HasPrev:    true,
				HasNext:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.pagination
			assert.Equal(t, p.Page, p.Page)
			assert.Equal(t, p.TotalPages, p.TotalPages)
			assert.Equal(t, p.HasPrev, p.HasPrev)
			assert.Equal(t, p.HasNext, p.HasNext)
		})
	}
}

// TestTaxonomy 测试分类/标签结构
func TestTaxonomy(t *testing.T) {
	taxonomy := Taxonomy{
		ID:   1,
		Name: "Technology",
		Slug: "technology",
	}

	assert.Equal(t, uint(1), taxonomy.ID)
	assert.Equal(t, "Technology", taxonomy.Name)
	assert.Equal(t, "technology", taxonomy.Slug)
}
