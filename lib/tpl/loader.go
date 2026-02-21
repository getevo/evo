package tpl

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Builder is the fluent API for loading and rendering templates.
//
//	tpl.File("path/to/file.html").Set(ctx).Render()
//	tpl.Text("Hello $name!").Set(ctx).Render()
type Builder struct {
	engine *Engine
	ctx    *Context
}

// File loads a template from a file path. The compiled Engine is cached by
// path, and the cache entry is invalidated automatically when the file's
// modification time changes.
// Returns a Builder whose context is empty until Set() is called.
func File(path string) *Builder {
	eng := fileCache.get(path)
	return &Builder{engine: eng, ctx: newContext(nil)}
}

// Text compiles a template from an inline string. The compiled Engine is cached
// by the full source string.
func Text(src string) *Builder {
	eng := textCache.get(src)
	return &Builder{engine: eng, ctx: newContext(nil)}
}

// Set adds a context value (struct, map, or any value dot.Get can traverse).
// Multiple Set() calls are additive; the first match wins during lookup.
func (b *Builder) Set(ctx any) *Builder {
	b.ctx.params = append(b.ctx.params, ctx)
	return b
}

// Render executes the template and returns the rendered string.
func (b *Builder) Render() string {
	if b.engine == nil {
		return ""
	}
	return b.engine.Execute(b.ctx)
}

// RenderWriter executes the template and writes the output to w.
func (b *Builder) RenderWriter(w io.Writer) {
	if b.engine == nil {
		return
	}
	b.engine.ExecuteWriter(b.ctx, w)
}

// ── Text cache ────────────────────────────────────────────────────────────────

const engineCacheLimit = 1000 // max compiled engines kept in the text-source cache

type engineCache struct {
	mu    sync.RWMutex
	store map[string]*Engine
}

func (c *engineCache) get(key string) *Engine {
	c.mu.RLock()
	e, ok := c.store[key]
	c.mu.RUnlock()
	if ok {
		return e
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// Double-check under write lock.
	if e, ok = c.store[key]; ok {
		return e
	}
	e = CompileEngine(key)
	// Only cache if we haven't exceeded the limit (avoids unbounded growth).
	if len(c.store) < engineCacheLimit {
		c.store[key] = e
	}
	return e
}

var textCache = &engineCache{store: make(map[string]*Engine)}

// ── File cache (with mtime invalidation) ──────────────────────────────────────

type cachedFileEngine struct {
	engine *Engine
	mtime  time.Time
}

type fileEngineCache struct {
	mu    sync.RWMutex
	store map[string]*cachedFileEngine
}

func (c *fileEngineCache) get(path string) *Engine {
	// Check if file exists and get current mtime
	info, statErr := os.Stat(path)

	c.mu.RLock()
	entry, ok := c.store[path]
	c.mu.RUnlock()

	if ok {
		if statErr != nil {
			// File no longer exists but we have a cached version; return empty engine
			return &Engine{}
		}
		if !info.ModTime().After(entry.mtime) {
			// Cache is still fresh
			return entry.engine
		}
		// File has been modified — fall through to recompile
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check under write lock
	if entry, ok = c.store[path]; ok {
		if statErr == nil && !info.ModTime().After(entry.mtime) {
			return entry.engine
		}
	}

	var eng *Engine
	if statErr != nil {
		eng = &Engine{}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			eng = &Engine{}
		} else {
			eng = CompileEngine(string(data))
		}
	}

	var mtime time.Time
	if statErr == nil {
		mtime = info.ModTime()
	}
	c.store[path] = &cachedFileEngine{engine: eng, mtime: mtime}
	return eng
}

var fileCache = &fileEngineCache{store: make(map[string]*cachedFileEngine)}

// ClearCache clears both the file and text template caches.
func ClearCache() {
	textCache.mu.Lock()
	textCache.store = make(map[string]*Engine)
	textCache.mu.Unlock()

	fileCache.mu.Lock()
	fileCache.store = make(map[string]*cachedFileEngine)
	fileCache.mu.Unlock()
}

// ── Convenience render functions ──────────────────────────────────────────────

// RenderText compiles src and executes it with the given params.
// This is the engine-aware version of the existing Render() function.
func RenderText(src string, params ...any) string {
	b := Text(src)
	for _, p := range params {
		b.Set(p)
	}
	return b.Render()
}

// RenderFile loads path and executes it with the given params.
func RenderFile(path string, params ...any) string {
	b := File(path)
	for _, p := range params {
		b.Set(p)
	}
	return b.Render()
}

// ── String builder io.Writer adapter ─────────────────────────────────────────

// ensure strings.Builder satisfies io.Writer (it does via WriteByte/WriteString)
var _ io.Writer = (*strings.Builder)(nil)
