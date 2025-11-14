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

### HTTP Status Functions

The library provides convenient helper functions for common HTTP status codes. All functions accept optional input and intelligently handle different data types:

#### Success Responses (2xx)

```go
// 200 OK
outcome.StatusOk()                                    // Default: "OK"
outcome.StatusOk("Operation successful")              // String
outcome.StatusOk(user)                                // Struct (JSON marshaled)

// 201 Created
outcome.Created()                                     // Default: "Created"
outcome.Created(newResource)                          // Return created resource

// 202 Accepted
outcome.Accepted()                                    // Default: "Accepted"
outcome.Accepted(map[string]string{"id": "12345"})    // Custom data

// 204 No Content
outcome.NoContent()                                   // Default: empty
outcome.NoContent("Done")                             // Optional message
```

#### Client Error Responses (4xx)

```go
// 400 Bad Request
outcome.BadRequest()                                  // Default: "Bad Request"
outcome.BadRequest("Invalid input parameters")        // Custom message

// 401 Unauthorized
outcome.UnAuthorized()                                // Default: "Unauthorized"
outcome.UnAuthorized("Invalid credentials")           // Custom message

// 404 Not Found
outcome.NotFound()                                    // Default: "Not Found"
outcome.NotFound("User not found")                    // Custom message

// 406 Not Acceptable
outcome.NotAcceptable()                               // Default: "Not Acceptable"
outcome.NotAcceptable("Unsupported media type")       // Custom message

// 408 Request Timeout
outcome.RequestTimeout()                              // Default: "Request Timeout"
outcome.RequestTimeout("Request took too long")       // Custom message

// 429 Too Many Requests
outcome.TooManyRequests()                             // Default: "Too Many Requests"
outcome.TooManyRequests("Rate limit exceeded")        // Custom message

// 451 Unavailable For Legal Reasons
outcome.UnavailableForLegalReasons()                  // Default: "Unavailable For Legal Reasons"
outcome.UnavailableForLegalReasons("Content blocked") // Custom message
```

#### Server Error Responses (5xx)

```go
// 500 Internal Server Error
outcome.InternalServerError()                         // Default: "Internal Server Error"
outcome.InternalServerError("Database connection failed") // Custom message
outcome.InternalServerError(errorDetails)             // Struct with error details
```

#### Type Handling

All status functions intelligently process input data:

- **string**: Returns as-is
- **[]byte**: Returns as-is
- **struct/map/slice/array**: JSON marshaled
- **other types**: Converted using `fmt.Sprint`
- **no input**: Returns default message

```go
// String input
outcome.BadRequest("Invalid email format")

// Byte slice
outcome.StatusOk([]byte("Success"))

// Struct (automatically JSON marshaled)
type ErrorDetails struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
outcome.BadRequest(ErrorDetails{Field: "email", Message: "invalid format"})

// Map (automatically JSON marshaled)
outcome.NotFound(map[string]any{
    "error": "user_not_found",
    "id": 123,
})

// Numbers or other types (converted with fmt.Sprint)
outcome.InternalServerError(12345)
```

## How It Works

The outcome library is built around the `Response` struct, which encapsulates all the data needed for an HTTP response. The library provides factory functions like `Text()`, `Html()`, and `Json()` to create responses with specific content types, as well as methods to customize the response.

When used with the EVO framework, these responses can be returned directly from handlers and will be properly rendered to the client.

For more detailed information, please refer to the source code and comments within the library.