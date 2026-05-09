package middleware

import (
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheConfig holds cache header configuration
type CacheConfig struct {
	// StaticAssetsTTL is the TTL for static assets (CSS, JS, images, fonts)
	StaticAssetsTTL time.Duration
	// ApiTTL is the TTL for API responses
	ApiTTL time.Duration
	// Enable ETags
	EnableETag bool
	// CustomPaths defines custom cache settings for specific paths
	CustomPaths map[string]CachePathConfig
}

// CachePathConfig defines cache settings for a specific path
type CachePathConfig struct {
	TTL            time.Duration
	CacheControl   string
	Private        bool
	NoCache        bool
	NoStore        bool
	MustRevalidate bool
}

// defaultCacheConfig returns default cache configuration
func defaultCacheConfig() CacheConfig {
	return CacheConfig{
		StaticAssetsTTL: 24 * time.Hour,
		ApiTTL:          5 * time.Minute,
		EnableETag:      true,
		CustomPaths:     make(map[string]CachePathConfig),
	}
}

// Common cache durations
const (
	CacheNone     = 0
	Cache1Min     = 1 * time.Minute
	Cache5Min     = 5 * time.Minute
	Cache15Min    = 15 * time.Minute
	Cache30Min    = 30 * time.Minute
	Cache1Hour    = 1 * time.Hour
	Cache6Hours   = 6 * time.Hour
	Cache12Hours  = 12 * time.Hour
	Cache1Day     = 24 * time.Hour
	Cache7Days    = 7 * 24 * time.Hour
	Cache30Days   = 30 * 24 * time.Hour
	Cache1Year    = 365 * 24 * time.Hour
)

// Static file extensions
var staticExtensions = map[string]bool{
	".css":  true,
	".js":   true,
	".json": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".avif": true,
	".ico":  true,
	".svg":  true,
	".woff": true,
	".woff2": true,
	".ttf":  true,
	".otf":  true,
	".eot":  true,
	".map":  true,
	".gz":   true,
}

// Static paths that should be cached
var staticPaths = map[string]bool{
	"/assets/":    true,
	"/static/":   true,
	"/themes/":    true,
	"/uploads/":   true,
	"/favicon.ico": true,
}

// CacheControl returns a Gin middleware that adds cache headers
func CacheControl(cfg ...CacheConfig) gin.HandlerFunc {
	config := defaultCacheConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		ext := strings.ToLower(path.Ext(urlPath))

		// Check for custom path configuration
		for pattern, customCfg := range config.CustomPaths {
			if strings.HasPrefix(urlPath, pattern) {
				setCacheHeaders(c, customCfg, config.EnableETag)
				return
			}
		}

		// Determine cache settings based on path
		if isStaticAsset(urlPath, ext) {
			// Static assets - long cache with revalidation
			setStaticCacheHeaders(c, config.StaticAssetsTTL, config.EnableETag)
		} else if strings.HasPrefix(urlPath, "/api/") {
			// API responses - short cache or no cache
			setApiCacheHeaders(c, config.ApiTTL, config.EnableETag)
		} else if strings.HasPrefix(urlPath, "/sitemap") {
			// Sitemaps - moderate cache
			setCacheHeaders(c, CachePathConfig{TTL: Cache1Hour}, config.EnableETag)
		} else {
			// HTML pages - minimal cache
			setCacheHeaders(c, CachePathConfig{TTL: Cache5Min}, config.EnableETag)
		}
	}
}

// isStaticAsset checks if the path is a static asset
func isStaticAsset(urlPath, ext string) bool {
	// Check by extension
	if staticExtensions[ext] {
		return true
	}
	// Check by path prefix
	for staticPath := range staticPaths {
		if strings.HasPrefix(urlPath, staticPath) {
			return true
		}
	}
	return false
}

// setStaticCacheHeaders sets cache headers for static assets
func setStaticCacheHeaders(c *gin.Context, ttl time.Duration, enableETag bool) {
	maxAge := int(ttl.Seconds())

	c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge)+", immutable")
	c.Header("Vary", "Accept-Encoding")

	if enableETag {
		setETag(c)
	}

	// Set expiry header
	c.Header("Expires", time.Now().Add(ttl).Format(http.TimeFormat))
}

// setApiCacheHeaders sets cache headers for API responses
func setApiCacheHeaders(c *gin.Context, ttl time.Duration, enableETag bool) {
	maxAge := int(ttl.Seconds())

	c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
	c.Header("Vary", "Accept-Encoding")

	if enableETag {
		setETag(c)
	}
}

// setCacheHeaders sets cache headers based on configuration
func setCacheHeaders(c *gin.Context, cfg CachePathConfig, enableETag bool) {
	headers := []string{}

	if cfg.NoCache {
		headers = append(headers, "no-cache")
	} else if cfg.NoStore {
		headers = append(headers, "no-store")
	} else if cfg.Private {
		headers = append(headers, "private")
	} else {
		headers = append(headers, "public")
	}

	if cfg.TTL > 0 {
		headers = append(headers, "max-age="+strconv.Itoa(int(cfg.TTL.Seconds())))
	}

	if cfg.MustRevalidate {
		headers = append(headers, "must-revalidate")
	}

	c.Header("Cache-Control", strings.Join(headers, ", "))
	c.Header("Vary", "Accept-Encoding")

	if enableETag && !cfg.NoStore && !cfg.NoCache {
		setETag(c)
	}

	if cfg.TTL > 0 && !cfg.NoStore {
		c.Header("Expires", time.Now().Add(cfg.TTL).Format(http.TimeFormat))
	}
}

// setETag sets an ETag header based on response content
func setETag(c *gin.Context) {
	// Skip if ETag is already set
	if c.Writer.Header().Get("ETag") != "" {
		return
	}

	// Use response size and timestamp as a simple ETag
	// In production, you'd want to use content hash
	contentLength := c.Writer.Size()
	if contentLength > 0 {
		etag := generateETag(c.Request.URL.Path, contentLength)
		c.Header("ETag", etag)

		// Check If-None-Match header
		ifNoneMatch := c.GetHeader("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			c.AbortWithStatus(http.StatusNotModified)
		}
	}
}

// generateETag generates a simple ETag
func generateETag(path string, size int) string {
	return `"` + strconv.Itoa(size) + "-" + strconv.Itoa(len(path)) + `"`
}

// NoCache returns a middleware that disables all caching
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Header("Vary", "*")
		c.Next()
	}
}

// StaticCache returns a middleware for static file caching
func StaticCache(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		maxAge := int(ttl.Seconds())
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
		c.Header("Vary", "Accept-Encoding")
		c.Header("Expires", time.Now().Add(ttl).Format(http.TimeFormat))
		c.Next()
	}
}

// ResourceCache returns a middleware for specific resource caching
func ResourceCache(ttl time.Duration, isPublic bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		maxAge := int(ttl.Seconds())
		cacheType := "private"
		if isPublic {
			cacheType = "public"
		}
		c.Header("Cache-Control", cacheType+", max-age="+strconv.Itoa(maxAge))
		c.Header("Vary", "Accept-Encoding")
		c.Header("Expires", time.Now().Add(ttl).Format(http.TimeFormat))
		c.Next()
	}
}
