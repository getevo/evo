# Validation

EVO provides a comprehensive validation system via `lib/validation`. Validators are defined using struct tags and support both standalone (no DB required) and database-aware rules.

## Import

```go
import "github.com/getevo/evo/v2/lib/validation"
```

## Basic usage

Add a `validation` tag to struct fields, then call `validation.Struct`:

```go
type CreateUserInput struct {
    Name     string `json:"name"     validation:"required,name,len>=2,len<=100"`
    Email    string `json:"email"    validation:"required,email"`
    Password string `json:"password" validation:"required,password(medium)"`
    Age      int    `json:"age"      validation:">=18,<=120"`
}

input := CreateUserInput{
    Name:     "Alice",
    Email:    "alice@example.com",
    Password: "MyPass1!",
    Age:      25,
}

if errs := validation.Struct(input); len(errs) > 0 {
    for _, e := range errs {
        fmt.Println(e)
    }
}
```

### Validate only non-zero fields (partial update)

```go
// Only validates fields that have a non-zero value
errs := validation.StructNonZeroFields(input)
```

### Validate a single value

```go
err := validation.Value("alice@example.com", "required,email")
if err != nil {
    fmt.Println(err)
}
```

### Context-aware validation (with timeout)

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

errs := validation.StructWithContext(ctx, input)
errs2 := validation.StructNonZeroFieldsWithContext(ctx, input)
```

## In HTTP handlers

```go
evo.Post("/api/users", func(r *evo.Request) any {
    var input CreateUserInput
    r.BodyParser(&input)

    if errs := validation.Struct(input); len(errs) > 0 {
        return outcome.UnprocessableEntity(errs)
    }

    return outcome.Created(createUser(input))
})
```

## Built-in validators

Multiple validators are combined with commas: `validation:"required,email,len<=255"`

### String validators

| Validator | Description |
|---|---|
| `required` | Field must be non-empty |
| `text` | No HTML tags allowed |
| `alpha` | Letters and spaces only |
| `alphanumeric` | Letters, digits, and spaces only |
| `latin` | Unicode letters only |
| `digit` | Digits only (`0-9`) |
| `slug` | Lowercase letters, digits, hyphens, underscores (1–200 chars) |
| `name` | Valid personal name (`letters, spaces, . ' -`) |
| `ascii` | ASCII characters only |
| `printable` | Printable characters only |
| `upperCase` | Must be all uppercase |
| `lowerCase` | Must be all lowercase |
| `no_html` | Must not contain HTML tags |
| `safe_html` | Must not contain XSS-dangerous HTML |

```go
type Post struct {
    Slug    string `validation:"required,slug"`
    Title   string `validation:"required,text,len>=3,len<=200"`
    Content string `validation:"safe_html"`
}
```

### Length validators

| Validator | Description |
|---|---|
| `len>N` | Length strictly greater than N |
| `len>=N` | Length greater than or equal to N |
| `len<N` | Length strictly less than N |
| `len<=N` | Length less than or equal to N |
| `len==N` | Exact length N |
| `len!=N` | Length not equal to N |

Length is measured in Unicode runes (not bytes).

```go
type Input struct {
    Username string `validation:"len>=3,len<=20"`
    Bio      string `validation:"len<=500"`
    Code     string `validation:"len==6"`
}
```

### Numerical validators

| Validator | Description |
|---|---|
| `>N` | Greater than N |
| `>=N` | Greater than or equal to N |
| `<N` | Less than N |
| `<=N` | Less than or equal to N |
| `==N` | Equal to N |
| `!=N` | Not equal to N |
| `int` | Valid integer |
| `+int` | Positive integer |
| `-int` | Negative integer |
| `float` | Valid float |
| `+float` | Positive float |
| `-float` | Negative float |

```go
type Product struct {
    Price    float64 `validation:">0"`
    Quantity int     `validation:">=0,<=10000"`
    Rating   int     `validation:">=1,<=5"`
}
```

### Internet / network validators

| Validator | Description |
|---|---|
| `email` | Valid email address |
| `url` | Valid URL |
| `domain` | Valid domain name |
| `ip` or `ipv4` | Valid IPv4 address |
| `ipv6` | Valid IPv6 address |
| `cidr` | Valid CIDR notation (`192.168.1.0/24`) |
| `mac` | Valid MAC address (`AA:BB:CC:DD:EE:FF`) |
| `port` | Valid port number (1–65535) |
| `phone` | Valid phone number |
| `e164` | Valid E.164 phone number (`+15551234567`) |

```go
type Server struct {
    IP      string `validation:"required,ipv4"`
    Port    int    `validation:"required,port"`
    Email   string `validation:"required,email"`
    Website string `validation:"url"`
}
```

### Date / time validators

| Validator | Description |
|---|---|
| `date` | Valid RFC3339 date/time |
| `time` | Valid RFC3339 timestamp |
| `unix_timestamp` | Valid Unix timestamp integer |
| `duration` | Valid Go duration string (`1h30m`, `500ms`) |
| `timezone` | Valid IANA timezone (`America/New_York`) |
| `before_now` | Date must be in the past |
| `after_now` | Date must be in the future |
| `date_format(layout)` | Date matches the Go layout string |

```go
type Event struct {
    StartAt    string `validation:"required,date,after_now"`
    EndAt      string `validation:"required,date"`
    Duration   string `validation:"duration"`
    Timezone   string `validation:"timezone"`
    BirthDate  string `validation:"before_now"`
    ReportDate string `validation:"date_format(2006-01-02)"`
}
```

### Password validator

| Validator | Requirement |
|---|---|
| `password(none)` | No requirements |
| `password(easy)` | Minimum 6 characters |
| `password(medium)` | Min 8 chars, at least 3 of: uppercase, lowercase, digits, symbols |
| `password(hard)` | Min 12 chars, all 4: uppercase, lowercase, digits, symbols |

```go
type Registration struct {
    Password string `validation:"required,password(medium)"`
}
```

### Regex validator

```go
type Input struct {
    Code     string `validation:"regex([A-Z]{2}[0-9]{4})"`
    PostCode string `validation:"regex(^[0-9]{5}(-[0-9]{4})?$)"`
}
```

### Inclusion / exclusion validators

| Validator | Description |
|---|---|
| `in(a,b,c)` | Value must be one of the listed values |
| `not_in(a,b,c)` | Value must not be any of the listed values |
| `contains(str)` | Value must contain the substring |
| `not_contains(str)` | Value must not contain the substring |
| `starts_with(str)` | Value must start with the prefix |
| `ends_with(str)` | Value must end with the suffix |

```go
type Order struct {
    Status    string `validation:"in(pending,processing,shipped,delivered)"`
    Priority  string `validation:"not_in(low,none)"`
    Reference string `validation:"starts_with(ORD-)"`
    Extension string `validation:"ends_with(.pdf)"`
    Notes     string `validation:"not_contains(spam)"`
}
```

### Array / slice validators

| Validator | Description |
|---|---|
| `min_items(N)` | Slice must have at least N elements |
| `max_items(N)` | Slice must have at most N elements |
| `unique_items` | All slice elements must be unique |

```go
type Form struct {
    Tags       []string `validation:"min_items(1),max_items(10),unique_items"`
    Recipients []string `validation:"min_items(1),max_items(50)"`
}
```

### Format validators

| Validator | Description |
|---|---|
| `json` | Valid JSON string |
| `uuid` | Valid UUID |
| `hex` | Valid hexadecimal string |
| `hex_color` | Valid hex color (`#RRGGBB`) |
| `rgb_color` | Valid RGB color (`rgb(r,g,b)`) |
| `rgba_color` | Valid RGBA color (`rgba(r,g,b,a)`) |
| `isbn` | Valid ISBN (10 digit) |
| `isbn10` | Valid ISBN-10 |
| `isbn13` | Valid ISBN-13 |
| `credit_card` | Valid credit card number (Luhn check) |
| `iban` | Valid IBAN format |
| `btc_address` | Valid Bitcoin address |
| `eth_address` | Valid Ethereum address |
| `country_alpha2` | Valid ISO 3166-1 alpha-2 country code |
| `country_alpha3` | Valid ISO 3166-1 alpha-3 country code |
| `cron` | Valid cron expression |
| `latitude` | Valid latitude (-90 to 90) |
| `longitude` | Valid longitude (-180 to 180) |

```go
type Payment struct {
    CardNumber string `validation:"required,credit_card"`
    IBAN       string `validation:"iban"`
    Currency   string `validation:"required,in(USD,EUR,GBP)"`
}

type Location struct {
    Lat float64 `validation:"latitude"`
    Lng float64 `validation:"longitude"`
}
```

## Database validators

These validators require a database connection. They are skipped silently when no database is configured.

| Validator | Description |
|---|---|
| `unique` | Value must be unique in the table column |
| `unique:col1\|col2` | Combination of columns must be unique |
| `fk` | Value must exist in the referenced table (via `gorm:"fk:table"` tag) |
| `enum` | Value must match enum values in the `gorm` tag |

```go
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex" validation:"required,email,unique"`
    Role  string `gorm:"type:enum('admin','user','guest')" validation:"required,enum"`
}

type Order struct {
    ID        uint `gorm:"primaryKey"`
    UserID    uint `gorm:"fk:users"    validation:"required,fk"`
    ProductID uint `gorm:"fk:products" validation:"required,fk"`
}
```

### Unique with composite columns

```go
type TeamMember struct {
    ID     uint `gorm:"primaryKey"`
    TeamID uint
    UserID uint `validation:"unique:team_id|user_id"` // unique together
}
```

## Cross-field validators (DB)

These compare fields within the same struct and require a database connection.

| Validator | Description |
|---|---|
| `confirmed` | Must match `{FieldName}Confirmation` field |
| `same(Field)` | Must equal the named field |
| `different(Field)` | Must differ from the named field |
| `before(Field)` | Time must be before the named field |
| `after(Field)` | Time must be after the named field |
| `gt_field(Field)` | Numeric: must be greater than the named field |
| `gte_field(Field)` | Numeric: must be ≥ the named field |
| `lt_field(Field)` | Numeric: must be less than the named field |
| `lte_field(Field)` | Numeric: must be ≤ the named field |

```go
type ChangePassword struct {
    Password             string `validation:"required,password(medium)"`
    PasswordConfirmation string // auto-checked by "confirmed"
    OldPassword          string `validation:"different(Password)"`
}

type DateRange struct {
    StartDate string `validation:"required,date"`
    EndDate   string `validation:"required,date,after(StartDate)"`
}

type PriceRange struct {
    MinPrice float64 `validation:"required,>0"`
    MaxPrice float64 `validation:"required,gt_field(MinPrice)"`
}
```

## Custom validators

### Register a custom validator

```go
import (
    "fmt"
    "strings"
    "github.com/getevo/evo/v2/lib/validation"
    "github.com/getevo/evo/v2/lib/generic"
)

func init() {
    validation.RegisterValidator(`^uppercase_required$`, func(match []string, value *generic.Value) error {
        v := value.String()
        if v != strings.ToUpper(v) {
            return fmt.Errorf("must be uppercase")
        }
        return nil
    })
}

type Config struct {
    Region string `validation:"required,uppercase_required"`
}
```

### Custom validator with parameters

```go
// matches: divisible_by(10), divisible_by(5), etc.
validation.RegisterValidator(`^divisible_by\((\d+)\)$`, func(match []string, value *generic.Value) error {
    n, _ := strconv.Atoi(match[1])
    if n == 0 {
        return fmt.Errorf("divisor cannot be zero")
    }
    if value.Int()%n != 0 {
        return fmt.Errorf("must be divisible by %d", n)
    }
    return nil
})

type Batch struct {
    Count int `validation:"divisible_by(10)"`
}
```

### Custom DB validator

```go
import (
    "github.com/getevo/evo/v2/lib/validation"
    "github.com/getevo/evo/v2/lib/db"
    "gorm.io/gorm"
    "gorm.io/gorm/schema"
)

validation.RegisterDBValidator(
    `^active_user$`,
    func(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
        var count int64
        db.Table("users").Where("id = ? AND active = ?", value.Input, true).Count(&count)
        if count == 0 {
            return fmt.Errorf("user must be active")
        }
        return nil
    },
)

type Post struct {
    AuthorID uint `validation:"required,fk,active_user"`
}
```

## Error format

Errors are returned as `[]error`. Each message includes the field name (from the `json` tag or struct field name) and the reason:

```
email invalid email
password password must contain at least 3 of: uppercase, lowercase, digits, symbols
age is smaller than or equal to 18
tags must have at least 1 items
```

## Full example

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/getevo/evo/v2/lib/validation"
)

type RegisterRequest struct {
    Username             string   `json:"username"   validation:"required,slug,len>=3,len<=30"`
    Email                string   `json:"email"      validation:"required,email,unique"`
    Password             string   `json:"password"   validation:"required,password(medium)"`
    PasswordConfirmation string   `json:"password_confirmation"`
    Age                  int      `json:"age"        validation:"required,>=13,<=120"`
    Country              string   `json:"country"    validation:"required,country_alpha2"`
    Website              string   `json:"website"    validation:"url"`
    Tags                 []string `json:"tags"       validation:"max_items(5),unique_items"`
    BirthDate            string   `json:"birth_date" validation:"date,before_now"`
}

func main() {
    req := RegisterRequest{
        Username:             "alice_dev",
        Email:               "alice@example.com",
        Password:            "MyPass1!",
        PasswordConfirmation: "MyPass1!",
        Age:                 25,
        Country:             "US",
        Website:             "https://alice.dev",
        Tags:                []string{"go", "backend"},
        BirthDate:           "1999-06-15T00:00:00Z",
    }

    if errs := validation.Struct(req); len(errs) > 0 {
        for _, e := range errs {
            fmt.Println("Error:", e)
        }
        return
    }
    fmt.Println("Validation passed!")

    // Partial update — only validate non-zero fields
    partial := RegisterRequest{Email: "new@example.com"}
    if errs := validation.StructNonZeroFields(partial); len(errs) > 0 {
        for _, e := range errs {
            fmt.Println("Error:", e)
        }
    }

    // Single value
    if err := validation.Value("+15551234567", "e164"); err != nil {
        fmt.Println(err)
    }

    // With context timeout (for DB validators)
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    errs := validation.StructWithContext(ctx, req)
    _ = errs
}
```

## See Also

- [Web Server](webserver.md)
- [Outcome](outcome.md)
- [Database](database.md)
