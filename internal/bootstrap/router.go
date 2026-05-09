package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/config"
	"github.com/yourorg/gopress/internal/content"
	"github.com/yourorg/gopress/internal/media"
	"github.com/yourorg/gopress/internal/middleware"
	"github.com/yourorg/gopress/internal/taxonomy"
	"github.com/yourorg/gopress/internal/theme"
	"github.com/yourorg/gopress/internal/user"
	"github.com/yourorg/gopress/pkg/performance"
	"go.uber.org/zap"
)

type RouterDeps struct {
	UserHandler   *user.Handler
	TermHandler   *taxonomy.Handler
	PostHandler   *content.Handler
	MediaHandler  *media.Handler
	ThemeHandler  *theme.Handler
	Config        *config.Config
	Log           *zap.Logger
	PerfMonitor   *performance.Monitor
}

func NewRouter(cfg *config.Config, deps RouterDeps) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.RequestLogger(deps.Log))
	r.Use(middleware.CORS(cfg.CORS))
	r.Use(middleware.RequestID())
	r.Use(middleware.RateLimit(cfg.RateLimit))

	// 性能监控中间件
	if deps.PerfMonitor != nil {
		r.Use(deps.PerfMonitor.Middleware())
	}

	// Gzip 压缩中间件
	r.Use(middleware.Gzip(middleware.GzipConfig{
		CompressionLevel: 5, // Balanced compression
		MinSize:           1024,
		ExcludedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif", ".ico", ".pdf", ".zip", ".tar", ".gz", ".mp4", ".mp3", ".ogg", ".woff", ".woff2", ".ttf", ".eot"},
		ExcludedPaths:     []string{"/health", "/metrics"},
		ExcludedPrefixes:  []string{"/api/v1/admin"},
	}))

	// HTTP 缓存控制头
	r.Use(middleware.CacheControl(middleware.CacheConfig{
		StaticAssetsTTL: 24 * 3600 * 1000000000, // 24 hours in nanoseconds
		ApiTTL:           5 * 60 * 1000000000,     // 5 minutes in nanoseconds
		EnableETag:       true,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 性能指标端点
	if deps.PerfMonitor != nil {
		r.GET("/metrics", func(c *gin.Context) {
			metrics := deps.PerfMonitor.GetMetrics()
			endpoints := deps.PerfMonitor.GetEndpointMetrics()
			c.JSON(200, gin.H{
				"metrics":   metrics,
				"endpoints": endpoints,
			})
		})
	}

	// API v1 路由组
	v1 := r.Group("/api/v1")

	// 公开路由（无需认证）
	deps.UserHandler.RegisterPublicRoutes(v1)
	deps.TermHandler.RegisterPublicRoutes(v1)
	deps.PostHandler.RegisterPublicRoutes(v1)

	// 主题公开路由（首页、文章页、页面等）- 不需要 /api/v1 前缀
	public := r.Group("")
	deps.ThemeHandler.RegisterPublicRoutes(public)

	// 需认证路由（JWT）
	auth := v1.Group("")
	auth.Use(middleware.JWTAuth(cfg.JWT))
	deps.UserHandler.RegisterAuthRoutes(auth)
	deps.PostHandler.RegisterAuthRoutes(auth)
	deps.MediaHandler.RegisterAuthRoutes(auth)

	// 管理员路由（JWT + Admin 角色）
	admin := v1.Group("/admin")
	admin.Use(middleware.JWTAuth(cfg.JWT))
	admin.Use(middleware.RequireRole(user.RoleAdmin))
	deps.UserHandler.RegisterAdminRoutes(admin)
	deps.TermHandler.RegisterAdminRoutes(admin)
	deps.MediaHandler.RegisterAdminRoutes(admin)

	return r
}
