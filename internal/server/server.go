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
	"github.com/ODudek/ConnX/internal/proxy"
	"github.com/panjf2000/gnet/v2"
)

type Server struct {
	gnet.BuiltinEventEngine
	cfg     *config.Config
	pool    *proxy.ServerPool
	checker *health.Checker
}

func New(cfg *config.Config) (*Server, error) {
	pool := proxy.NewServerPool()

	// Initialize backends
	for _, backendURL := range cfg.Backends {
		backend, err := proxy.NewBackend(backendURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend for %s: %v", backendURL, err)
		}
		pool.AddBackend(backend)
	}

	checker := health.NewChecker(pool, cfg.HealthCheck.Interval, cfg.HealthCheck.Timeout)

	return &Server{
		cfg:     cfg,
		pool:    pool,
		checker: checker,
	}, nil
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	log.Printf("Server is running on %s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.checker.Start()
	return gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	if buf == nil {
		return gnet.None
	}

	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil {
		c.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return gnet.Close
	}

	backend := s.pool.GetNextPeer()
	if backend == nil {
		c.Write([]byte("HTTP/1.1 503 Service Unavailable\r\n\r\n"))
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
		c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return gnet.Close
	}

	// Copy headers
	proxyReq.Header = request.Header

	// Send request to backend
	resp, err := client.Do(proxyReq)
	if err != nil {
		backend.IncrementErrors()
		c.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return gnet.Close
	}
	defer resp.Body.Close()

	// Write response back to client
	responseData, err := httputil.DumpResponse(resp, true)
	if err != nil {
		c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return gnet.Close
	}

	c.Write(responseData)
	return gnet.None
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	return gnet.Run(s, fmt.Sprintf("tcp://%s", addr),
		gnet.WithMulticore(true),
		gnet.WithReusePort(true))
}
