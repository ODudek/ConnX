package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/ODudek/ConnX/internal/config"
	"github.com/ODudek/ConnX/internal/health"
	"github.com/ODudek/ConnX/internal/loadbalancer"
	"github.com/ODudek/ConnX/internal/metrics"
	"github.com/ODudek/ConnX/internal/proxy"
	"github.com/ODudek/ConnX/internal/ratelimit"
	"github.com/panjf2000/gnet/v2"
)

type Server struct {
	gnet.BuiltinEventEngine
	cfg         *config.Config
	pool        *proxy.ServerPool
	checker     *health.Checker
	metrics     *metrics.Metrics
	rateLimiter *ratelimit.RateLimiter
	engine      gnet.Engine
	shutdownCtx context.Context
	shutdownFn  context.CancelFunc
	activeConns sync.WaitGroup
}

func New(cfg *config.Config) (*Server, error) {
	// Create load balancer based on configuration
	var balancer loadbalancer.Balancer
	switch loadbalancer.Algorithm(cfg.LoadBalancing.Algorithm) {
	case loadbalancer.WeightedRoundRobin:
		balancer = loadbalancer.NewWeightedRoundRobinBalancer()
	case loadbalancer.LeastConnections:
		balancer = loadbalancer.NewLeastConnectionsBalancer()
	default:
		balancer = loadbalancer.NewRoundRobinBalancer()
	}

	pool := proxy.NewServerPool(balancer)
	metricsCollector := metrics.NewMetrics()

	// Initialize backends with weights
	for _, backendCfg := range cfg.Backends {
		backend, err := proxy.NewBackend(backendCfg.URL, backendCfg.Weight)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend for %s: %v", backendCfg.URL, err)
		}
		pool.AddBackend(backend)
		metricsCollector.UpdateBackendStatus(backendCfg.URL, true)
	}

	checker := health.NewChecker(pool, cfg.HealthCheck.Interval, cfg.HealthCheck.Timeout, metricsCollector)

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(ratelimit.Config{
		Enabled:        cfg.RateLimit.Enabled,
		RequestsPerSec: cfg.RateLimit.RequestsPerSec,
		Burst:          cfg.RateLimit.Burst,
	})

	if cfg.RateLimit.Enabled {
		rateLimiter.StartCleanup()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		cfg:         cfg,
		pool:        pool,
		checker:     checker,
		metrics:     metricsCollector,
		rateLimiter: rateLimiter,
		shutdownCtx: ctx,
		shutdownFn:  cancel,
	}, nil
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.engine = eng
	log.Printf("Server is running on %s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Printf("Load balancing algorithm: %s", s.cfg.LoadBalancing.Algorithm)
	log.Printf("Circuit breaker enabled: %v", s.cfg.CircuitBreaker.Enabled)
	log.Printf("Rate limiting enabled: %v", s.cfg.RateLimit.Enabled)
	s.checker.Start()
	return gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	s.activeConns.Add(1)
	defer s.activeConns.Done()

	startTime := time.Now()
	buf, _ := c.Next(-1)
	if buf == nil {
		return gnet.None
	}

	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil {
		s.metrics.IncrementRequests("UNKNOWN", 400)
		s.metrics.IncrementErrors()
		_, err := c.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		if err != nil {
			log.Println("Failed to write response:", err)
		}
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}

	// Check if this is a metrics request
	if request.URL.Path == "/metrics" {
		return s.handleMetricsRequest(c, request, startTime)
	}

	// Rate limiting check
	if s.cfg.RateLimit.Enabled {
		var allowed bool
		if s.cfg.RateLimit.PerIP {
			// Extract IP from connection
			remoteAddr := c.RemoteAddr().String()
			ip, _, _ := net.SplitHostPort(remoteAddr)
			allowed = s.rateLimiter.AllowIP(ip)
		} else {
			allowed = s.rateLimiter.AllowGlobal()
		}

		if !allowed {
			s.metrics.IncrementRequests(request.Method, 429)
			s.metrics.IncrementErrors()
			_, err := c.Write([]byte("HTTP/1.1 429 Too Many Requests\r\n\r\n"))
			if err != nil {
				log.Println("Failed to write rate limit response:", err)
			}
			s.metrics.RecordRequestDuration(time.Since(startTime))
			return gnet.Close
		}
	}

	backend := s.pool.GetNextPeer()
	if backend == nil {
		s.metrics.IncrementRequests(request.Method, 503)
		s.metrics.IncrementErrors()
		_, err := c.Write([]byte("HTTP/1.1 503 Service Unavailable\r\n\r\n"))
		if err != nil {
			log.Println("Failed to get backend:", err)
		}
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}
	defer s.pool.ReleaseBackend(backend)

	// Forward request to backend
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Create new URL
	backendURL := *backend.URL
	backendURL.Path = request.URL.Path
	backendURL.RawQuery = request.URL.RawQuery

	// Create new request
	proxyReq, err := http.NewRequest(request.Method, backendURL.String(), request.Body)
	if err != nil {
		s.metrics.IncrementRequests(request.Method, 500)
		s.metrics.IncrementErrors()
		_, err := c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		if err != nil {
			log.Println("Failed to create request:", err)
		}
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}

	// Copy headers
	proxyReq.Header = request.Header

	// Send request to backend
	resp, err := client.Do(proxyReq)
	if err != nil {
		backend.IncrementErrors()
		backend.CircuitBreaker.RecordFailure()
		s.metrics.IncrementRequests(request.Method, 502)
		s.metrics.IncrementErrors()
		_, err := c.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		if err != nil {
			log.Println("Failed to write response:", err)
		}
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}
	defer resp.Body.Close()

	// Record successful request
	backend.CircuitBreaker.RecordSuccess()
	s.metrics.IncrementRequests(request.Method, resp.StatusCode)
	
	// Write response back to client
	responseData, err := httputil.DumpResponse(resp, true)
	if err != nil {
		s.metrics.IncrementRequests(request.Method, 500)
		s.metrics.IncrementErrors()
		_, err := c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		if err != nil {
			log.Println("Failed to write response:", err)
		}
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}

	_, err = c.Write(responseData)
	if err != nil {
		log.Println("Failed to write response:", err)
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}
	
	s.metrics.RecordRequestDuration(time.Since(startTime))
	return gnet.None
}

// handleMetricsRequest handles the /metrics endpoint
func (s *Server) handleMetricsRequest(c gnet.Conn, request *http.Request, startTime time.Time) gnet.Action {
	if request.Method != "GET" {
		s.metrics.IncrementRequests(request.Method, 405)
		response := "HTTP/1.1 405 Method Not Allowed\r\nContent-Type: text/plain\r\n\r\nMethod Not Allowed"
		_, err := c.Write([]byte(response))
		if err != nil {
			log.Println("Failed to write response:", err)
		}
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}

	// Get metrics data
	metricsData := s.metrics.GetPrometheusMetrics()
	
	// Build HTTP response
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s",
		len(metricsData), metricsData)
	
	s.metrics.IncrementRequests("GET", 200)
	_, err := c.Write([]byte(response))
	if err != nil {
		log.Println("Failed to write metrics response:", err)
		s.metrics.RecordRequestDuration(time.Since(startTime))
		return gnet.Close
	}
	
	s.metrics.RecordRequestDuration(time.Since(startTime))
	return gnet.Close
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	return gnet.Run(s, fmt.Sprintf("tcp://%s", addr),
		gnet.WithMulticore(true),
		gnet.WithReusePort(true))
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	log.Println("Initiating graceful shutdown...")

	// Signal shutdown
	s.shutdownFn()

	// Note: gnet doesn't support graceful shutdown via API
	// The server will stop when main goroutine exits
	log.Println("Server shutdown initiated...")

	// Stop health checker
	if s.checker != nil {
		log.Println("Stopping health checker...")
		s.checker.Stop()
	}

	// Wait for active connections to finish with timeout
	done := make(chan struct{})
	go func() {
		s.activeConns.Wait()
		close(done)
	}()

	timeout := time.Duration(s.cfg.Server.ShutdownTimeout) * time.Second
	select {
	case <-done:
		log.Println("All connections closed")
	case <-time.After(timeout):
		log.Printf("Shutdown timeout reached (%v), forcing shutdown", timeout)
	}

	log.Println("Server stopped")
	return nil
}
