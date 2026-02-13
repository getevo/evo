package evo

import (
	"github.com/getevo/evo/v2/lib/outcome"
	"sync"
)

var (
	// healthCheckHooks stores all registered health check functions
	healthCheckHooks []HealthCheckFunc
	// readyCheckHooks stores all registered readiness check functions
	readyCheckHooks []HealthCheckFunc
	// healthMutex protects concurrent access to health check hooks
	healthMutex sync.RWMutex
	// readyMutex protects concurrent access to ready check hooks
	readyMutex sync.RWMutex
)

// HealthCheckFunc is a function that performs a health or readiness check
// It should return nil if the check passes, or an error describing the problem
type HealthCheckFunc func() error

// OnHealthCheck registers a health check function that will be called on /health endpoint
// Health checks determine if the application is running and able to handle requests
//
// Example:
//
//	evo.OnHealthCheck(func() error {
//	    if appIsHealthy() {
//	        return nil
//	    }
//	    return fmt.Errorf("app is unhealthy")
//	})
func OnHealthCheck(fn HealthCheckFunc) {
	healthMutex.Lock()
	defer healthMutex.Unlock()
	healthCheckHooks = append(healthCheckHooks, fn)
}

// OnReadyCheck registers a readiness check function that will be called on /ready endpoint
// Readiness checks determine if the application is ready to receive traffic
// (e.g., database connections established, caches warmed up)
//
// Example:
//
//	evo.OnReadyCheck(func() error {
//	    if db.Ping() == nil {
//	        return nil
//	    }
//	    return fmt.Errorf("database not ready")
//	})
func OnReadyCheck(fn HealthCheckFunc) {
	readyMutex.Lock()
	defer readyMutex.Unlock()
	readyCheckHooks = append(readyCheckHooks, fn)
}

// healthHandler handles GET /health requests
// Returns 200 OK if all health checks pass, 503 Service Unavailable otherwise
func healthHandler(r *Request) any {
	healthMutex.RLock()
	hooks := make([]HealthCheckFunc, len(healthCheckHooks))
	copy(hooks, healthCheckHooks)
	healthMutex.RUnlock()

	// Run all health checks
	for _, hook := range hooks {
		if err := hook(); err != nil {
			return outcome.ServiceUnavailable(map[string]any{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
	}

	return outcome.OK(map[string]any{
		"status": "ok",
	})
}

// readyHandler handles GET /ready requests
// Returns 200 OK if all readiness checks pass, 503 Service Unavailable otherwise
func readyHandler(r *Request) any {
	readyMutex.RLock()
	hooks := make([]HealthCheckFunc, len(readyCheckHooks))
	copy(hooks, readyCheckHooks)
	readyMutex.RUnlock()

	// Run all readiness checks
	for _, hook := range hooks {
		if err := hook(); err != nil {
			return outcome.ServiceUnavailable(map[string]any{
				"status": "not ready",
				"error":  err.Error(),
			})
		}
	}

	return outcome.OK(map[string]any{
		"status": "ok",
	})
}

// registerHealthCheckEndpoints registers the /health and /ready endpoints
// This is called automatically during Setup()
func registerHealthCheckEndpoints() {
	Get("/health", healthHandler)
	Get("/ready", readyHandler)
}
