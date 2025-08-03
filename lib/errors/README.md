# errors Library

The errors library provides HTTP error handling functionality with predefined error types and customizable error responses.

## Installation

```go
import "github.com/getevo/evo/v2/lib/errors"
```

## Features

- **HTTP Error Handling**: Create standardized HTTP error responses
- **Predefined Error Types**: Common HTTP errors like NotFound, BadRequest, Unauthorized, etc.
- **Customizable**: Set custom error messages and status codes
- **JSON Serialization**: Errors are automatically serialized to JSON

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/errors"
)

func main() {
    // Create a basic error
    err := errors.New("Something went wrong")
    fmt.Println(err) // Internal Server Error with status code 500
    
    // Create an error with custom status code
    err = errors.New("Not Found", 404)
    
    // Use predefined errors
    err = errors.NotFound
    err = errors.BadRequest
    err = errors.Unauthorized
    
    // Change status code of an existing error
    err = errors.Internal.Code(503) // Changes status code to 503
}
```

### Available Predefined Errors

```go
errors.NotFound                    // 404 Not Found
errors.BadRequest                  // 400 Bad Request
errors.Unauthorized                // 401 Unauthorized
errors.Forbidden                   // 403 Forbidden
errors.Internal                    // 500 Internal Server Error
errors.NotAcceptable               // 406 Not Acceptable
errors.Conflict                    // 409 Conflict
errors.Precondition                // 412 Precondition Failed
errors.UnsupportedMedia            // 415 Unsupported Media Type
errors.Gone                        // 410 Gone
errors.RequestTimeout              // 408 Request Timeout
errors.RequestEntityTooLarge       // 413 Request Entity Too Large
errors.RequestURITooLong           // 414 Request URI Too Long
errors.RequestHeaderFieldsTooLarge // 431 Request Header Fields Too Large
errors.UnavailableForLegalReasons  // 451 Unavailable For Legal Reasons
errors.PayloadTooLarge             // 413 Payload Too Large
errors.TooManyRequests             // 429 Too Many Requests
```

## How It Works

The errors library creates standardized HTTP error responses with appropriate status codes. Each error is represented as an HTTPError type that includes a status code and error message. The library automatically serializes errors to JSON format for easy integration with HTTP responses.

For more detailed information, please refer to the source code and comments within the library.