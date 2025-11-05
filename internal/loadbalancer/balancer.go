package loadbalancer

// Algorithm represents a load balancing algorithm type
type Algorithm string

const (
	RoundRobin         Algorithm = "round-robin"
	WeightedRoundRobin Algorithm = "weighted-round-robin"
	LeastConnections   Algorithm = "least-connections"
)

// Backend interface defines what a backend needs to provide for load balancing
type Backend interface {
	IsAvailable() bool
	GetActiveConns() int64
}

// WeightedBackend extends Backend with weight support
type WeightedBackend interface {
	Backend
	GetWeight() int
}

// Balancer interface for load balancing algorithms
type Balancer interface {
	Next(backends []Backend) Backend
}
