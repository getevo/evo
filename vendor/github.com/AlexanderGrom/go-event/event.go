// Package event is a simple event system.
package event

type (
	// Eventer interface
	Eventer interface {
		StopPropagation()
		IsPropagationStopped() bool
	}

	// Event is the base class for classes containing event data
	Event struct {
		stopped bool
	}

	// handle aliase
	handle = func(Eventer) error
)

// StopPropagation Stops the propagation of the event to further event listeners
func (e *Event) StopPropagation() {
	e.stopped = true
}

// IsPropagationStopped returns whether further event listeners should be triggered
func (e *Event) IsPropagationStopped() bool {
	return e.stopped
}
