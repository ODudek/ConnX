package loadbalancer

import (
	"sync/atomic"
)

// WeightedRoundRobinBalancer implements weighted round-robin load balancing
type WeightedRoundRobinBalancer struct {
	current        uint64
	currentWeight  int
	maxWeight      int
	gcdWeight      int
	totalBackends  int
}

// NewWeightedRoundRobinBalancer creates a new weighted round-robin balancer
func NewWeightedRoundRobinBalancer() *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{}
}

// Next returns the next backend using weighted round-robin algorithm
func (w *WeightedRoundRobinBalancer) Next(backends []Backend) Backend {
	if len(backends) == 0 {
		return nil
	}

	// Recalculate weights if backend count changed
	if w.totalBackends != len(backends) {
		w.recalculateWeights(backends)
	}

	// If all weights are 1 (or 0), use simple round-robin
	if w.maxWeight <= 1 {
		return w.simpleRoundRobin(backends)
	}

	// Weighted round-robin algorithm
	for i := 0; i < len(backends)*w.maxWeight; i++ {
		idx := int(atomic.AddUint64(&w.current, 1) % uint64(len(backends)))
		backend := backends[idx]

		if !backend.IsAvailable() {
			continue
		}

		// Try to get weight from backend
		var weight int
		if wb, ok := backend.(WeightedBackend); ok {
			weight = wb.GetWeight()
		} else {
			weight = 1
		}

		if w.currentWeight <= 0 {
			w.currentWeight = w.maxWeight
		}

		if weight >= w.currentWeight {
			w.currentWeight -= w.gcdWeight
			return backend
		}

		w.currentWeight -= w.gcdWeight
	}

	return nil
}

// simpleRoundRobin fallback for backends with equal or no weights
func (w *WeightedRoundRobinBalancer) simpleRoundRobin(backends []Backend) Backend {
	startIdx := int(atomic.AddUint64(&w.current, 1) % uint64(len(backends)))

	for i := 0; i < len(backends); i++ {
		idx := (startIdx + i) % len(backends)
		backend := backends[idx]
		if backend.IsAvailable() {
			return backend
		}
	}

	return nil
}

// recalculateWeights calculates max weight and GCD for weighted selection
func (w *WeightedRoundRobinBalancer) recalculateWeights(backends []Backend) {
	w.totalBackends = len(backends)
	w.maxWeight = 0
	w.gcdWeight = 0

	// Find max weight and calculate GCD
	for _, backend := range backends {
		var weight int
		if wb, ok := backend.(WeightedBackend); ok {
			weight = wb.GetWeight()
		} else {
			weight = 1
		}

		if weight > w.maxWeight {
			w.maxWeight = weight
		}
		if w.gcdWeight == 0 {
			w.gcdWeight = weight
		} else {
			w.gcdWeight = gcd(w.gcdWeight, weight)
		}
	}

	if w.gcdWeight == 0 {
		w.gcdWeight = 1
	}
	if w.maxWeight == 0 {
		w.maxWeight = 1
	}

	w.currentWeight = w.maxWeight
}

// gcd calculates the greatest common divisor
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
