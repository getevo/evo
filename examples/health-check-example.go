package main

import (
	"fmt"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db"
	"time"
)

// Example application demonstrating health check and readiness check usage

type App struct{}

func (App) Name() string { return "example" }

func (App) Register() error {
	// Register health checks during app registration
	// Health checks determine if the app is alive and functioning

	// 1. Simple health check - always healthy
	evo.OnHealthCheck(func() error {
		// Basic liveness check - is the app running?
		return nil
	})

	// 2. Database health check
	evo.OnHealthCheck(func() error {
		// Check if database connection is alive
		if !db.IsEnabled() {
			return nil // DB not required
		}

		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}

		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}

		return nil
	})

	// 3. Memory health check
	evo.OnHealthCheck(func() error {
		// Check if memory usage is within acceptable limits
		// This is a simplified example
		// var m runtime.MemStats
		// runtime.ReadMemStats(&m)
		// if m.Alloc > 1024*1024*1024 { // 1GB
		//     return fmt.Errorf("memory usage too high: %d MB", m.Alloc/1024/1024)
		// }
		return nil
	})

	// Register readiness checks
	// Readiness checks determine if the app is ready to handle traffic

	// 1. Database readiness check
	evo.OnReadyCheck(func() error {
		if !db.IsEnabled() {
			return nil // DB not required
		}

		// Check database is ready to accept queries
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("database not ready: %w", err)
		}

		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("database not responding: %w", err)
		}

		// Check if migrations are complete (if applicable)
		// if !areMigrationsComplete() {
		//     return fmt.Errorf("database migrations pending")
		// }

		return nil
	})

	// 2. Cache readiness check
	evo.OnReadyCheck(func() error {
		// Check if cache is warmed up
		// if !cache.IsReady() {
		//     return fmt.Errorf("cache not initialized")
		// }
		return nil
	})

	// 3. External service readiness check
	evo.OnReadyCheck(func() error {
		// Check if required external services are reachable
		// This example uses a simple timeout-based check
		timeout := time.After(100 * time.Millisecond)
		done := make(chan bool)

		go func() {
			// Simulate checking external service
			// err := checkExternalService()
			// if err != nil {
			//     return
			// }
			done <- true
		}()

		select {
		case <-timeout:
			return fmt.Errorf("external service check timed out")
		case <-done:
			return nil
		}
	})

	return nil
}

func (App) Router() error {
	// Register your application routes here
	evo.Get("/", func(r *evo.Request) any {
		return "Hello World!"
	})

	return nil
}

func (App) WhenReady() error {
	// This runs after all apps are registered
	// Good place for background tasks
	return nil
}

func main() {
	// Setup EVO
	evo.Setup()

	// Register your application
	evo.Register(&App{})

	// Start the server
	// Health endpoints are automatically available at:
	// GET /health - returns 200 if all health checks pass
	// GET /ready  - returns 200 if all readiness checks pass
	evo.Run()
}

/*
Testing the endpoints:

1. Health check (liveness):
   curl http://localhost:8080/health
   Response: {"status":"ok"}

   If any health check fails:
   Response: {"status":"unhealthy","error":"database ping failed: connection refused"}
   Status: 503 Service Unavailable

2. Readiness check:
   curl http://localhost:8080/ready
   Response: {"status":"ok"}

   If any readiness check fails:
   Response: {"status":"not ready","error":"external service check timed out"}
   Status: 503 Service Unavailable

Kubernetes Integration Example:

apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: myapp
    image: myapp:latest
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 3
      periodSeconds: 10
      failureThreshold: 3
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
      failureThreshold: 3
*/
