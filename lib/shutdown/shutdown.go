// Package shutdown provides a global registry for shutdown hooks.
// It is a leaf package with no internal dependencies so it can be safely
// imported by connectors, middleware, and the root evo package alike
// without creating import cycles.
package shutdown

import "sync"

var (
	mu    sync.Mutex
	hooks []func()
)

// Register appends fn to the list of functions that will be called by Run.
// Hooks are invoked in registration order.
func Register(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	hooks = append(hooks, fn)
}

// Run calls every registered hook in registration order.
// It is safe to call from multiple goroutines; only the first call executes
// the hooks — subsequent calls are no-ops.
func Run() {
	mu.Lock()
	fns := make([]func(), len(hooks))
	copy(fns, hooks)
	hooks = nil // prevent double-execution
	mu.Unlock()

	for _, fn := range fns {
		fn()
	}
}
