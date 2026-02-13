package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// HealthCheckResult contains the results of a database health check
type HealthCheckResult struct {
	Healthy          bool          `json:"healthy"`
	ResponseTime     time.Duration `json:"response_time"`
	ConnectionsOpen  int           `json:"connections_open"`
	ConnectionsInUse int           `json:"connections_in_use"`
	ConnectionsIdle  int           `json:"connections_idle"`
	MaxOpenConns     int           `json:"max_open_conns"`
	Error            string        `json:"error,omitempty"`
}

// Ping checks if the database connection is alive and returns basic stats
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// HealthCheck performs a comprehensive health check on the database connection
// and returns detailed statistics about the connection pool
func HealthCheck(ctx context.Context, db *gorm.DB) HealthCheckResult {
	result := HealthCheckResult{
		Healthy: false,
	}

	sqlDB, err := db.DB()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get database instance: %v", err)
		return result
	}

	// Measure ping response time
	start := time.Now()
	if err := sqlDB.PingContext(ctx); err != nil {
		result.Error = fmt.Sprintf("ping failed: %v", err)
		return result
	}
	result.ResponseTime = time.Since(start)
	result.Healthy = true

	// Get connection pool stats
	stats := sqlDB.Stats()
	result.ConnectionsOpen = stats.OpenConnections
	result.ConnectionsInUse = stats.InUse
	result.ConnectionsIdle = stats.Idle
	result.MaxOpenConns = stats.MaxOpenConnections

	return result
}

// WaitForDB waits for the database to become available with retries
// Useful for application startup when the database might not be ready immediately
func WaitForDB(ctx context.Context, db *gorm.DB, maxRetries int, retryInterval time.Duration) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	for i := 0; i < maxRetries; i++ {
		if err := sqlDB.PingContext(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// Continue to next retry
		}
	}

	return fmt.Errorf("database not available after %d retries", maxRetries)
}

// GetConnectionStats returns current database connection pool statistics
func GetConnectionStats(db *gorm.DB) (sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return sql.DBStats{}, fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Stats(), nil
}

// CloseConnection gracefully closes the database connection
func CloseConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}
