package loadbalancer

// LeastConnectionsBalancer implements least connections load balancing
type LeastConnectionsBalancer struct{}

// NewLeastConnectionsBalancer creates a new least connections balancer
func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{}
}

// Next returns the backend with the least active connections
func (l *LeastConnectionsBalancer) Next(backends []Backend) Backend {
	if len(backends) == 0 {
		return nil
	}

	var selected Backend
	minConns := int64(-1)

	for _, backend := range backends {
		if !backend.IsAvailable() {
			continue
		}

		conns := backend.GetActiveConns()
		if minConns == -1 || conns < minConns {
			minConns = conns
			selected = backend
		}
	}

	return selected
}
