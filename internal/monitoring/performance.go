package monitoring

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/telman03/ai-backend-generator/internal/logging"
)

// PerformanceMonitor tracks application performance metrics
type PerformanceMonitor struct {
	mu                    sync.RWMutex
	operationMetrics      map[string]*OperationMetrics
	systemMetrics         *SystemMetrics
	alertThresholds       *AlertThresholds
	logger                *logging.Logger
	startTime             time.Time
	lastSystemMetricsTime time.Time
}

// OperationMetrics tracks metrics for a specific operation
type OperationMetrics struct {
	Name            string        `json:"name"`
	TotalCalls      int64         `json:"total_calls"`
	SuccessfulCalls int64         `json:"successful_calls"`
	FailedCalls     int64         `json:"failed_calls"`
	TotalDuration   time.Duration `json:"total_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	LastCall        time.Time     `json:"last_call"`
	ErrorRate       float64       `json:"error_rate"`
	AvgDuration     time.Duration `json:"avg_duration"`
}

// SystemMetrics tracks system-level performance metrics
type SystemMetrics struct {
	Timestamp         time.Time `json:"timestamp"`
	CPUUsage          float64   `json:"cpu_usage_percent"`
	MemoryUsage       uint64    `json:"memory_usage_bytes"`
	MemoryUsagePercent float64  `json:"memory_usage_percent"`
	GoroutineCount    int       `json:"goroutine_count"`
	HeapSize          uint64    `json:"heap_size_bytes"`
	HeapObjects       uint64    `json:"heap_objects"`
	GCPauses          uint64    `json:"gc_pauses_total"`
	NextGC            uint64    `json:"next_gc_bytes"`
}

// AlertThresholds defines thresholds for performance alerts
type AlertThresholds struct {
	MaxResponseTime   time.Duration `json:"max_response_time"`
	MaxErrorRate      float64       `json:"max_error_rate"`
	MaxMemoryUsage    float64       `json:"max_memory_usage_percent"`
	MaxCPUUsage       float64       `json:"max_cpu_usage_percent"`
	MaxGoroutines     int           `json:"max_goroutines"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Timestamp   time.Time   `json:"timestamp"`
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Message     string      `json:"message"`
	Value       interface{} `json:"value"`
	Threshold   interface{} `json:"threshold"`
	Operation   string      `json:"operation,omitempty"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		operationMetrics: make(map[string]*OperationMetrics),
		systemMetrics:    &SystemMetrics{},
		alertThresholds: &AlertThresholds{
			MaxResponseTime:   5 * time.Second,
			MaxErrorRate:      10.0, // 10%
			MaxMemoryUsage:    80.0, // 80%
			MaxCPUUsage:       80.0, // 80%
			MaxGoroutines:     1000,
		},
		logger:    logging.GetLogger().WithComponent("performance"),
		startTime: time.Now(),
	}
}

// RecordOperation records metrics for an operation
func (pm *PerformanceMonitor) RecordOperation(name string, duration time.Duration, success bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metrics, exists := pm.operationMetrics[name]
	if !exists {
		metrics = &OperationMetrics{
			Name:        name,
			MinDuration: duration,
			MaxDuration: duration,
		}
		pm.operationMetrics[name] = metrics
	}

	// Update metrics
	metrics.TotalCalls++
	metrics.TotalDuration += duration
	metrics.LastCall = time.Now()

	if success {
		metrics.SuccessfulCalls++
	} else {
		metrics.FailedCalls++
	}

	// Update min/max duration
	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}

	// Calculate derived metrics
	if metrics.TotalCalls > 0 {
		metrics.ErrorRate = float64(metrics.FailedCalls) / float64(metrics.TotalCalls) * 100
		metrics.AvgDuration = metrics.TotalDuration / time.Duration(metrics.TotalCalls)
	}

	// Check for performance alerts
	pm.checkOperationAlerts(metrics)
}

// UpdateSystemMetrics updates system-level metrics
func (pm *PerformanceMonitor) UpdateSystemMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pm.systemMetrics = &SystemMetrics{
		Timestamp:          time.Now(),
		MemoryUsage:        memStats.Alloc,
		MemoryUsagePercent: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		GoroutineCount:     runtime.NumGoroutine(),
		HeapSize:           memStats.HeapAlloc,
		HeapObjects:        memStats.HeapObjects,
		GCPauses:           uint64(memStats.NumGC),
		NextGC:             memStats.NextGC,
	}

	pm.lastSystemMetricsTime = time.Now()

	// Check for system alerts
	pm.checkSystemAlerts()
}

// GetOperationMetrics returns metrics for a specific operation
func (pm *PerformanceMonitor) GetOperationMetrics(name string) *OperationMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if metrics, exists := pm.operationMetrics[name]; exists {
		// Return a copy to avoid race conditions
		metricsCopy := *metrics
		return &metricsCopy
	}
	return nil
}

// GetAllOperationMetrics returns all operation metrics
func (pm *PerformanceMonitor) GetAllOperationMetrics() map[string]*OperationMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*OperationMetrics)
	for name, metrics := range pm.operationMetrics {
		metricsCopy := *metrics
		result[name] = &metricsCopy
	}
	return result
}

// GetSystemMetrics returns current system metrics
func (pm *PerformanceMonitor) GetSystemMetrics() *SystemMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Update system metrics if they're stale (older than 30 seconds)
	if time.Since(pm.lastSystemMetricsTime) > 30*time.Second {
		pm.mu.RUnlock()
		pm.UpdateSystemMetrics()
		pm.mu.RLock()
	}

	metricsCopy := *pm.systemMetrics
	return &metricsCopy
}

// GetPerformanceSummary returns a comprehensive performance summary
func (pm *PerformanceMonitor) GetPerformanceSummary() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := map[string]interface{}{
		"uptime":            time.Since(pm.startTime).String(),
		"system_metrics":    pm.systemMetrics,
		"operation_metrics": pm.operationMetrics,
		"alert_thresholds":  pm.alertThresholds,
		"total_operations":  len(pm.operationMetrics),
	}

	// Calculate aggregate metrics
	var totalCalls, totalSuccessful, totalFailed int64
	var totalDuration time.Duration

	for _, metrics := range pm.operationMetrics {
		totalCalls += metrics.TotalCalls
		totalSuccessful += metrics.SuccessfulCalls
		totalFailed += metrics.FailedCalls
		totalDuration += metrics.TotalDuration
	}

	summary["aggregate_metrics"] = map[string]interface{}{
		"total_calls":      totalCalls,
		"successful_calls": totalSuccessful,
		"failed_calls":     totalFailed,
		"overall_error_rate": func() float64 {
			if totalCalls > 0 {
				return float64(totalFailed) / float64(totalCalls) * 100
			}
			return 0
		}(),
		"average_duration": func() time.Duration {
			if totalCalls > 0 {
				return totalDuration / time.Duration(totalCalls)
			}
			return 0
		}(),
	}

	return summary
}

// checkOperationAlerts checks for operation-specific performance alerts
func (pm *PerformanceMonitor) checkOperationAlerts(metrics *OperationMetrics) {
	// Check response time alert
	if metrics.AvgDuration > pm.alertThresholds.MaxResponseTime {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			Type:      "SLOW_OPERATION",
			Severity:  "WARNING",
			Message:   "Operation average response time exceeds threshold",
			Value:     metrics.AvgDuration,
			Threshold: pm.alertThresholds.MaxResponseTime,
			Operation: metrics.Name,
		}
		pm.logAlert(alert)
	}

	// Check error rate alert
	if metrics.ErrorRate > pm.alertThresholds.MaxErrorRate {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			Type:      "HIGH_ERROR_RATE",
			Severity:  "WARNING",
			Message:   "Operation error rate exceeds threshold",
			Value:     metrics.ErrorRate,
			Threshold: pm.alertThresholds.MaxErrorRate,
			Operation: metrics.Name,
		}
		pm.logAlert(alert)
	}
}

// checkSystemAlerts checks for system-level performance alerts
func (pm *PerformanceMonitor) checkSystemAlerts() {
	// Check memory usage alert
	if pm.systemMetrics.MemoryUsagePercent > pm.alertThresholds.MaxMemoryUsage {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			Type:      "HIGH_MEMORY_USAGE",
			Severity:  "WARNING",
			Message:   "System memory usage exceeds threshold",
			Value:     pm.systemMetrics.MemoryUsagePercent,
			Threshold: pm.alertThresholds.MaxMemoryUsage,
		}
		pm.logAlert(alert)
	}

	// Check goroutine count alert
	if pm.systemMetrics.GoroutineCount > pm.alertThresholds.MaxGoroutines {
		alert := PerformanceAlert{
			Timestamp: time.Now(),
			Type:      "HIGH_GOROUTINE_COUNT",
			Severity:  "WARNING",
			Message:   "Goroutine count exceeds threshold",
			Value:     pm.systemMetrics.GoroutineCount,
			Threshold: pm.alertThresholds.MaxGoroutines,
		}
		pm.logAlert(alert)
	}
}

// logAlert logs a performance alert
func (pm *PerformanceMonitor) logAlert(alert PerformanceAlert) {
	pm.logger.Warn("Performance alert", map[string]interface{}{
		"alert_type":  alert.Type,
		"severity":    alert.Severity,
		"message":     alert.Message,
		"value":       alert.Value,
		"threshold":   alert.Threshold,
		"operation":   alert.Operation,
	})
}

// StartPeriodicMonitoring starts periodic system metrics collection
func (pm *PerformanceMonitor) StartPeriodicMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.UpdateSystemMetrics()
			
			// Log performance summary every 5 minutes
			if interval >= 5*time.Minute {
				pm.logPerformanceSummary()
			}
		}
	}
}

// logPerformanceSummary logs a periodic performance summary
func (pm *PerformanceMonitor) logPerformanceSummary() {
	summary := pm.GetPerformanceSummary()
	
	pm.logger.Info("Performance summary", map[string]interface{}{
		"uptime":           summary["uptime"],
		"total_operations": summary["total_operations"],
		"system_metrics":   summary["system_metrics"],
		"aggregate_metrics": summary["aggregate_metrics"],
	})
}

// Reset resets all performance metrics (useful for testing)
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.operationMetrics = make(map[string]*OperationMetrics)
	pm.systemMetrics = &SystemMetrics{}
	pm.startTime = time.Now()
	pm.lastSystemMetricsTime = time.Time{}
}

// SetAlertThresholds updates the alert thresholds
func (pm *PerformanceMonitor) SetAlertThresholds(thresholds *AlertThresholds) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.alertThresholds = thresholds
}

// Global performance monitor instance
var globalPerformanceMonitor *PerformanceMonitor

// InitPerformanceMonitor initializes the global performance monitor
func InitPerformanceMonitor() {
	globalPerformanceMonitor = NewPerformanceMonitor()
	
	// Start periodic monitoring in the background
	go func() {
		ctx := context.Background()
		globalPerformanceMonitor.StartPeriodicMonitoring(ctx, 1*time.Minute)
	}()
}

// GetPerformanceMonitor returns the global performance monitor
func GetPerformanceMonitor() *PerformanceMonitor {
	if globalPerformanceMonitor == nil {
		globalPerformanceMonitor = NewPerformanceMonitor()
	}
	return globalPerformanceMonitor
}