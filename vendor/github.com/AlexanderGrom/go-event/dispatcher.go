package event

import (
	"errors"
	"reflect"
	"sync"
)

// Event implementation
type event struct {
	sync.RWMutex

	events map[string][]interface{}
}

// New returns a new event.Dispatcher
func New() Dispatcher {
	return &event{
		events: make(map[string][]interface{}),
	}
}

// On set new listener
func (e *event) On(name string, fn interface{}) error {
	e.Lock()
	defer e.Unlock()

	if fn == nil {
		return errors.New("fn is nil")
	}
	if _, ok := fn.(handle); ok {
		e.events[name] = append(e.events[name], fn)
		return nil
	}

	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return errors.New("fn is not a function")
	}
	if t.NumOut() != 1 {
		return errors.New("fn must have one return value")
	}
	if t.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return errors.New("fn must return an error message")
	}
	if list, ok := e.events[name]; ok && len(list) > 0 {
		tt := reflect.TypeOf(list[0])
		if tt.NumIn() != t.NumIn() {
			return errors.New("fn signature is not equal")
		}
		for i := 0; i < tt.NumIn(); i++ {
			if tt.In(i) != t.In(i) {
				return errors.New("fn signature is not equal")
			}
		}
	}

	e.events[name] = append(e.events[name], fn)
	return nil
}

// Go firing an event
func (e *event) Go(name string, params ...interface{}) error {
	e.RLock()
	defer e.RUnlock()

	fns := e.events[name]
	for i := len(fns) - 1; i >= 0; i-- {
		stopped, err := e.call(fns[i], params...)
		if err != nil {
			return err
		}
		if stopped {
			break
		}
	}

	return nil
}

func (e *event) call(fn interface{}, params ...interface{}) (stopped bool, err error) {
	if f, ok := fn.(handle); ok {
		if len(params) != 1 {
			return stopped, errors.New("parameters mismatched")
		}
		event, ok := (params[0]).(Eventer)
		if !ok {
			return stopped, errors.New("parameters mismatched")
		}
		err = f(event)
		return event.IsPropagationStopped(), err
	}

	var (
		f     = reflect.ValueOf(fn)
		t     = f.Type()
		numIn = t.NumIn()
		in    = make([]reflect.Value, 0, numIn)
	)

	if t.IsVariadic() {
		n := numIn - 1
		if len(params) < n {
			return stopped, errors.New("parameters mismatched")
		}
		for _, param := range params[:n] {
			in = append(in, reflect.ValueOf(param))
		}
		s := reflect.MakeSlice(t.In(n), 0, len(params[n:]))
		for _, param := range params[n:] {
			s = reflect.Append(s, reflect.ValueOf(param))
		}
		in = append(in, s)

		err, _ = f.CallSlice(in)[0].Interface().(error)
		return stopped, err
	}

	if len(params) != numIn {
		return stopped, errors.New("parameters mismatched")
	}
	for _, param := range params {
		in = append(in, reflect.ValueOf(param))
	}

	err, _ = f.Call(in)[0].Interface().(error)
	return stopped, err
}

// Has returns true if a event exists
func (e *event) Has(name string) bool {
	e.RLock()
	defer e.RUnlock()
	_, ok := e.events[name]
	return ok
}

// List returns list events
func (e *event) List() []string {
	e.RLock()
	defer e.RUnlock()
	list := make([]string, 0, len(e.events))
	for name := range e.events {
		list = append(list, name)
	}
	return list
}

// Remove delete events from the event list
func (e *event) Remove(names ...string) {
	e.Lock()
	defer e.Unlock()
	if len(names) > 0 {
		for _, name := range names {
			delete(e.events, name)
		}
		return
	}
	e.events = make(map[string][]interface{})
}
