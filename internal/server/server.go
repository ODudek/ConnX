package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/ODudek/ConnX/internal/config"
	"github.com/ODudek/ConnX/internal/health"
	"github.com/ODudek/ConnX/internal/metrics"
	"github.com/ODudek/ConnX/internal/proxy"
	"github.com/panjf2000/gnet/v2"
)

type Server struct {
	gnet.BuiltinEventEngine
	cfg     *config.Config
	pool    *proxy.ServerPool
	checker *health.Checker
	metrics *metrics.Metrics
}

func New(cfg *config.Config) (*Server, error) {
	pool := proxy.NewServerPool()
	metrics := metrics.NewMetrics()

	// Initialize backends
	for _, backendURL := range cfg.Backends {
		backend, err := proxy.NewBackend(backendURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend for %s: %v", backendURL, err)
		}
		pool.AddBackend(backend)
		metrics.UpdateBackendStatus(backendURL, true)
	}

	checker := health.NewChecker(pool, cfg.HealthCheck.Interval, cfg.HealthCheck.Timeout, metrics)

	return &Server{
		cfg:     cfg,
		pool:    pool,
		checker: checker,
		metrics: metrics,
	}, nil
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	log.Printf("Server is running on %s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.checker.Start()
	return gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
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
