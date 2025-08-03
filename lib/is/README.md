# is Library

The is library provides a comprehensive set of validation functions for checking various types of data. It offers simple, easy-to-use functions that return boolean values indicating whether the input matches the expected format.

## Installation

```go
import "github.com/getevo/evo/v2/lib/is"
```

## Features

- **Data Type Validation**: Check if values are integers, floats, alphanumeric, etc.
- **Format Validation**: Validate emails, URLs, UUIDs, credit cards, etc.
- **Range Checking**: Verify if values fall within specified ranges
- **String Validation**: Check string length, case, content
- **Network Validation**: Validate IPs, MAC addresses, hostnames
- **Geographic Validation**: Check latitude/longitude coordinates
- **Document Validation**: Validate ISBN numbers, SSNs, etc.
- **Web Content Validation**: Check HTML safety, JSON validity

## Usage Examples

### Basic Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/is"
)

func main() {
    // Email validation
    fmt.Println(is.Email("user@example.com"))  // true
    fmt.Println(is.Email("invalid-email"))     // false
    
    // URL validation
    fmt.Println(is.URL("https://example.com")) // true
    fmt.Println(is.URL("not-a-url"))           // false
    
    // Numeric validation
    fmt.Println(is.Numeric("12345"))           // true
    fmt.Println(is.Numeric("123abc"))          // false
    
    // Integer validation
    fmt.Println(is.Int("42"))                  // true
    fmt.Println(is.Int("42.5"))                // false
    
    // Float validation
    fmt.Println(is.Float("42.5"))              // true
    fmt.Println(is.Float("not-a-float"))       // false
}
```

### Range and Type Checking

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/is"
)

func main() {
    // Range checking
    fmt.Println(is.InRange(5, 1, 10))          // true
    fmt.Println(is.InRange(15, 1, 10))         // false
    
    // String length checking
    fmt.Println(is.StringLength("hello", 3, 10)) // true
    fmt.Println(is.StringLength("hi", 3, 10))    // false
    
    // Case checking
    fmt.Println(is.UpperCase("HELLO"))         // true
    fmt.Println(is.UpperCase("Hello"))         // false
    
    fmt.Println(is.LowerCase("hello"))         // true
    fmt.Println(is.LowerCase("Hello"))         // false
    
    // Number type checking
    fmt.Println(is.Whole(5.0))                 // true
    fmt.Println(is.Whole(5.5))                 // false
    
    fmt.Println(is.Natural(5.0))               // true
    fmt.Println(is.Natural(-5.0))              // false
}
```

### Web and Network Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/is"
)

func main() {
    // IP address validation
    fmt.Println(is.IP("192.168.1.1"))          // true
    fmt.Println(is.IP("not-an-ip"))            // false
    
    fmt.Println(is.IPv4("192.168.1.1"))        // true
    fmt.Println(is.IPv6("2001:db8::1"))        // true
    
    // MAC address validation
    fmt.Println(is.MAC("01:23:45:67:89:ab"))   // true
    fmt.Println(is.MAC("not-a-mac"))           // false
    
    // JSON validation
    fmt.Println(is.JSON(`{"key": "value"}`))   // true
    fmt.Println(is.JSON("not-json"))           // false
    
    // HTML safety checking
    fmt.Println(is.SafeHTML("<p>Safe HTML</p>"))                // true
    fmt.Println(is.SafeHTML("<script>alert('XSS')</script>"))   // false
    
    // DNS name validation
    fmt.Println(is.DNSName("example.com"))     // true
    fmt.Println(is.DNSName("not a domain"))    // false
}
```

### Document and Format Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/is"
)

func main() {
    // Credit card validation
    fmt.Println(is.CreditCard("4111111111111111"))  // true (Visa)
    fmt.Println(is.CreditCard("invalid-card"))      // false
    
    // ISBN validation
    fmt.Println(is.ISBN10("0-306-40615-2"))         // true
    fmt.Println(is.ISBN13("978-3-16-148410-0"))     // true
    
    // UUID validation
    fmt.Println(is.UUID("550e8400-e29b-41d4-a716-446655440000"))  // true
    fmt.Println(is.UUIDv4("550e8400-e29b-41d4-a716-446655440000")) // false
    
    // Semantic versioning
    fmt.Println(is.Semver("v1.2.3"))               // true
    fmt.Println(is.Semver("1.2"))                  // false
    
    // Phone number validation
    fmt.Println(is.PhoneNumber("+1-555-123-4567")) // true
    fmt.Println(is.PhoneNumber("abc"))             // false
}
```

### Color Format Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/is"
)

func main() {
    // Hex color validation
    fmt.Println(is.HexColor("#FF5733"))        // true
    fmt.Println(is.HexColor("FF5733"))         // true
    fmt.Println(is.HexColor("#ZZ5733"))        // false
    
    // RGB color validation
    fmt.Println(is.RGBColor("rgb(255, 87, 51)"))  // true
    fmt.Println(is.RGBColor("not-rgb"))           // false
    
    // RGBA color validation
    fmt.Println(is.RGBAColor("rgba(255, 87, 51, 255)"))  // true
    fmt.Println(is.RGBAColor("not-rgba"))                // false
}
```

## How It Works

The is library provides a collection of functions that validate different types of data. Each function takes an input and returns a boolean value indicating whether the input matches the expected format.

The library uses a combination of techniques for validation:
- Regular expressions for pattern matching
- Built-in Go functions for type checking
- Custom algorithms for complex validations (like credit card validation)
- Standard libraries for network and URL parsing

Most validation functions are designed to be simple and focused on a specific validation task, making them easy to use and combine for complex validation requirements.

For more detailed information, please refer to the source code and comments within the library.