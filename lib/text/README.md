# text Library

The text library provides a collection of utility functions for manipulating and processing text strings. It includes functions for case conversion, HTML processing, pattern matching, text parsing, and more.

## Installation

```go
import "github.com/getevo/evo/v2/lib/text"
```

## Features

- **Case Conversion**: Convert between camelCase, snake_case, kebab-case, and more
- **HTML Processing**: Convert HTML to plain text
- **Pattern Matching**: Match strings against patterns with wildcards
- **Text Parsing**: Parse text with wildcards, split strings, convert to JSON
- **URL Slugs**: Generate URL-friendly slugs from text
- **Random Strings**: Generate random strings of specified length

## Usage Examples

### Case Conversion

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/text"
)

func main() {
    // Convert to camelCase
    camel := text.UpperCamelCase("hello_world")
    fmt.Println(camel) // Output: HelloWorld
    
    camel = text.LowerCamelCase("hello-world")
    fmt.Println(camel) // Output: helloWorld
    
    // Convert to snake_case
    snake := text.SnakeCase("HelloWorld")
    fmt.Println(snake) // Output: hello_world
    
    snake = text.UpperSnakeCase("helloWorld")
    fmt.Println(snake) // Output: HELLO_WORLD
    
    // Convert to kebab-case
    kebab := text.KebabCase("HelloWorld")
    fmt.Println(kebab) // Output: hello-world
    
    kebab = text.UpperKebabCase("helloWorld")
    fmt.Println(kebab) // Output: HELLO-WORLD
}
```

### HTML Processing

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/text"
)

func main() {
    html := "<p>Hello <strong>World</strong>!</p><br/><p>This is a test.</p>"
    plainText := text.FromHTML(html)
    fmt.Println(plainText) // Output: Hello World!\nThis is a test.
}
```

### Pattern Matching and Parsing

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/text"
)

func main() {
    // Match a string against a pattern
    matched := text.Match("hello.txt", "*.txt")
    fmt.Println(matched) // Output: true
    
    matched = text.Match("hello.jpg", "*.txt")
    fmt.Println(matched) // Output: false
    
    // Parse a string with wildcards
    parts := text.ParseWildCard("user-123-profile.jpg", "user-*-*.jpg")
    fmt.Println(parts) // Output: [123 profile]
    
    // Split a string on any of the given separators
    parts = text.SplitAny("hello,world;foo:bar", ",;:")
    fmt.Printf("%q\n", parts) // Output: ["hello" "world" "foo" "bar"]
    
    // Convert a value to JSON
    jsonStr := text.ToJSON(map[string]interface{}{
        "name": "John",
        "age":  30,
    })
    fmt.Println(jsonStr) // Output: {"age":30,"name":"John"}
}
```

### URL Slugs and Random Strings

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/text"
)

func main() {
    // Generate a URL-friendly slug
    slug := text.Slugify("Hello World! This is a test.")
    fmt.Println(slug) // Output: hello-world-this-is-a-test
    
    // Generate a random string
    random := text.Random(10)
    fmt.Println(random) // Output: random 10-character string
}
```

## Available Functions

### Case Conversion
- `UpperCamelCase(s string) string`: Converts to CamelCase with uppercase first letter
- `LowerCamelCase(s string) string`: Converts to camelCase with lowercase first letter
- `SnakeCase(s string) string`: Converts to snake_case
- `UpperSnakeCase(s string) string`: Converts to UPPER_SNAKE_CASE
- `KebabCase(s string) string`: Converts to kebab-case
- `UpperKebabCase(s string) string`: Converts to KEBAB-CASE

### HTML Processing
- `FromHTML(html string) string`: Converts HTML to plain text

### Pattern Matching and Parsing
- `Match(input, pattern string) bool`: Checks if a string matches a pattern
- `ParseWildCard(input, expr string) []string`: Parses a string using a wildcard pattern
- `SplitAny(s string, seps string) []string`: Splits a string on any of the given separators
- `ToJSON(v any) string`: Converts a value to a JSON string

### URL Slugs and Random Strings
- `Slugify(text string) string`: Generates a URL-friendly slug
- `Random(length int) string`: Generates a random string of specified length

## How It Works

The text library provides a set of utility functions for common text manipulation tasks. These functions are designed to be simple to use while providing powerful functionality.

The case conversion functions use a combination of character-by-character processing and regular expressions to convert between different case formats. The HTML processing function uses regular expressions to remove HTML tags and convert line breaks to newlines.

The pattern matching and parsing functions build on the standard library's capabilities, adding support for wildcard matching and text extraction. The URL slug function normalizes text and removes special characters to create URL-friendly strings.

For more detailed information, please refer to the source code and comments within the library.