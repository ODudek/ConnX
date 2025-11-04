package loadbalancer

import (
	"sync/atomic"
)

// RoundRobinBalancer implements simple round-robin load balancing
type RoundRobinBalancer struct {
	current uint64
}

// NewRoundRobinBalancer creates a new round-robin balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Next returns the next backend using round-robin algorithm
func (r *RoundRobinBalancer) Next(backends []Backend) Backend {
	if len(backends) == 0 {
		return nil
	}

	// Try to find an available backend
	startIdx := int(atomic.AddUint64(&r.current, 1) % uint64(len(backends)))

	for i := 0; i < len(backends); i++ {
		idx := (startIdx + i) % len(backends)
		backend := backends[idx]
		if backend.IsAvailable() {
			return backend
		}
	}

	return nil
}
