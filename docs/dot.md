# dot package
The dot package provides functionality for accessing and manipulating map, struct, and array values using dot notation.

## Usage

To use this package, import it in your Go code:
```go
import "github.com/getevo/evo/v2/lib/dot"
```

### Functions
#### Get
```go
func Get(obj interface{}, prop string) (interface{}, error)
```
The **`Get`** function retrieves the value of a property specified by **`prop`** from the given **`obj`**. It supports accessing properties in maps, structs, and arrays using dot notation.

- **`obj`**: The object from which to retrieve the property value.
- **`prop`**: The property to retrieve, specified in dot notation (e.g., "nested.property").
Returns the value of the property and nil if the property exists and can be retrieved successfully. If the property does not exist, both the return value and the error will be nil. If an error occurs during retrieval, the function returns nil and an error describing the issue.

#### Set
```go
func Set(input interface{}, prop string, value interface{}) error
```

The `Set` function sets the value of a property specified by `prop` in the given `input` object. It supports setting properties in maps, structs, and arrays using dot notation.

- **`input`**: The object in which to set the property value.
-  **`prop`**: The property to set, specified in dot notation (e.g., "nested.property").
-  **`value`**: The value to set for the property.

Returns nil if the property is set successfully. If an error occurs during property setting, it returns an error describing the issue.

## Usage Example
Here's an example demonstrating the usage of the **`Get`** and **`Set`** functions:
```go
package main

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/dot" // Replace with your actual package import path
)

func main() {
	// Example using Get
	obj := map[string]interface{}{
		"foo": "bar",
	}
	result, err := dot.Get(obj, "foo")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Value:", result)
	}

	// Example using Set
	obj2 := map[string]interface{}{
		"nested": map[string]interface{}{
			"property": "old value",
		},
	}
	err = dot.Set(obj2, "nested.property", "new value")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Updated Object:", obj2)
	}
}
```