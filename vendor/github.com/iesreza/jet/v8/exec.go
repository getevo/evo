// Copyright 2016 José Santos <henrique_1609@me.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Jet is a fast and dynamic template engine for the Go programming language, set of features
// includes very fast template execution, a dynamic and flexible language, template inheritance, low number of allocations,
// special interfaces to allow even further optimizations.

package jet

import (
	"bytes"
	"hash"
	"hash/fnv"
	"io"
	"reflect"
	"sort"
	"sync"
	"time"
)

type VarMap map[string]reflect.Value

// SortedKeys returns a sorted slice of VarMap keys
func (scope VarMap) SortedKeys() []string {
	keys := make([]string, 0, len(scope))
	for k := range scope {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (scope VarMap) Set(name string, v interface{}) VarMap {
	scope[name] = reflect.ValueOf(v)
	return scope
}

func (scope VarMap) SetFunc(name string, v Func) VarMap {
	scope[name] = reflect.ValueOf(v)
	return scope
}

func (scope VarMap) SetWriter(name string, v SafeWriter) VarMap {
	scope[name] = reflect.ValueOf(v)
	return scope
}

// Execute executes the template into w.
func (t *Template) Execute(w io.Writer, variables VarMap, data interface{}) (err error) {
	st := pool_State.Get().(*Runtime)
	defer st.recover(&err)

	st.blocks = t.processedBlocks
	st.variables = variables
	st.set = t.set
	st.Writer = w

	// resolve extended template
	for t.extends != nil {
		t = t.extends
	}

	if data != nil {
		st.context = reflect.ValueOf(data)
	}

	st.executeList(t.Root)
	return
}

var _cache = sync.Map{}

func fnv32a(v string) uint32 {
	algorithm := fnv.New32a()
	return uint32Hasher(algorithm, v)
}
func uint32Hasher(algorithm hash.Hash32, text string) uint32 {
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}

var janitor = false

func Janitor(duration time.Duration) {
	if janitor {
		return
	}
	go func() {
		ticker := time.NewTicker(duration)
		for _ = range ticker.C {
			var now = time.Now()
			var deleteBefore = now.Add(-1 * duration)
			_cache.Range(func(key, value interface{}) bool {
				if value.(*Template).lastAccess.Before(deleteBefore) {
					_cache.Delete(key)
				}
				return true
			})
		}
	}()
}

// Execute executes the template
func Execute(template string, variables VarMap, data interface{}) (rendered string, err error) {
	var w = bytes.Buffer{}
	var t *Template
	var key = fnv32a(template)
	v, ok := _cache.Load(key)
	if ok {
		t = v.(*Template)
	} else {
		t = &Template{
			text:         template,
			passedBlocks: make(map[string]*BlockNode),
		}
		defer t.recover(&err)

		lexer := lex("", template, false)
		lexer.setDelimiters("{{", "}}")
		lexer.run()
		t.startParse(lexer)
		t.parseTemplate(false)
		t.stopParse()

		if t.extends != nil {
			t.addBlocks(t.extends.processedBlocks)
		}

		for _, _import := range t.imports {
			t.addBlocks(_import.processedBlocks)
		}

		t.addBlocks(t.passedBlocks)
		_cache.Store(key, t)
	}
	t.lastAccess = time.Now()
	st := pool_State.Get().(*Runtime)

	defer st.recover(&err)

	st.blocks = t.processedBlocks
	st.variables = variables
	st.set = t.set
	st.Writer = &w

	// resolve extended template
	for t.extends != nil {
		t = t.extends
	}

	if data != nil {
		st.context = reflect.ValueOf(data)
	}

	st.executeList(t.Root)
	return w.String(), nil
}
