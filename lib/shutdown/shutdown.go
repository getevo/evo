// Package shutdown provides a global registry for shutdown hooks.
// It is a leaf package with no internal dependencies so it can be safely
// imported by connectors, middleware, and the root evo package alike
// without creating import cycles.
package shutdown

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// hook pairs a registered function with the source location that registered it.
type hook struct {
	fn     func()
	caller string // "file.go:line"
}

var (
	mu    sync.Mutex
	hooks []hook
)

// Register appends fn to the list of functions that will be called by Run.
// The call site (file:line) is captured at registration time and printed
// when the hook fires during shutdown.
// Hooks are invoked in registration order.
func Register(fn func()) {
	caller := captureCallerOutsidePackage()
	mu.Lock()
	defer mu.Unlock()
	hooks = append(hooks, hook{fn: fn, caller: caller})
}

// Run calls every registered hook in registration order, printing the
// originating source location before each one.
// It is safe to call from multiple goroutines; only the first call executes
// the hooks — subsequent calls are no-ops.
func Run() {
	mu.Lock()
	fns := make([]hook, len(hooks))
	copy(fns, hooks)
	hooks = nil // prevent double-execution
	mu.Unlock()

	for _, h := range fns {
		fmt.Printf("shutting down  %s\n", h.caller)
		h.fn()
	}
}

// captureCallerOutsidePackage walks the call stack and returns the file:line
// of the first frame that belongs neither to this package nor to the evo
// root-package wrapper (evo.OnShutdown). This means the location always
// points to the actual user/connector code that registered the hook.
func captureCallerOutsidePackage() string {
	pcs := make([]uintptr, 16)
	// Skip runtime.Callers + captureCallerOutsidePackage + Register itself.
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		fn := frame.Function
		// Skip frames that are part of the shutdown registry itself or the
		// thin evo.OnShutdown wrapper so the location always points to the
		// caller's code.
		if !strings.Contains(fn, "getevo/evo/v2/lib/shutdown") &&
			fn != "github.com/getevo/evo/v2.OnShutdown" {
			return fmt.Sprintf("%s:%d", frame.File, frame.Line)
		}
		if !more {
			break
		}
	}
	return "unknown"
}
