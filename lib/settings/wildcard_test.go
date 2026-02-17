package settings

import (
	"sync"
	"testing"
)

// TestTrackWildcard tests wildcard pattern matching in callbacks
func TestTrackWildcard(t *testing.T) {
	// Reset callbacks for clean test
	mu.Lock()
	changeCallbacks = []callbackEntry{}
	mu.Unlock()

	// Track callback invocations
	var callbacksMu sync.Mutex
	exactCalls := []string{}
	prefixCalls := []string{}
	allCalls := []string{}

	// Register exact match callback
	Track("DATABASE.HOST", func() {
		callbacksMu.Lock()
		exactCalls = append(exactCalls, "DATABASE_HOST")
		callbacksMu.Unlock()
	})

	// Register prefix wildcard callback
	Track("DATABASE.*", func() {
		callbacksMu.Lock()
		// Track that callback was called (can't know which key without params)
		prefixCalls = append(prefixCalls, "called")
		callbacksMu.Unlock()
	})

	// Register global wildcard callback
	Track("*", func() {
		callbacksMu.Lock()
		allCalls = append(allCalls, "called")
		callbacksMu.Unlock()
	})

	// Test: Set DATABASE.HOST
	Set("DATABASE.HOST", "localhost")

	// Verify: Should trigger exact, prefix, and global callbacks
	// Note: +1 because callbacks are called immediately on registration
	callbacksMu.Lock()
	if len(exactCalls) != 2 { // 1 on registration + 1 on Set
		t.Errorf("Expected exactCalls count=2, got %d", len(exactCalls))
	}
	if len(prefixCalls) != 2 { // 1 on registration + 1 on Set
		t.Errorf("Expected prefixCalls count=2, got %d", len(prefixCalls))
	}
	if len(allCalls) != 2 { // 1 on registration + 1 on Set
		t.Errorf("Expected allCalls count=2, got %d", len(allCalls))
	}
	callbacksMu.Unlock()

	// Test: Set DATABASE.PORT
	Set("DATABASE.PORT", 3306)

	// Verify: Should trigger prefix and global, but NOT exact
	callbacksMu.Lock()
	if len(exactCalls) != 2 { // No additional call
		t.Errorf("Expected exactCalls count=2, got %d", len(exactCalls))
	}
	if len(prefixCalls) != 3 { // +1 more
		t.Errorf("Expected prefixCalls count=3, got %d", len(prefixCalls))
	}
	if len(allCalls) != 3 { // +1 more
		t.Errorf("Expected allCalls count=3, got %d", len(allCalls))
	}
	callbacksMu.Unlock()

	// Test: Set CACHE.HOST (different prefix)
	Set("CACHE.HOST", "redis")

	// Verify: Should only trigger global, not exact or prefix
	callbacksMu.Lock()
	if len(exactCalls) != 2 {
		t.Errorf("Expected exactCalls count=2, got %d", len(exactCalls))
	}
	if len(prefixCalls) != 3 {
		t.Errorf("Expected prefixCalls count=3, got %d", len(prefixCalls))
	}
	if len(allCalls) != 4 { // +1 more
		t.Errorf("Expected allCalls count=4, got %d", len(allCalls))
	}
	callbacksMu.Unlock()
}

// TestTrackMultipleWildcards tests multiple wildcard patterns
func TestTrackMultipleWildcards(t *testing.T) {
	// Reset callbacks and data
	mu.Lock()
	changeCallbacks = []callbackEntry{}
	data = map[string]any{}
	mu.Unlock()

	var mu1 sync.Mutex
	dbCalls := 0
	cacheCalls := 0

	// Watch DATABASE.*
	Track("DATABASE.*", func() {
		mu1.Lock()
		dbCalls++
		t.Logf("DATABASE.* callback triggered")
		mu1.Unlock()
	})

	// Watch CACHE.*
	Track("CACHE.*", func() {
		mu1.Lock()
		cacheCalls++
		t.Logf("CACHE.* callback triggered")
		mu1.Unlock()
	})

	// Trigger callbacks
	Set("DATABASE.HOST", "localhost")
	Set("DATABASE.PORT", 3306)
	Set("CACHE.HOST", "redis")
	Set("CACHE.PORT", 6379)
	Set("APP.NAME", "test") // Should not trigger either

	// Verify counts (includes initial call on registration)
	mu1.Lock()
	if dbCalls != 3 { // 1 on registration + 2 Set calls
		t.Errorf("Expected dbCalls=3, got %d", dbCalls)
	}
	if cacheCalls != 3 { // 1 on registration + 2 Set calls
		t.Errorf("Expected cacheCalls=3, got %d", cacheCalls)
	}
	mu1.Unlock()
}

// BenchmarkTrackExact benchmarks exact match callbacks
func BenchmarkTrackExact(b *testing.B) {
	mu.Lock()
	changeCallbacks = []callbackEntry{}
	mu.Unlock()

	Track("TEST.KEY", func() {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set("TEST.KEY", i)
	}
}

// BenchmarkTrackWildcard benchmarks wildcard match callbacks
func BenchmarkTrackWildcard(b *testing.B) {
	mu.Lock()
	changeCallbacks = []callbackEntry{}
	mu.Unlock()

	Track("TEST.*", func() {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set("TEST.KEY", i)
	}
}

// BenchmarkTrackGlobal benchmarks global wildcard callbacks
func BenchmarkTrackGlobal(b *testing.B) {
	mu.Lock()
	changeCallbacks = []callbackEntry{}
	mu.Unlock()

	Track("*", func() {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set("TEST.KEY", i)
	}
}
