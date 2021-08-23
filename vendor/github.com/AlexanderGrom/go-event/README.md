# go-event
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexanderGrom/go-event)](https://goreportcard.com/report/github.com/AlexanderGrom/go-event) [![GoDoc](https://godoc.org/github.com/AlexanderGrom/go-event?status.svg)](https://godoc.org/github.com/AlexanderGrom/go-event)

Go-event is a simple event system.

## Get the package
```bash
$ go get -u github.com/AlexanderGrom/go-event
```

## Examples
```go
e := event.New()
e.On("my.event.name.1", func() error {
    fmt.Println("Fire event")
    return nil
})

e.On("my.event.name.2", func(text string) error {
    fmt.Println("Fire", text)
    return nil
})

e.On("my.event.name.3", func(i, j int) error {
    fmt.Println("Fire", i+j)
    return nil
})

e.On("my.event.name.4", func(name string, params ...string) error {
    fmt.Println(name, params)
    return nil
})

e.Go("my.event.name.1") // Print: Fire event
e.Go("my.event.name.2", "some event") // Print: Fire some event
e.Go("my.event.name.3", 1, 2) // Print: Fire 3
e.Go("my.event.name.4", "params:", "a", "b", "c") // Print: params: [a b c]
```

A couple more examples
```go
package main

import (
    "fmt"

    "github.com/AlexanderGrom/go-event"
)

func EventFunc(text string) error {
    fmt.Println("Fire:", text, "1")
    return nil
}

type EventStruct struct{}

func (e *EventStruct) EventFunc(text string) error {
    fmt.Println("Fire:", text, "2")
    return nil
}

func main() {
    event.On("my.event.name.1", EventFunc)
    event.On("my.event.name.1", (&EventStruct{}).EventFunc)

    event.Go("my.event.name.1", "event")
    // Print: Fire event 1
    // Print: Fire event 2
}
```

### A more complex example

```go
package main

import (
    "fmt"

    "github.com/AlexanderGrom/go-event"
)

type (
    Listener interface {
        Name() string
        Handle(e event.Eventer) error
    }
)

// ...

type (
    fooEvent struct {
        event.Event

        i, j int
    }

    fooListener struct {
        name string
    }
)

func NewFooEvent(i, j int) event.Eventer {
    return &fooEvent{i:i, j:j}
}

func NewFooListener() Listener {
    return &fooListener{
        name: "my.foo.event",
    }
}

func (l *fooListener) Name() string {
    return l.name
}

func (l *fooListener) Handle(e event.Eventer) error {
    ev := e.(*fooEvent)
    ev.StopPropagation()

    fmt.Println("Fire", ev.i+ev.j)
    return nil
}

// ...

func main() {
    e := event.New()

    // Collection
    collect := []Listener{
        NewFooListener(),
        // ...
    }

    // ...

    // Registration
    for _, l := range collect {
        e.On(l.Name(), l.Handle)
    }

    // ...

    // Call
    e.Go("my.foo.event", NewFooEvent(1, 2))
    // Print: Fire 3
}
```