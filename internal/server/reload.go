package server

import (
	"log"

	"github.com/ODudek/ConnX/internal/config"
	"github.com/ODudek/ConnX/internal/health"
	"github.com/ODudek/ConnX/internal/loadbalancer"
	"github.com/ODudek/ConnX/internal/metrics"
	"github.com/ODudek/ConnX/internal/proxy"
	"github.com/ODudek/ConnX/internal/ratelimit"
)

// newHealthChecker creates a new health checker instance
func newHealthChecker(pool *proxy.ServerPool, interval, timeout int, metrics *metrics.Metrics) *health.Checker {
	return health.NewChecker(pool, interval, timeout, metrics)
}

// Reload reloads configuration dynamically without restarting
func (s *Server) Reload(newCfg *config.Config) error {
	log.Println("=== Starting configuration reload ===")

	// Update rate limiter if settings changed
	if s.cfg.RateLimit.Enabled != newCfg.RateLimit.Enabled ||
		s.cfg.RateLimit.RequestsPerSec != newCfg.RateLimit.RequestsPerSec ||
		s.cfg.RateLimit.Burst != newCfg.RateLimit.Burst ||
		s.cfg.RateLimit.PerIP != newCfg.RateLimit.PerIP {

		log.Println("Rate limiter configuration changed, updating...")
		s.rateLimiter = ratelimit.NewRateLimiter(ratelimit.Config{
			Enabled:        newCfg.RateLimit.Enabled,
			RequestsPerSec: newCfg.RateLimit.RequestsPerSec,
			Burst:          newCfg.RateLimit.Burst,
		})

		if newCfg.RateLimit.Enabled {
			s.rateLimiter.StartCleanup()
		}
		log.Printf("Rate limiter updated: enabled=%v, rps=%d, burst=%d, perIP=%v",
			newCfg.RateLimit.Enabled, newCfg.RateLimit.RequestsPerSec,
			newCfg.RateLimit.Burst, newCfg.RateLimit.PerIP)
	}

	// Update load balancing algorithm if changed
	if s.cfg.LoadBalancing.Algorithm != newCfg.LoadBalancing.Algorithm {
		log.Printf("Load balancing algorithm changed: %s -> %s",
			s.cfg.LoadBalancing.Algorithm, newCfg.LoadBalancing.Algorithm)

		var balancer loadbalancer.Balancer
		switch loadbalancer.Algorithm(newCfg.LoadBalancing.Algorithm) {
		case loadbalancer.WeightedRoundRobin:
			balancer = loadbalancer.NewWeightedRoundRobinBalancer()
		case loadbalancer.LeastConnections:
			balancer = loadbalancer.NewLeastConnectionsBalancer()
		default:
			balancer = loadbalancer.NewRoundRobinBalancer()
		}

		s.pool.SetBalancer(balancer)
		log.Printf("Load balancer updated to: %s", newCfg.LoadBalancing.Algorithm)
	}

	// Update backends
	s.updateBackends(newCfg)

	// Update health checker intervals if changed
	if s.cfg.HealthCheck.Interval != newCfg.HealthCheck.Interval ||
		s.cfg.HealthCheck.Timeout != newCfg.HealthCheck.Timeout {

		log.Println("Health check configuration changed, restarting checker...")
		s.checker.Stop()
		s.checker = nil

		// Create new checker with new intervals
		s.checker = newHealthChecker(s.pool, newCfg.HealthCheck.Interval,
			newCfg.HealthCheck.Timeout, s.metrics)
		s.checker.Start()

		log.Printf("Health checker updated: interval=%ds, timeout=%ds",
			newCfg.HealthCheck.Interval, newCfg.HealthCheck.Timeout)
	}

	// Update configuration reference
	s.cfg = newCfg

	log.Println("=== Configuration reload completed ===")
	return nil
}

// updateBackends dynamically updates the backend pool
func (s *Server) updateBackends(newCfg *config.Config) {
	log.Println("Updating backends...")

	// Create map of new backends for quick lookup
	newBackendsMap := make(map[string]config.BackendConfig)
	for _, bCfg := range newCfg.Backends {
		newBackendsMap[bCfg.URL] = bCfg
	}

	// Create map of existing backends
	existingBackends := s.pool.GetBackends()
	existingMap := make(map[string]*proxy.Backend)
	for _, b := range existingBackends {
		existingMap[b.URL.String()] = b
	}

	// Find backends to add and update
	var toAdd []*proxy.Backend
	for url, bCfg := range newBackendsMap {
		if existing, ok := existingMap[url]; ok {
			// Backend exists, check if weight changed
			if existing.Weight != bCfg.Weight {
				log.Printf("Backend %s weight changed: %d -> %d",
					url, existing.Weight, bCfg.Weight)
				existing.Weight = bCfg.Weight
			}
		} else {
			// New backend - add it
			backend, err := proxy.NewBackend(bCfg.URL, bCfg.Weight)
			if err != nil {
				log.Printf("Error creating backend %s: %v", bCfg.URL, err)
				continue
			}
			toAdd = append(toAdd, backend)
			log.Printf("Adding new backend: %s (weight: %d)", bCfg.URL, bCfg.Weight)
		}
	}

	// Find backends to remove
	var toRemove []*proxy.Backend
	for url, backend := range existingMap {
		if _, ok := newBackendsMap[url]; !ok {
			toRemove = append(toRemove, backend)
			log.Printf("Removing backend: %s", url)
		}
	}

	// Apply changes
	for _, backend := range toAdd {
		s.pool.AddBackend(backend)
		s.metrics.UpdateBackendStatus(backend.URL.String(), true)
	}

	for _, backend := range toRemove {
		s.pool.RemoveBackend(backend)
		s.metrics.UpdateBackendStatus(backend.URL.String(), false)
	}

	log.Printf("Backends updated: %d added, %d removed, %d total",
		len(toAdd), len(toRemove), len(newCfg.Backends))
}
