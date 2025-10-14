package monitoring

import (
	"sync"
	"time"

	"github.com/telman03/gocraft-backend/internal/logging"
)

// MetricsCollector collects and tracks application metrics
type MetricsCollector struct {
	mu                    sync.RWMutex
	requestCount          map[string]int64
	errorCount            map[string]int64
	responseTimeSum       map[string]int64
	responseTimeCount     map[string]int64
	activeConnections     int64
	totalConnections      int64
	securityEvents        []SecurityEvent
	performanceMetrics    []PerformanceMetric
	logger                *logging.Logger
	startTime             time.Time
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Event       string    `json:"event"`
	Severity    string    `json:"severity"`
	UserID      *uint     `json:"user_id,omitempty"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Details     string    `json:"details"`
	RequestID   string    `json:"request_id,omitempty"`
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	Timestamp   time.Time              `json:"timestamp"`
	Operation   string                 `json:"operation"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	UserID      *uint                  `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MetricsSummary provides a summary of collected metrics
type MetricsSummary struct {
	Uptime              time.Duration            `json:"uptime"`
	TotalRequests       int64                    `json:"total_requests"`
	TotalErrors         int64                    `json:"total_errors"`
	ErrorRate           float64                  `json:"error_rate"`
	AverageResponseTime float64                  `json:"average_response_time_ms"`
	ActiveConnections   int64                    `json:"active_connections"`
	TotalConnections    int64                    `json:"total_connections"`
	RequestsByEndpoint  map[string]int64         `json:"requests_by_endpoint"`
	ErrorsByType        map[string]int64         `json:"errors_by_type"`
	ResponseTimes       map[string]float64       `json:"response_times_by_endpoint"`
	SecurityEvents      int                      `json:"security_events_count"`
	RecentSecurityEvents []SecurityEvent         `json:"recent_security_events,omitempty"`
	PerformanceMetrics  []PerformanceMetric      `json:"recent_performance_metrics,omitempty"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestCount:      make(map[string]int64),
		errorCount:        make(map[string]int64),
		responseTimeSum:   make(map[string]int64),
		responseTimeCount: make(map[string]int64),
		securityEvents:    make([]SecurityEvent, 0),
		performanceMetrics: make([]PerformanceMetric, 0),
		logger:            logging.GetLogger().WithComponent("metrics"),
		startTime:         time.Now(),
	}
}

// RecordRequest records a request metric
func (mc *MetricsCollector) RecordRequest(endpoint string, duration time.Duration, statusCode int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requestCount[endpoint]++
	
	durationMs := duration.Milliseconds()
	mc.responseTimeSum[endpoint] += durationMs
	mc.responseTimeCount[endpoint]++

	// Record errors (4xx and 5xx status codes)
	if statusCode >= 400 {
		mc.errorCount[endpoint]++
	}

	// Log performance metrics for slow requests (>1 second)
	if duration > time.Second {
		mc.logger.Warn("Slow request detected", map[string]interface{}{
			"endpoint":     endpoint,
			"duration_ms":  durationMs,
			"status_code":  statusCode,
		})
	}
}

// RecordError records an error metric
func (mc *MetricsCollector) RecordError(errorType string, endpoint string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := errorType + ":" + endpoint
	mc.errorCount[key]++
}

// RecordSecurityEvent records a security event
func (mc *MetricsCollector) RecordSecurityEvent(event SecurityEvent) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.securityEvents = append(mc.securityEvents, event)

	// Keep only the last 100 security events
	if len(mc.securityEvents) > 100 {
		mc.securityEvents = mc.securityEvents[len(mc.securityEvents)-100:]
	}

	// Log high severity events immediately
	if event.Severity == "high" || event.Severity == "critical" {
		mc.logger.Error("High severity security event", nil, map[string]interface{}{
			"event":      event.Event,
			"severity":   event.Severity,
			"user_id":    event.UserID,
			"ip":         event.IP,
			"details":    event.Details,
			"request_id": event.RequestID,
		})
	}
}

// RecordPerformanceMetric records a performance metric
func (mc *MetricsCollector) RecordPerformanceMetric(metric PerformanceMetric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.performanceMetrics = append(mc.performanceMetrics, metric)

	// Keep only the last 1000 performance metrics
	if len(mc.performanceMetrics) > 1000 {
		mc.performanceMetrics = mc.performanceMetrics[len(mc.performanceMetrics)-1000:]
	}

	// Log slow operations
	if metric.Duration > 5*time.Second {
		mc.logger.Warn("Slow operation detected", map[string]interface{}{
			"operation":   metric.Operation,
			"duration_ms": metric.Duration.Milliseconds(),
			"success":     metric.Success,
			"user_id":     metric.UserID,
			"metadata":    metric.Metadata,
		})
	}
}

// IncrementActiveConnections increments the active connections counter
func (mc *MetricsCollector) IncrementActiveConnections() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.activeConnections++
	mc.totalConnections++
}

// DecrementActiveConnections decrements the active connections counter
func (mc *MetricsCollector) DecrementActiveConnections() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.activeConnections > 0 {
		mc.activeConnections--
	}
}

// GetSummary returns a summary of all collected metrics
func (mc *MetricsCollector) GetSummary() MetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	summary := MetricsSummary{
		Uptime:              time.Since(mc.startTime),
		ActiveConnections:   mc.activeConnections,
		TotalConnections:    mc.totalConnections,
		RequestsByEndpoint:  make(map[string]int64),
		ErrorsByType:        make(map[string]int64),
		ResponseTimes:       make(map[string]float64),
		SecurityEvents:      len(mc.securityEvents),
	}

	// Calculate total requests and errors
	var totalRequests, totalErrors int64
	var totalResponseTime int64

	for endpoint, count := range mc.requestCount {
		summary.RequestsByEndpoint[endpoint] = count
		totalRequests += count

		// Calculate average response time for this endpoint
		if mc.responseTimeCount[endpoint] > 0 {
			avgTime := float64(mc.responseTimeSum[endpoint]) / float64(mc.responseTimeCount[endpoint])
			summary.ResponseTimes[endpoint] = avgTime
			totalResponseTime += mc.responseTimeSum[endpoint]
		}
	}

	for errorType, count := range mc.errorCount {
		summary.ErrorsByType[errorType] = count
		totalErrors += count
	}

	summary.TotalRequests = totalRequests
	summary.TotalErrors = totalErrors

	// Calculate error rate
	if totalRequests > 0 {
		summary.ErrorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	// Calculate average response time
	if totalRequests > 0 {
		summary.AverageResponseTime = float64(totalResponseTime) / float64(totalRequests)
	}

	// Include recent security events (last 10)
	if len(mc.securityEvents) > 0 {
		start := 0
		if len(mc.securityEvents) > 10 {
			start = len(mc.securityEvents) - 10
		}
		summary.RecentSecurityEvents = mc.securityEvents[start:]
	}

	// Include recent performance metrics (last 20)
	if len(mc.performanceMetrics) > 0 {
		start := 0
		if len(mc.performanceMetrics) > 20 {
			start = len(mc.performanceMetrics) - 20
		}
		summary.PerformanceMetrics = mc.performanceMetrics[start:]
	}

	return summary
}

// GetSecurityEvents returns all security events within a time range
func (mc *MetricsCollector) GetSecurityEvents(since time.Time) []SecurityEvent {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var events []SecurityEvent
	for _, event := range mc.securityEvents {
		if event.Timestamp.After(since) {
			events = append(events, event)
		}
	}

	return events
}

// GetPerformanceMetrics returns performance metrics within a time range
func (mc *MetricsCollector) GetPerformanceMetrics(since time.Time, operation string) []PerformanceMetric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var metrics []PerformanceMetric
	for _, metric := range mc.performanceMetrics {
		if metric.Timestamp.After(since) && (operation == "" || metric.Operation == operation) {
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

// Reset resets all metrics (useful for testing)
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requestCount = make(map[string]int64)
	mc.errorCount = make(map[string]int64)
	mc.responseTimeSum = make(map[string]int64)
	mc.responseTimeCount = make(map[string]int64)
	mc.activeConnections = 0
	mc.totalConnections = 0
	mc.securityEvents = make([]SecurityEvent, 0)
	mc.performanceMetrics = make([]PerformanceMetric, 0)
	mc.startTime = time.Now()
}

// LogMetricsSummary logs a summary of metrics periodically
func (mc *MetricsCollector) LogMetricsSummary() {
	summary := mc.GetSummary()
	
	mc.logger.Info("Metrics summary", map[string]interface{}{
		"uptime":                summary.Uptime.String(),
		"total_requests":        summary.TotalRequests,
		"total_errors":          summary.TotalErrors,
		"error_rate":            summary.ErrorRate,
		"avg_response_time_ms":  summary.AverageResponseTime,
		"active_connections":    summary.ActiveConnections,
		"total_connections":     summary.TotalConnections,
		"security_events_count": summary.SecurityEvents,
	})
}

// Global metrics collector instance
var globalMetrics *MetricsCollector

// InitMetrics initializes the global metrics collector
func InitMetrics() {
	globalMetrics = NewMetricsCollector()
	
	// Start periodic metrics logging (every 5 minutes)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			globalMetrics.LogMetricsSummary()
		}
	}()
}

// GetMetrics returns the global metrics collector
func GetMetrics() *MetricsCollector {
	if globalMetrics == nil {
		globalMetrics = NewMetricsCollector()
	}
	return globalMetrics
}