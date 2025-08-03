# Validation Library

The Validation library provides comprehensive data validation capabilities for Go applications. It offers a wide range of validators for different data types and formats, making it easy to validate user input, API requests, and database records.

## Installation

```go
import "github.com/getevo/evo/v2/lib/validation"
```

## Features

- **Struct Validation**: Validate entire structs using field tags
- **Individual Value Validation**: Validate single values against specific rules
- **Database Integration**: Validate against database constraints like uniqueness and foreign keys
- **Comprehensive Validators**: Over 50 built-in validators for various data types and formats
- **Custom Error Messages**: Detailed error messages for validation failures
- **Tag-Based Configuration**: Simple configuration using struct tags
- **Nested Struct Support**: Validate nested structs and slices of structs
- **Conditional Validation**: Apply validators conditionally based on other field values

## Validator Categories

### Basic Validators
- `required`: Field must not be empty
- `text`: Text without HTML tags
- `alpha`: Only alphabetic characters
- `alphanumeric`: Only alphanumeric characters
- `digit`: Only digits
- `email`: Valid email address
- `url`: Valid URL
- `domain`: Valid domain name
- `regex(pattern)`: Match against a regular expression

### Numeric Validators
- `int`, `+int`, `-int`: Integer values (optional sign constraints)
- `float`, `+float`, `-float`: Floating-point values (optional sign constraints)
- `>N`, `<N`, `>=N`, `<=N`, `==N`, `!=N`: Numerical comparisons
- `len>N`, `len<N`, `len>=N`, `len<=N`, `len==N`, `len!=N`: Length comparisons

### Format Validators
- `date`: Valid date in RFC3339 format
- `time`: Valid timestamp
- `timezone`: Valid timezone
- `duration`: Valid duration format
- `json`: Valid JSON
- `uuid`: Valid UUID
- `ip`, `ipv4`, `ipv6`: Valid IP addresses
- `mac`: Valid MAC address
- `cidr`: Valid CIDR notation
- `phone`: Valid phone number
- `e164`: Valid E.164 phone number
- `credit-card`: Valid credit card number
- `isbn`, `isbn10`, `isbn13`: Valid ISBN numbers

### Database Validators
- `unique`: Value must be unique in the database
- `fk`: Value must match a foreign key
- `enum`: Value must be one of the enumerated values
- `before(field)`: Date must be before another field's date
- `after(field)`: Date must be after another field's date

### Special Validators
- `password(level)`: Password with specified complexity level (easy, medium, hard)
- `slug`: Valid URL slug
- `name`: Valid person name
- `upperCase`: All uppercase
- `lowerCase`: All lowercase
- `hex-color`, `rgb-color`, `rgba-color`: Valid color formats
- `country-alpha-2`, `country-alpha-3`: Valid country codes
- `btc-address`, `eth-address`: Valid cryptocurrency addresses

## Usage Examples

### Struct Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/validation"
)

type User struct {
    Name     string `json:"name" validation:"required,name"`
    Email    string `json:"email" validation:"required,email"`
    Password string `json:"password" validation:"required,password(medium)"`
    Age      int    `json:"age" validation:">=18"`
    Website  string `json:"website" validation:"url"`
}

func main() {
    user := User{
        Name:     "John Doe",
        Email:    "invalid-email",
        Password: "weak",
        Age:      16,
        Website:  "example",
    }
    
    errors := validation.Struct(user)
    
    if len(errors) > 0 {
        fmt.Println("Validation errors:")
        for _, err := range errors {
            fmt.Println("-", err)
        }
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### Individual Value Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/validation"
)

func main() {
    // Validate an email address
    err := validation.Value("user@example.com", "email")
    if err != nil {
        fmt.Println("Email validation failed:", err)
    } else {
        fmt.Println("Email is valid")
    }
    
    // Validate a password with medium complexity
    err = validation.Value("password123", "password(medium)")
    if err != nil {
        fmt.Println("Password validation failed:", err)
    } else {
        fmt.Println("Password is valid")
    }
}
```

### Nested Struct Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/validation"
)

type Address struct {
    Street  string `json:"street" validation:"required"`
    City    string `json:"city" validation:"required,alpha"`
    ZipCode string `json:"zip_code" validation:"required,digit,len==5"`
    Country string `json:"country" validation:"required,country-alpha-2"`
}

type Customer struct {
    Name    string  `json:"name" validation:"required,name"`
    Email   string  `json:"email" validation:"required,email"`
    Address Address `json:"address" validation:"required"`
    Orders  []Order `json:"orders" validation:"required,len>=1"`
}

type Order struct {
    ID        string  `json:"id" validation:"required,uuid"`
    Amount    float64 `json:"amount" validation:"required,>0"`
    CreatedAt string  `json:"created_at" validation:"required,date"`
}

func main() {
    customer := Customer{
        Name:  "Jane Smith",
        Email: "jane@example.com",
        Address: Address{
            Street:  "123 Main St",
            City:    "New York",
            ZipCode: "10001",
            Country: "US",
        },
        Orders: []Order{
            {
                ID:        "550e8400-e29b-41d4-a716-446655440000",
                Amount:    99.99,
                CreatedAt: "2025-08-03T11:24:00Z",
            },
        },
    }
    
    errors := validation.Struct(customer)
    
    if len(errors) > 0 {
        fmt.Println("Validation errors:")
        for _, err := range errors {
            fmt.Println("-", err)
        }
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### Conditional Validation

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/validation"
)

type PaymentMethod struct {
    Type          string `json:"type" validation:"required,enum(credit_card,bank_transfer,paypal)"`
    CardNumber    string `json:"card_number" validation:"required_if(type=credit_card),credit-card"`
    ExpiryDate    string `json:"expiry_date" validation:"required_if(type=credit_card),regex(^(0[1-9]|1[0-2])/[0-9]{2}$)"`
    CVV           string `json:"cvv" validation:"required_if(type=credit_card),digit,len==3"`
    AccountNumber string `json:"account_number" validation:"required_if(type=bank_transfer),digit,len>=8,len<=12"`
    BankCode      string `json:"bank_code" validation:"required_if(type=bank_transfer),alphanumeric,len==8"`
    PayPalEmail   string `json:"paypal_email" validation:"required_if(type=paypal),email"`
}

func main() {
    // Credit card payment
    ccPayment := PaymentMethod{
        Type:       "credit_card",
        CardNumber: "4111111111111111",
        ExpiryDate: "12/25",
        CVV:        "123",
    }
    
    errors := validation.Struct(ccPayment)
    if len(errors) > 0 {
        fmt.Println("Credit card validation errors:")
        for _, err := range errors {
            fmt.Println("-", err)
        }
    } else {
        fmt.Println("Credit card validation passed!")
    }
    
    // Bank transfer payment
    bankPayment := PaymentMethod{
        Type:          "bank_transfer",
        AccountNumber: "12345678",
        BankCode:      "ABCD1234",
    }
    
    errors = validation.Struct(bankPayment)
    if len(errors) > 0 {
        fmt.Println("Bank transfer validation errors:")
        for _, err := range errors {
            fmt.Println("-", err)
        }
    } else {
        fmt.Println("Bank transfer validation passed!")
    }
}
```

## Advanced Usage

### Custom Validation Messages

You can provide custom error messages for validation failures using the `message` tag:

```go
type Product struct {
    Name  string  `json:"name" validation:"required" message:"Product name is required"`
    Price float64 `json:"price" validation:"required,>0" message:"Price must be greater than zero"`
    SKU   string  `json:"sku" validation:"required,regex(^[A-Z]{2}-\d{4}$)" message:"SKU must be in format XX-0000"`
}
```

### Handling Validation Errors

The validation library returns a list of error strings. You can process these errors to create a structured response:

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/getevo/evo/v2/lib/validation"
    "strings"
)

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ValidationResponse struct {
    Success bool              `json:"success"`
    Errors  []ValidationError `json:"errors,omitempty"`
}

func processValidationErrors(errors []string) []ValidationError {
    var result []ValidationError
    
    for _, err := range errors {
        parts := strings.SplitN(err, ":", 2)
        if len(parts) == 2 {
            result = append(result, ValidationError{
                Field:   strings.TrimSpace(parts[0]),
                Message: strings.TrimSpace(parts[1]),
            })
        } else {
            result = append(result, ValidationError{
                Message: err,
            })
        }
    }
    
    return result
}

func main() {
    // ... validation code ...
    
    errors := validation.Struct(user)
    
    response := ValidationResponse{
        Success: len(errors) == 0,
    }
    
    if len(errors) > 0 {
        response.Errors = processValidationErrors(errors)
    }
    
    jsonResponse, _ := json.MarshalIndent(response, "", "  ")
    fmt.Println(string(jsonResponse))
}
```

### Database Validation Examples

When using database validators, the library integrates with the database to perform validations:

```go
type Product struct {
    ID        uint   `json:"id"`
    Name      string `json:"name" validation:"required"`
    SKU       string `json:"sku" validation:"required,unique(products)"`
    CategoryID uint   `json:"category_id" validation:"required,fk(categories.id)"`
}
```

In this example:
- The `unique(products)` validator checks that the SKU is unique in the products table
- The `fk(categories.id)` validator checks that the category_id exists in the categories table

## Best Practices

1. **Use Appropriate Validators**: Choose validators that match your data requirements. Combine multiple validators when necessary.

2. **Validate Early**: Validate input data as early as possible in your application flow to prevent invalid data from propagating.

3. **Provide Clear Error Messages**: Use custom error messages to provide clear guidance to users about validation failures.

4. **Group Related Validations**: For complex validation scenarios, consider creating custom validators that encapsulate related validation rules.

5. **Test Edge Cases**: Ensure your validation logic handles edge cases correctly, such as empty strings, zero values, and special characters.

6. **Consider Performance**: For high-volume applications, be mindful of the performance impact of database validators.

7. **Use Conditional Validation**: Apply validators conditionally based on other field values to create flexible validation rules.

## How It Works

The validation library uses struct tags to define validation rules for each field. When you call `validation.Struct()`, it examines each field with a "validation" tag and applies the specified validators.

For database-related validations, the library integrates with the database layer to perform checks against the database, such as uniqueness constraints and foreign key validations.

The validation process returns a list of errors, one for each field that failed validation. Each error includes the field name and a description of why the validation failed.

You can also validate individual values using the `validation.Value()` function, which applies the specified validators to a single value.

The library is designed to be extensible, allowing you to add custom validators if needed.
