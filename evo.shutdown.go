package evo

import "github.com/getevo/evo/v2/lib/shutdown"

// OnShutdown registers fn to be called during graceful shutdown, before the
// HTTP server stops accepting connections. Hooks are called in the order they
// were registered. Safe to call from any goroutine at any time before shutdown.
func OnShutdown(fn func()) {
	shutdown.Register(fn)
}
