# outcome Library

The outcome library provides a flexible way to handle HTTP responses in the EVO framework. It allows you to create, customize, and manage different types of HTTP responses with a fluent API.

## Installation

```go
import "github.com/getevo/evo/v2/lib/outcome"
```

## Features

- **Response Types**: Create Text, HTML, JSON, and Redirect responses
- **Status Codes**: Set custom HTTP status codes
- **Headers**: Add custom HTTP headers
- **Cookies**: Manage cookies with fine-grained control
- **Error Handling**: Easily create error responses
- **Cache Control**: Set cache control headers

## Usage Examples

### Basic Usage

```go
package main

import (
    "github.com/getevo/evo/v2/lib/outcome"
)

func main() {
    // Create a text response
    textResponse := outcome.Text("Hello, World!")
    
    // Create an HTML response
    htmlResponse := outcome.Html("<h1>Hello, World!</h1>")
    
    // Create a JSON response
    jsonResponse := outcome.Json(map[string]string{"message": "Hello, World!"})
    
    // Create a redirect response
    redirectResponse := outcome.Redirect("/new-location", 302)
    
    // Permanent redirect
    permanentRedirect := outcome.RedirectPermanent("/new-location")
    
    // Temporary redirect
    temporaryRedirect := outcome.RedirectTemporary("/new-location")
}
```

### Chaining Methods

```go
package main

import (
    "github.com/getevo/evo/v2/lib/outcome"
    "time"
)

func main() {
    // Create a response with chained methods
    response := outcome.Json(map[string]string{"message": "Success"}).
        Status(200).
        Header("X-Custom-Header", "CustomValue").
        Cookie("session", "abc123", "Path=/", "HttpOnly", time.Hour*24).
        SetCacheControl(time.Hour)
        
    // Create an error response
    errorResponse := outcome.Text("Not Found").
        Status(404).
        Error("Resource not found")
}
```

### Setting Content

```go
package main

import (
    "github.com/getevo/evo/v2/lib/outcome"
)

func main() {
    // Create a response and set its content
    response := outcome.Text("Initial content")
    
    // Change the content
    response.Content("Updated content")
    
    // Set a filename for download responses
    response.Filename("data.txt")
}
```

## How It Works

The outcome library is built around the `Response` struct, which encapsulates all the data needed for an HTTP response. The library provides factory functions like `Text()`, `Html()`, and `Json()` to create responses with specific content types, as well as methods to customize the response.

When used with the EVO framework, these responses can be returned directly from handlers and will be properly rendered to the client.

For more detailed information, please refer to the source code and comments within the library.