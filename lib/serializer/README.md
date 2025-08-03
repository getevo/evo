# serializer Library

The serializer library provides a consistent interface for marshaling and unmarshaling data in different formats. It simplifies working with various serialization methods by providing a common API.

## Installation

```go
import "github.com/getevo/evo/v2/lib/serializer"
```

## Features

- **Common Interface**: Unified API for different serialization formats
- **JSON Serialization**: Built-in support for JSON marshaling and unmarshaling
- **Binary Serialization**: Built-in support for binary marshaling and unmarshaling
- **Custom Serializers**: Create your own serializers with custom marshal/unmarshal functions
- **Type Safety**: Proper error handling for serialization failures

## Usage Examples

### Using Built-in Serializers

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/serializer"
)

type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Address string `json:"address"`
}

func main() {
    // Create a person
    person := Person{
        Name:    "John Doe",
        Age:     30,
        Address: "123 Main St",
    }
    
    // Serialize to JSON
    jsonData, err := serializer.JSON.Marshal(person)
    if err == nil {
        fmt.Printf("JSON: %s\n", jsonData)
    }
    
    // Deserialize from JSON
    var decodedPerson Person
    err = serializer.JSON.Unmarshal(jsonData, &decodedPerson)
    if err == nil {
        fmt.Printf("Decoded: %+v\n", decodedPerson)
    }
    
    // Serialize to binary
    binaryData, err := serializer.Binary.Marshal(person)
    if err == nil {
        fmt.Printf("Binary data length: %d bytes\n", len(binaryData))
    }
    
    // Deserialize from binary
    var decodedFromBinary Person
    err = serializer.Binary.Unmarshal(binaryData, &decodedFromBinary)
    if err == nil {
        fmt.Printf("Decoded from binary: %+v\n", decodedFromBinary)
    }
}
```

### Creating a Custom Serializer

```go
package main

import (
    "encoding/xml"
    "fmt"
    "github.com/getevo/evo/v2/lib/serializer"
)

type Product struct {
    ID    int    `xml:"id"`
    Name  string `xml:"name"`
    Price float64 `xml:"price"`
}

func main() {
    // Create a custom XML serializer
    xmlSerializer := serializer.New(
        // Marshal function
        func(v any) ([]byte, error) {
            return xml.Marshal(v)
        },
        // Unmarshal function
        func(data []byte, v any) error {
            return xml.Unmarshal(data, v)
        },
    )
    
    // Create a product
    product := Product{
        ID:    101,
        Name:  "Laptop",
        Price: 999.99,
    }
    
    // Serialize to XML
    xmlData, err := xmlSerializer.Marshal(product)
    if err == nil {
        fmt.Printf("XML: %s\n", xmlData)
    }
    
    // Deserialize from XML
    var decodedProduct Product
    err = xmlSerializer.Unmarshal(xmlData, &decodedProduct)
    if err == nil {
        fmt.Printf("Decoded: %+v\n", decodedProduct)
    }
}
```

### Using Serializers with Different Types

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/serializer"
)

func main() {
    // Serialize primitive types
    intValue := 42
    intData, _ := serializer.JSON.Marshal(intValue)
    fmt.Printf("Serialized int: %s\n", intData)
    
    // Serialize maps
    mapValue := map[string]interface{}{
        "name": "John",
        "age":  30,
        "hobbies": []string{
            "reading",
            "coding",
            "hiking",
        },
    }
    mapData, _ := serializer.JSON.Marshal(mapValue)
    fmt.Printf("Serialized map: %s\n", mapData)
    
    // Deserialize into map
    var decodedMap map[string]interface{}
    serializer.JSON.Unmarshal(mapData, &decodedMap)
    fmt.Printf("Decoded map: %+v\n", decodedMap)
    
    // Serialize slices
    sliceValue := []int{1, 2, 3, 4, 5}
    sliceData, _ := serializer.JSON.Marshal(sliceValue)
    fmt.Printf("Serialized slice: %s\n", sliceData)
}
```

## How It Works

The serializer library is built around the `Interface` struct, which contains two function fields:

```go
type Interface struct {
    Marshal   func(v any) ([]byte, error)
    Unmarshal func(data []byte, v any) error
}
```

The library provides two predefined serializers:

1. `JSON`: Uses the standard library's JSON marshaling and unmarshaling functions
2. `Binary`: Uses binary marshaling and unmarshaling for efficient binary serialization

You can also create custom serializers using the `New` function, which takes custom marshal and unmarshal functions.

This design allows you to use different serialization formats through a consistent interface, making it easy to switch between formats or create new ones as needed.

For more detailed information, please refer to the source code and comments within the library.