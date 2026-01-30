package router

import (
	"sync"
)

// ServiceRegistry manages backend service endpoints
type ServiceRegistry struct {
	services map[string][]string
	current  map[string]int
	mu       sync.RWMutex
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string][]string),
		current:  make(map[string]int),
	}
}

// RegisterService registers a service endpoint
func (sr *ServiceRegistry) RegisterService(name string, endpoints ...string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.services[name] = endpoints
	sr.current[name] = 0
}

// GetEndpoint returns the next endpoint using round-robin
func (sr *ServiceRegistry) GetEndpoint(name string) string {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	endpoints, exists := sr.services[name]
	if !exists || len(endpoints) == 0 {
		return ""
	}

	// Round-robin selection
	idx := sr.current[name]
	endpoint := endpoints[idx]
	sr.current[name] = (idx + 1) % len(endpoints)

	return endpoint
}

// GetAllEndpoints returns all registered endpoints for a service
func (sr *ServiceRegistry) GetAllEndpoints(name string) []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.services[name]
}
