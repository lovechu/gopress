package performance

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Metrics holds performance metrics
type Metrics struct {
	RequestCount    uint64
	RequestDuration int64 // milliseconds (aggregated)
	DBQueryCount    uint64
	DBQueryDuration int64 // milliseconds (aggregated)
	CacheHitCount   uint64
	CacheMissCount  uint64
	ErrorCount      uint64
	AvgLatency      int64 // milliseconds
	MaxLatency      int64 // milliseconds
	MinLatency      int64 // milliseconds
	ActiveRequests  int64
}

// EndpointMetrics holds metrics for a specific endpoint
type EndpointMetrics struct {
	Path        string
	Method      string
	Count       uint64
	AvgLatency  int64
	MaxLatency  int64
	P50Latency  int64
	P95Latency  int64
	P99Latency  int64
	ErrorCount  uint64
}

// Monitor provides performance monitoring functionality
type Monitor struct {
	mu           sync.RWMutex
	metrics      Metrics
	endpoints    map[string]*EndpointMetrics
	histograms   map[string]*Histogram
	logger       *zap.Logger
	enableDBMonitoring bool
}

// Histogram holds latency distribution data
type Histogram struct {
	mu       sync.Mutex
	count    uint64
	sum      int64
	min      int64
	max      int64
	buckets  []uint64
	bucketBoundaries []int64
}

// NewMonitor creates a new performance monitor
func NewMonitor(logger *zap.Logger) *Monitor {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	m := &Monitor{
		endpoints:    make(map[string]*EndpointMetrics),
		histograms:   make(map[string]*Histogram),
		logger:       logger,
		enableDBMonitoring: true,
	}

	// Initialize request latency histogram
	m.histograms["request"] = NewHistogram([]int64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000})
	m.histograms["db_query"] = NewHistogram([]int64{1, 5, 10, 25, 50, 100, 250, 500, 1000})

	return m
}

// NewHistogram creates a new histogram with bucket boundaries
func NewHistogram(boundaries []int64) *Histogram {
	return &Histogram{
		bucketBoundaries: boundaries,
		buckets:          make([]uint64, len(boundaries)+1),
		min:              int64(^uint64(0) >> 1), // Max int64
	}
}

// Record records a latency value
func (h *Histogram) Record(latencyMs int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.count++
	h.sum += latencyMs
	if latencyMs < h.min {
		h.min = latencyMs
	}
	if latencyMs > h.max {
		h.max = latencyMs
	}

	// Find bucket
	for i, boundary := range h.bucketBoundaries {
		if latencyMs <= boundary {
			h.buckets[i]++
			return
		}
	}
	h.buckets[len(h.buckets)-1]++
}

// RecordRequest records a request with its latency
func (m *Monitor) RecordRequest(path, method string, latencyMs int64, statusCode int, isError bool) {
	atomic.AddUint64(&m.metrics.RequestCount, 1)
	atomic.AddInt64(&m.metrics.RequestDuration, latencyMs)

	// Update max/min latency
	for {
		current := atomic.LoadInt64(&m.metrics.MaxLatency)
		if latencyMs <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.metrics.MaxLatency, current, latencyMs) {
			break
		}
	}

	for {
		current := atomic.LoadInt64(&m.metrics.MinLatency)
		if current == 0 || latencyMs >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.metrics.MinLatency, current, latencyMs) {
			break
		}
	}

	if isError {
		atomic.AddUint64(&m.metrics.ErrorCount, 1)
	}

	// Record in histogram
	if h, ok := m.histograms["request"]; ok {
		h.Record(latencyMs)
	}

	// Update endpoint metrics
	key := method + ":" + path
	m.mu.Lock()
	if ep, ok := m.endpoints[key]; ok {
		ep.Count++
		if latencyMs > ep.MaxLatency {
			ep.MaxLatency = latencyMs
		}
		if ep.Count > 1 {
			ep.AvgLatency = (ep.AvgLatency*(int64(ep.Count)-1) + latencyMs) / int64(ep.Count)
		} else {
			ep.AvgLatency = latencyMs
		}
		if isError {
			ep.ErrorCount++
		}
	} else {
		m.endpoints[key] = &EndpointMetrics{
			Path:       path,
			Method:     method,
			Count:      1,
			AvgLatency: latencyMs,
			MaxLatency: latencyMs,
			ErrorCount: 0,
		}
		if isError {
			m.endpoints[key].ErrorCount = 1
		}
	}
	m.mu.Unlock()
}

// RecordDBQuery records a database query with its latency
func (m *Monitor) RecordDBQuery(queryType string, latencyMs int64) {
	if !m.enableDBMonitoring {
		return
	}

	atomic.AddUint64(&m.metrics.DBQueryCount, 1)
	atomic.AddInt64(&m.metrics.DBQueryDuration, latencyMs)

	// Record in histogram
	if h, ok := m.histograms["db_query"]; ok {
		h.Record(latencyMs)
	}

	m.logger.Debug("db query recorded",
		zap.String("type", queryType),
		zap.Int64("latency_ms", latencyMs),
	)
}

// RecordCacheHit records a cache hit
func (m *Monitor) RecordCacheHit() {
	atomic.AddUint64(&m.metrics.CacheHitCount, 1)
}

// RecordCacheMiss records a cache miss
func (m *Monitor) RecordCacheMiss() {
	atomic.AddUint64(&m.metrics.CacheMissCount, 1)
}

// IncrementActiveRequests increments the active request counter
func (m *Monitor) IncrementActiveRequests() {
	atomic.AddInt64(&m.metrics.ActiveRequests, 1)
}

// DecrementActiveRequests decrements the active request counter
func (m *Monitor) DecrementActiveRequests() {
	atomic.AddInt64(&m.metrics.ActiveRequests, -1)
}

// GetMetrics returns current metrics
func (m *Monitor) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := Metrics{
		RequestCount:   atomic.LoadUint64(&m.metrics.RequestCount),
		DBQueryCount:   atomic.LoadUint64(&m.metrics.DBQueryCount),
		CacheHitCount:  atomic.LoadUint64(&m.metrics.CacheHitCount),
		CacheMissCount: atomic.LoadUint64(&m.metrics.CacheMissCount),
		ErrorCount:     atomic.LoadUint64(&m.metrics.ErrorCount),
		ActiveRequests: atomic.LoadInt64(&m.metrics.ActiveRequests),
	}

	if metrics.RequestCount > 0 {
		metrics.RequestDuration = atomic.LoadInt64(&m.metrics.RequestDuration)
		metrics.AvgLatency = metrics.RequestDuration / int64(metrics.RequestCount)
	}
	if metrics.DBQueryCount > 0 {
		metrics.DBQueryDuration = atomic.LoadInt64(&m.metrics.DBQueryDuration)
	}
	metrics.MaxLatency = atomic.LoadInt64(&m.metrics.MaxLatency)
	metrics.MinLatency = atomic.LoadInt64(&m.metrics.MinLatency)

	return metrics
}

// GetEndpointMetrics returns metrics for all endpoints
func (m *Monitor) GetEndpointMetrics() []EndpointMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]EndpointMetrics, 0, len(m.endpoints))
	for _, ep := range m.endpoints {
		result = append(result, *ep)
	}
	return result
}

// GetHistogram returns histogram for a key
func (m *Monitor) GetHistogram(key string) *Histogram {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.histograms[key]
}

// Reset resets all metrics
func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreUint64(&m.metrics.RequestCount, 0)
	atomic.StoreInt64(&m.metrics.RequestDuration, 0)
	atomic.StoreUint64(&m.metrics.DBQueryCount, 0)
	atomic.StoreInt64(&m.metrics.DBQueryDuration, 0)
	atomic.StoreUint64(&m.metrics.CacheHitCount, 0)
	atomic.StoreUint64(&m.metrics.CacheMissCount, 0)
	atomic.StoreUint64(&m.metrics.ErrorCount, 0)
	atomic.StoreInt64(&m.metrics.AvgLatency, 0)
	atomic.StoreInt64(&m.metrics.MaxLatency, 0)
	atomic.StoreInt64(&m.metrics.MinLatency, 0)
	atomic.StoreInt64(&m.metrics.ActiveRequests, 0)

	// Reset histograms
	for _, h := range m.histograms {
		h.Reset()
	}
}

// Reset resets a histogram
func (h *Histogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count = 0
	h.sum = 0
	h.min = int64(^uint64(0) >> 1)
	h.max = 0
	for i := range h.buckets {
		h.buckets[i] = 0
	}
}

// Middleware returns a Gin middleware for request monitoring
func (m *Monitor) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		m.IncrementActiveRequests()
		defer m.DecrementActiveRequests()

		c.Next()

		latencyMs := time.Since(start).Milliseconds()
		statusCode := c.Writer.Status()
		isError := statusCode >= 400

		m.RecordRequest(c.FullPath(), c.Request.Method, latencyMs, statusCode, isError)

		// Log slow requests
		if latencyMs > 1000 {
			m.logger.Warn("slow request detected",
				zap.String("path", c.FullPath()),
				zap.String("method", c.Request.Method),
				zap.Int64("latency_ms", latencyMs),
				zap.Int("status", statusCode),
			)
		}
	}
}

// StartPeriodicReport starts a goroutine that periodically logs metrics
func (m *Monitor) StartPeriodicReport(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics := m.GetMetrics()
				m.logger.Info("performance metrics",
					zap.Uint64("request_count", metrics.RequestCount),
					zap.Int64("active_requests", metrics.ActiveRequests),
					zap.Int64("avg_latency_ms", metrics.AvgLatency),
					zap.Int64("max_latency_ms", metrics.MaxLatency),
					zap.Uint64("db_query_count", metrics.DBQueryCount),
					zap.Uint64("cache_hits", metrics.CacheHitCount),
					zap.Uint64("cache_misses", metrics.CacheMissCount),
				)
			}
		}
	}()
}

// DBQueryTimer is a helper to time database queries
type DBQueryTimer struct {
	monitor   *Monitor
	queryType string
	start     time.Time
}

// StartDBQuery starts timing a database query
func (m *Monitor) StartDBQuery(queryType string) *DBQueryTimer {
	return &DBQueryTimer{
		monitor:   m,
		queryType: queryType,
		start:     time.Now(),
	}
}

// End ends the timing and records the result
func (t *DBQueryTimer) End() {
	latencyMs := time.Since(t.start).Milliseconds()
	t.monitor.RecordDBQuery(t.queryType, latencyMs)
}

// WithDBTiming wraps a function with database query timing
func (m *Monitor) WithDBTiming(queryType string, fn func() error) error {
	timer := m.StartDBQuery(queryType)
	defer timer.End()
	return fn()
}
