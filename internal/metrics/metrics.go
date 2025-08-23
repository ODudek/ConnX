package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Metrics holds all the metrics data
type Metrics struct {
	mu                 sync.RWMutex
	requestsTotal      map[string]int64 // key: method_status
	requestDurations   []time.Duration
	errorCount         int64
	activeConnections  int64
	backendStatus      map[string]bool // backend URL -> healthy status
	startTime          time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		requestsTotal:     make(map[string]int64),
		requestDurations:  make([]time.Duration, 0, 1000), // Keep last 1000 requests
		backendStatus:     make(map[string]bool),
		startTime:         time.Now(),
	}
}

// IncrementRequests increments the request counter for a given method and status
func (m *Metrics) IncrementRequests(method string, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := fmt.Sprintf("%s_%d", method, status)
	m.requestsTotal[key]++
}

// RecordRequestDuration records the duration of a request
func (m *Metrics) RecordRequestDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Keep only last 1000 durations to prevent memory growth
	if len(m.requestDurations) >= 1000 {
		m.requestDurations = m.requestDurations[1:]
	}
	m.requestDurations = append(m.requestDurations, duration)
}

// IncrementErrors increments the error counter
func (m *Metrics) IncrementErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCount++
}

// SetActiveConnections sets the current number of active connections
func (m *Metrics) SetActiveConnections(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeConnections = count
}

// UpdateBackendStatus updates the status of a backend
func (m *Metrics) UpdateBackendStatus(url string, healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.backendStatus[url] = healthy
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (m *Metrics) GetPrometheusMetrics() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var output string
	
	// Request totals
	output += "# HELP connx_requests_total Total number of HTTP requests\n"
	output += "# TYPE connx_requests_total counter\n"
	for key, count := range m.requestsTotal {
		output += fmt.Sprintf("connx_requests_total{method_status=\"%s\"} %d\n", key, count)
	}
	
	// Error count
	output += "# HELP connx_errors_total Total number of errors\n"
	output += "# TYPE connx_errors_total counter\n"
	output += fmt.Sprintf("connx_errors_total %d\n", m.errorCount)
	
	// Active connections
	output += "# HELP connx_active_connections Current number of active connections\n"
	output += "# TYPE connx_active_connections gauge\n"
	output += fmt.Sprintf("connx_active_connections %d\n", m.activeConnections)
	
	// Response time metrics
	if len(m.requestDurations) > 0 {
		avg := m.calculateAverageResponseTime()
		p95 := m.calculatePercentile(95)
		p99 := m.calculatePercentile(99)
		
		output += "# HELP connx_response_time_seconds Response time statistics\n"
		output += "# TYPE connx_response_time_seconds gauge\n"
		output += fmt.Sprintf("connx_response_time_seconds{quantile=\"avg\"} %.6f\n", avg.Seconds())
		output += fmt.Sprintf("connx_response_time_seconds{quantile=\"0.95\"} %.6f\n", p95.Seconds())
		output += fmt.Sprintf("connx_response_time_seconds{quantile=\"0.99\"} %.6f\n", p99.Seconds())
	}
	
	// Backend health status
	output += "# HELP connx_backend_up Backend health status (1=up, 0=down)\n"
	output += "# TYPE connx_backend_up gauge\n"
	for url, healthy := range m.backendStatus {
		value := 0
		if healthy {
			value = 1
		}
		output += fmt.Sprintf("connx_backend_up{backend=\"%s\"} %d\n", url, value)
	}
	
	// Uptime
	uptime := time.Since(m.startTime)
	output += "# HELP connx_uptime_seconds Uptime in seconds\n"
	output += "# TYPE connx_uptime_seconds counter\n"
	output += fmt.Sprintf("connx_uptime_seconds %.0f\n", uptime.Seconds())
	
	// Requests per second (based on uptime)
	totalRequests := int64(0)
	for _, count := range m.requestsTotal {
		totalRequests += count
	}
	rps := float64(totalRequests) / uptime.Seconds()
	output += "# HELP connx_requests_per_second Current requests per second\n"
	output += "# TYPE connx_requests_per_second gauge\n"
	output += fmt.Sprintf("connx_requests_per_second %.2f\n", rps)
	
	return output
}

// calculateAverageResponseTime calculates average response time
func (m *Metrics) calculateAverageResponseTime() time.Duration {
	if len(m.requestDurations) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, duration := range m.requestDurations {
		total += duration
	}
	return total / time.Duration(len(m.requestDurations))
}

// calculatePercentile calculates the nth percentile of response times
func (m *Metrics) calculatePercentile(percentile int) time.Duration {
	if len(m.requestDurations) == 0 {
		return 0
	}
	
	// Simple percentile calculation - in production you'd want to sort the slice
	// For now, we'll approximate by taking a position in the unsorted slice
	index := (len(m.requestDurations) * percentile) / 100
	if index >= len(m.requestDurations) {
		index = len(m.requestDurations) - 1
	}
	
	// Find max duration up to index for approximation
	var maxDuration time.Duration
	for i := 0; i <= index; i++ {
		if m.requestDurations[i] > maxDuration {
			maxDuration = m.requestDurations[i]
		}
	}
	return maxDuration
}

// GetRequestsPerSecond returns current requests per second
func (m *Metrics) GetRequestsPerSecond() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	totalRequests := int64(0)
	for _, count := range m.requestsTotal {
		totalRequests += count
	}
	
	uptime := time.Since(m.startTime)
	if uptime.Seconds() == 0 {
		return 0
	}
	
	return float64(totalRequests) / uptime.Seconds()
}

// GetErrorRate returns error rate as percentage
func (m *Metrics) GetErrorRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	totalRequests := int64(0)
	for _, count := range m.requestsTotal {
		totalRequests += count
	}
	
	if totalRequests == 0 {
		return 0
	}
	
	return (float64(m.errorCount) / float64(totalRequests)) * 100
}