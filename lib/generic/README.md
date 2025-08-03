# generic Library

The generic library provides powerful utilities for working with generic values and type conversions in Go. It allows you to wrap any value and easily convert it between different types, access properties, and perform various operations regardless of the underlying type.

## Installation

```go
import "github.com/getevo/evo/v2/lib/generic"
```

## Features

- **Type Conversion**: Convert between different types (string, int, float, bool, time, etc.)
- **Property Access**: Access and manipulate properties of structs and maps
- **Type Checking**: Check types and compare types between values
- **Size Utilities**: Convert between different size formats (KB, MB, GB, etc.)
- **Serialization**: JSON and YAML marshaling/unmarshaling
- **Database Integration**: Implements database/sql interfaces
- **Reflection Utilities**: Work with reflection in a simplified way

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/generic"
)

func main() {
    // Parse any value
    value := generic.Parse("123")
    
    // Convert to different types
    fmt.Println(value.Int())    // 123
    fmt.Println(value.String()) // "123"
    fmt.Println(value.Bool())   // true
    
    // Parse a size string
    size := generic.Parse("5MB")
    fmt.Println(size.SizeInBytes()) // 5242880
    
    // Format bytes as human-readable size
    bytes := generic.Parse(5242880)
    fmt.Println(bytes.ByteCount()) // "5.0 MB"
}
```

### Working with Structs and Maps

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/generic"
)

type Person struct {
    Name    string
    Age     int
    Address Address
}

type Address struct {
    City    string
    Country string
}

func main() {
    // Create a struct
    person := Person{
        Name: "John",
        Age:  30,
        Address: Address{
            City:    "New York",
            Country: "USA",
        },
    }
    
    // Access properties
    p := generic.Parse(person)
    fmt.Println(p.Prop("Name").String())                // "John"
    fmt.Println(p.Prop("Age").Int())                    // 30
    fmt.Println(p.Prop("Address").Prop("City").String()) // "New York"
    
    // Modify properties
    personPtr := &Person{}
    generic.Parse(personPtr).SetProp("Name", "Alice")
    generic.Parse(personPtr).SetProp("Age", 25)
    fmt.Println(personPtr.Name) // "Alice"
    fmt.Println(personPtr.Age)  // 25
    
    // Work with maps
    m := map[string]interface{}{
        "name": "Bob",
        "age":  40,
    }
    
    mp := generic.Parse(m)
    fmt.Println(mp.Prop("name").String()) // "Bob"
    fmt.Println(mp.Prop("age").Int())     // 40
    
    // Set map property
    generic.Parse(&m).SetProp("city", "London")
    fmt.Println(m["city"]) // "London"
}
```

### Type Checking

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/generic"
    "reflect"
)

func main() {
    // Check types
    str := generic.Parse("hello")
    num := generic.Parse(123)
    
    fmt.Println(str.IsAny(reflect.String))        // true
    fmt.Println(num.IsAny(reflect.Int, reflect.String)) // true
    
    // Check if types are the same
    fmt.Println(str.SameAs("world"))  // true
    fmt.Println(num.SameAs(int64(0))) // false
    
    // Check if value is nil or empty
    nilVal := generic.Parse(nil)
    fmt.Println(nilVal.IsNil())   // true
    fmt.Println(nilVal.IsEmpty()) // true
}
```

### Time and Duration Handling

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/generic"
    "time"
)

func main() {
    // Parse time strings
    timeStr := generic.Parse("2023-01-15T15:30:45Z")
    t, _ := timeStr.Time()
    fmt.Println(t.Year()) // 2023
    
    // Parse duration strings
    durStr := generic.Parse("1h30m")
    d, _ := durStr.Duration()
    fmt.Println(d.Minutes()) // 90
}
```

## How It Works

The generic library works by wrapping any value in a `Value` struct, which provides methods to convert the value to different types and perform various operations. It uses reflection to inspect and manipulate the underlying value, making it easy to work with values of unknown or dynamic types.

The library handles type conversions automatically, attempting to convert between compatible types and providing sensible defaults when conversion isn't possible. It also provides utilities for working with structs and maps, allowing you to access and modify properties regardless of the underlying structure.

For more detailed information, please refer to the source code and comments within the library.