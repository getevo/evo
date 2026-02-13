package evo

import (
	"fmt"
	"testing"
)

// ExampleOnHealthCheck demonstrates how to register a health check
func ExampleOnHealthCheck() {
	// Register a health check that always passes
	OnHealthCheck(func() error {
		// Check if application is healthy
		// Return nil if healthy, error otherwise
		return nil
	})

	// Register multiple health checks
	OnHealthCheck(func() error {
		// Check database connection
		// if db.Ping() != nil {
		//     return fmt.Errorf("database not responding")
		// }
		return nil
	})

	OnHealthCheck(func() error {
		// Check external service
		// if !externalService.IsAlive() {
		//     return fmt.Errorf("external service unreachable")
		// }
		return nil
	})

	fmt.Println("Health checks registered")
	// Output: Health checks registered
}

// ExampleOnReadyCheck demonstrates how to register a readiness check
func ExampleOnReadyCheck() {
	// Register readiness check for database
	OnReadyCheck(func() error {
		// Check if database is ready
		// if !db.IsReady() {
		//     return fmt.Errorf("database not ready")
		// }
		return nil
	})

	// Register readiness check for cache
	OnReadyCheck(func() error {
		// Check if cache is ready
		// if !cache.IsWarmedUp() {
		//     return fmt.Errorf("cache not warmed up")
		// }
		return nil
	})

	fmt.Println("Readiness checks registered")
	// Output: Readiness checks registered
}

func TestHealthCheckRegistration(t *testing.T) {
	// Reset hooks for testing
	healthCheckHooks = nil
	readyCheckHooks = nil

	// Register health checks
	OnHealthCheck(func() error { return nil })
	OnHealthCheck(func() error { return fmt.Errorf("test error") })

	if len(healthCheckHooks) != 2 {
		t.Errorf("Expected 2 health checks, got %d", len(healthCheckHooks))
	}

	// Register ready checks
	OnReadyCheck(func() error { return nil })

	if len(readyCheckHooks) != 1 {
		t.Errorf("Expected 1 ready check, got %d", len(readyCheckHooks))
	}
}

func TestHealthCheckConcurrency(t *testing.T) {
	// Reset hooks
	healthCheckHooks = nil

	// Register checks concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			OnHealthCheck(func() error { return nil })
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if len(healthCheckHooks) != 10 {
		t.Errorf("Expected 10 health checks, got %d", len(healthCheckHooks))
	}
}
