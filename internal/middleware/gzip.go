package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// GzipConfig holds Gzip middleware configuration
type GzipConfig struct {
	// CompressionLevel: 1-9, default gzip.BestSpeed (1)
	// Use gzip.BestCompression (9) for better compression but slower
	CompressionLevel int
	// MinSize is the minimum response size to compress (bytes)
	// Responses smaller than this won't be compressed
	MinSize int
	// ExcludedExtensions are file extensions to exclude from compression
	ExcludedExtensions []string
	// ExcludedPaths are URL paths to exclude from compression
	ExcludedPaths []string
	// ExcludedPrefixes are URL path prefixes to exclude from compression
	ExcludedPrefixes []string
}

// defaultGzipConfig returns default Gzip configuration
func defaultGzipConfig() GzipConfig {
	return GzipConfig{
		CompressionLevel:   gzip.BestSpeed,
		MinSize:            1024, // 1KB minimum
		ExcludedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif", ".ico", ".pdf", ".zip", ".tar", ".gz", ".mp4", ".mp3", ".ogg", ".woff", ".woff2", ".ttf", ".eot"},
		ExcludedPaths:      []string{"/health", "/metrics"},
		ExcludedPrefixes:   []string{"/api/v1/admin"},
	}
}

// Gzip returns a Gzip compression middleware
func Gzip(cfg ...GzipConfig) gin.HandlerFunc {
	config := defaultGzipConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	// Convert extensions to lowercase for comparison
	excludedExts := make(map[string]bool)
	for _, ext := range config.ExcludedExtensions {
		excludedExts[strings.ToLower(ext)] = true
	}

	// Create excluded paths map for O(1) lookup
	excludedPaths := make(map[string]bool)
	for _, path := range config.ExcludedPaths {
		excludedPaths[path] = true
	}

	level := config.CompressionLevel
	if level < gzip.BestSpeed || level > gzip.BestCompression {
		level = gzip.BestSpeed
	}

	return func(c *gin.Context) {
		// Check excluded paths
		if excludedPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Check excluded prefixes
		for _, prefix := range config.ExcludedPrefixes {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				c.Next()
				return
			}
		}

		// Check excluded extensions
		path := c.Request.URL.Path
		for ext := range excludedExts {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				c.Next()
				return
			}
		}

		// Check if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Create gzip writer
		gz, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			c.Next()
			return
		}

		// Set headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Remove Content-Length as it changes with compression
		c.Writer.Header().Del("Content-Length")

		// Create a wrapper to track the size
		gzWriter := &gzipResponseWriter{
			Writer: gz,
			ResponseWriter: c.Writer,
		}

		c.Writer = gzWriter
		c.Next()

		// Flush and close gzip writer
		gz.Close()
	}
}

// gzipResponseWriter wraps gin.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer *gzip.Writer
	size   int
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	n, err := w.Writer.Write(data)
	w.size += n
	return n, err
}

func (w *gzipResponseWriter) Size() int {
	return w.size
}

// ShouldCompress checks if the request should be compressed
func ShouldCompress(c *gin.Context) bool {
	// Check Accept-Encoding header
	acceptEncoding := c.GetHeader("Accept-Encoding")
	if !strings.Contains(acceptEncoding, "gzip") {
		return false
	}

	// Don't compress if request method is not GET or POST
	method := c.Request.Method
	if method != http.MethodGet && method != http.MethodPost {
		return false
	}

	return true
}

// GzipWriter wraps gin.ResponseWriter with Gzip compression support
type GzipWriter struct {
	gin.ResponseWriter
	writer   io.Writer
	gzWriter *gzip.Writer
	pool     sync.Pool
}

// NewGzipWriter creates a new Gzip writer
func NewGzipWriter(w io.Writer, level int) *GzipWriter {
	gz, _ := gzip.NewWriterLevel(w, level)
	return &GzipWriter{
		writer:   w,
		gzWriter: gz,
	}
}

// Write writes data through gzip
func (w *GzipWriter) Write(data []byte) (int, error) {
	return w.gzWriter.Write(data)
}

// Flush flushes the gzip writer
func (w *GzipWriter) Flush() error {
	return w.gzWriter.Flush()
}

// Close closes the gzip writer
func (w *GzipWriter) Close() error {
	return w.gzWriter.Close()
}
