package tpl

import "github.com/getevo/evo/v2/lib/dot"

// Context holds the execution state during template rendering.
// Scopes form a chain: each for/if block gets a child context.
type Context struct {
	vars         map[string]any
	params       []any    // original Set() arguments
	parent       *Context
	includeDepth int      // tracks nested include/require depth to prevent infinite recursion
}

func newContext(params []any) *Context {
	return &Context{vars: make(map[string]any), params: params}
}

// child creates a nested scope that inherits this context's params, parent chain,
// and current include depth.
func (c *Context) child() *Context {
	return &Context{
		vars:         make(map[string]any),
		params:       c.params,
		parent:       c,
		includeDepth: c.includeDepth,
	}
}

// setLocal sets a variable only in the current (innermost) scope.
func (c *Context) setLocal(name string, val any) {
	c.vars[name] = val
}

// Set assigns a variable. If the name already exists in an ancestor scope,
// it updates there; otherwise it creates it in the current scope.
func (c *Context) Set(name string, val any) {
	if _, ok := c.vars[name]; ok {
		c.vars[name] = val
		return
	}
	if c.parent != nil && c.parent.hasSome(name) {
		c.parent.Set(name, val)
		return
	}
	c.vars[name] = val
}

// hasSome reports whether name exists anywhere in the scope chain.
func (c *Context) hasSome(name string) bool {
	if _, ok := c.vars[name]; ok {
		return true
	}
	if c.parent != nil {
		return c.parent.hasSome(name)
	}
	return false
}

// Get looks up a bare variable name (no dots or brackets).
// Checks scope chain first, then the original params.
func (c *Context) Get(name string) (any, bool) {
	for sc := c; sc != nil; sc = sc.parent {
		if v, ok := sc.vars[name]; ok {
			return v, true
		}
	}
	for _, p := range c.params {
		v, err := dot.Get(p, name)
		if err == nil && v != nil {
			return v, true
		}
	}
	return nil, false
}

// GetPath resolves a full dotted/indexed path such as "user.Name" or "arr[0]".
func (c *Context) GetPath(path string) (any, bool) {
	root := pathRoot(path)
	rootVal, ok := c.Get(root)
	if !ok {
		return nil, false
	}
	if path == root {
		return rootVal, true
	}
	// Wrap rootVal in a synthetic map so dot.Get can traverse the full path,
	// including bracket access like [0] or [key] that needs the root name.
	wrapper := map[string]any{root: rootVal}
	v, err := dot.Get(wrapper, path)
	if err != nil || v == nil {
		return nil, false
	}
	return v, true
}

// GetIndex resolves a path that may include a dynamic variable index at the end.
// E.g.  GetIndex("m", VarExpr{Path:"k"})  → m[value-of-k]
func (c *Context) GetIndex(base any, idx any) any {
	v, err := dot.Get(base, stringify(idx))
	if err != nil {
		return nil
	}
	return v
}

// mergedParams returns a params slice with scope variables prepended as a map.
// Used to feed the existing text-interpolation engine.
func (c *Context) mergedParams() []any {
	merged := make(map[string]any)
	for sc := c; sc != nil; sc = sc.parent {
		for k, v := range sc.vars {
			if _, exists := merged[k]; !exists {
				merged[k] = v
			}
		}
	}
	out := make([]any, 0, len(c.params)+1)
	if len(merged) > 0 {
		out = append(out, merged)
	}
	out = append(out, c.params...)
	return out
}

// pathRoot returns the part of path before the first '.' or '['.
func pathRoot(path string) string {
	for i, r := range path {
		if r == '.' || r == '[' {
			return path[:i]
		}
	}
	return path
}
