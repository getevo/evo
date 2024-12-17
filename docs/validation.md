# Validation Library

This library provides a flexible and extensible validation framework for Go structs and values. It allows you to define validation rules through struct tags, validate structs or single values, and even extend the existing set of validators with your own custom logic.

## Table of Contents
1. [Defining Validation Tags](#defining-validation-tags)
    - [Basic Usage](#basic-usage)
    - [Multiple Validators](#multiple-validators)
2. [Built-in Validators](#built-in-validators)
    - [Non-Database Validators (Validators)](#non-database-validators-validators)
    - [Database-Related Validators (DBValidators)](#database-related-validators-dbvalidators)
3. [Validating Values](#validating-values)
    - [Validating Structs](#validating-structs)
    - [Validating Structs with Non-Zero Fields](#validating-structs-with-non-zero-fields)
    - [Validating Single Values](#validating-single-values)
4. [Extending Validators](#extending-validators)
    - [Adding a Simple Validator](#adding-a-simple-validator)

---

## Defining Validation Tags

The validation rules for struct fields are defined using the `validation` tag. Each field can have one or more validators separated by commas. For example:

```go
type User struct {
    Email string `validation:"required,email"`
    Age   int    `validation:">=18"`
}
```

Here, `Email` must be `required` and a valid `email`, and `Age` must be greater than or equal to `18`.

### Basic Usage

```go
type Product struct {
    Name  string `validation:"required,alpha"`
    Price string `validation:"+float"`
}
```

- `Name` must be non-empty and contain only letters.
- `Price` must be a positive float.

### Multiple Validators

You can chain multiple validators using commas:

```go
type Account struct {
    Username string `validation:"required,alphanumeric,len>=6"`
}
```

This means `Username` must be:
- `required` (cannot be empty),
- `alphanumeric` (only letters and digits),
- and have a length `>= 6` characters.

---

## Built-in Validators

### Non-Database Validators (Validators)

These validators do not interact with the database. They only check the value in memory.

| Validator              | Description                                                             | Example Usage                                |
|------------------------|-------------------------------------------------------------------------|----------------------------------------------|
| `text`                 | Ensures string contains no HTML tags.                                   | `validation:"text"`                          |
| `name`                 | Checks if the value is a valid name (letters, spaces, etc.).            | `validation:"name"`                          |
| `alpha`                | Only alphabetical characters allowed.                                   | `validation:"alpha"`                         |
| `latin`                | Only Unicode letters are allowed.                                       | `validation:"latin"`                         |
| `digit`                | Only digits [0-9] allowed.                                              | `validation:"digit"`                         |
| `alphanumeric`         | Letters, digits, and spaces allowed.                                     | `validation:"alphanumeric"`                  |
| `required`             | Value cannot be empty.                                                  | `validation:"required"`                      |
| `email`                | Checks for a valid email format.                                         | `validation:"email"`                         |
| `regex(...)`           | Matches value against a custom regex.                                    | `validation:"regex([a-z]{2,})"`              |
| `len<`, `len>`, etc.   | Compares string length. Supports `<`, `>`, `<=`, `>=`, `==`, `!=`.       | `validation:"len==10"`                       |
| Numeric comparisons     | Compares numeric value (`>`, `<`, `>=`, `<=`, etc.) with a given number.| `validation:">=18"`                          |
| `int`, `+int`, `-int`  | Checks if value is integer (`+` for positive, `-` for negative).         | `validation:"+int"`                         |
| `float`, `+float`, `-float` | Checks if value is float (`+` for positive, `-` for negative).     | `validation:"-float"`                        |
| `password(...)`         | Checks complexity (`none`, `easy`, `medium`, `hard`).                  | `validation:"password(medium)"`              |
| `domain`                | Valid domain format.                                                    | `validation:"domain"`                        |
| `url`                   | Valid URL format.                                                       | `validation:"url"`                           |
| `ip`, `ip4`, `ip6`      | Valid IP address (IPv4 or IPv6).                                        | `validation:"ip"`                           |
| `cidr`                  | Valid CIDR notation.                                                    | `validation:"cidr"`                         |
| `mac`                   | Valid MAC address.                                                      | `validation:"mac"`                          |
| `date`                  | Valid date in RFC3339 format.                                           | `validation:"date"`                         |
| `longitude`             | Valid longitude.                                                        | `validation:"longitude"`                    |
| `latitude`              | Valid latitude.                                                         | `validation:"latitude"`                     |
| `port`                  | Valid port number.                                                      | `validation:"port"`                        |
| `json`                  | Valid JSON format.                                                      | `validation:"json"`                        |
| `ISBN`, `ISBN10`, `ISBN13` | Checks if value is a valid ISBN.                                     | `validation:"ISBN13"`                       |
| `creditcard`            | Checks if the value is a valid credit card number.                      | `validation:"creditcard"`                   |
| `uuid`                  | Checks if the value is a valid UUID.                                    | `validation:"uuid"`                        |
| `uppercase`             | Checks if string is uppercase.                                          | `validation:"uppercase"`                   |
| `lowercase`             | Checks if string is lowercase.                                          | `validation:"lowercase"`                   |
| `rgbcolor`, `rgba`, `hexcolor`, `hex` | Validates various color formats.                          | `validation:"hexcolor"`                    |
| `countryalpha2`, `countryalpha3` | Valid ISO country code formats.                                | `validation:"countryalpha2"`               |
| `btcaddress`, `ethaddress` | Checks if value is a valid Bitcoin or Ethereum address.              | `validation:"btcaddress"`                  |
| `cron`                  | Valid CRON expression.                                                  | `validation:"cron"`                        |
| `duration`              | Valid Go duration format.                                               | `validation:"duration"`                    |
| `time`                  | Valid RFC3339 timestamp.                                                | `validation:"time"`                        |
| `unixTimestamp`          | Valid unix timestamp.                                                   | `validation:"unixTimestamp"`               |
| `timezone`              | Valid timezone string.                                                  | `validation:"timezone"`                    |
| `e164`                  | Valid E164 phone number format.                                         | `validation:"e164"`                       |
| `safeHTML`              | Checks string for possible XSS patterns.                                | `validation:"safeHTML"`                   |
| `noHTML`                | Ensures string does not contain HTML tags.                              | `validation:"noHTML"`                     |
| `phone`                 | Checks if string is a valid phone number.                               | `validation:"phone"`                      |

### Database-Related Validators (DBValidators)

These validators require database access and use GORM’s statement to validate against the DB.

| Validator           | Description                                                                | Example Usage                          |
|---------------------|----------------------------------------------------------------------------|----------------------------------------|
| `unique`            | Ensures the field value is unique in the table.                             | `validation:"unique"`                  |
| `unique:col1\|col2` | Ensures a combination of columns is unique.                                 | `validation:"unique:country,vat_number"` |
| `fk`                | Checks the field references a valid foreign key in another table.           | `validation:"fk"`                      |
| `enum`              | Checks that the value matches one of the allowed ENUM values in the schema. | `validation:"enum"`                    |
| `before(field)`     | Checks the timestamp is before another field’s timestamp.                   | `validation:"before(CreatedAt)"`       |
| `after(field)`      | Checks the timestamp is after another field’s timestamp.                    | `validation:"after(UpdatedAt)"`        |

---

## Validating Values

### Validating Structs

```go
import "github.com/getevo/evo/v2/lib/validation"

type User struct {
    Email string `validation:"required,email"`
    Age   int    `validation:">=18"`
}

user := User{Email: "john@example.com", Age: 20}
errs := validation.Struct(user)
if len(errs) > 0 {
    // Handle validation errors
}
```

If any field fails validation, you will receive errors describing the issues.

### Validating Structs with Non-Zero Fields

`StructNonZeroFields` only validates fields that are not at their zero value, useful for partial updates:

```go
partialUser := User{Age: 25} // Email is zero value
errs := validation.StructNonZeroFields(partialUser)
```

Only `Age` will be validated in this case.

### Validating Single Values

You can validate a standalone value:

```go
err := validation.Value("john@example.com", "required,email")
if err != nil {
    // Handle validation error
}
```

---

## Extending Validators

### Adding a Simple Validator

You can easily add a custom validator. For example, to add a `country` validator:

```go
import (
    "fmt"
    "regexp"
    "github.com/getevo/evo/v2/lib/validation"
    "github.com/getevo/evo/v2/lib/generic"
)

var CountryMap = map[string]string{
    "US": "United States",
    "CA": "Canada",
    // ...other countries
}

// Register a new validator
validation.Validators[regexp.MustCompile("^country$")] = countryValidator

func countryValidator(match []string, value *generic.Value) error {
    var v = value.String()
    if value.Input == nil || v == "" || v == "<nil>" {
        return nil
    }
    if _, ok := CountryMap[v]; !ok {
        return fmt.Errorf("invalid country %s", v)
    }
    return nil
}
```

Now you can use `validation:"country"` in your struct tags or `validation.Value()` calls.


### Extending DBValidators

To extend `DBValidators` for custom database-related validations, you can add a new entry to the `DBValidators` map. Here's an example of adding a custom validator that ensures a field value exists in a specific database table:

```go
import (
    "context"
    "fmt"
    "regexp"
    "github.com/getevo/evo/v2/lib/validation"
    "github.com/getevo/evo/v2/lib/generic"
    "gorm.io/gorm"
)

// Register a new DBValidator
validation.DBValidators[regexp.MustCompile("^exists:(.+)$")] = existsValidator

func existsValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
    tableName := match[1] // Extract table name from validator, e.g., "exists:users"
    var count int64
    if err := stmt.DB.Table(tableName).Where(fmt.Sprintf("%s = ?", field.DBName), value.Input).Count(&count).Error; err != nil {
        return fmt.Errorf("database error: %s", err)
    }
    if count == 0 {
        return fmt.Errorf("value does not exist in table %s", tableName)
    }
    return nil
}
```

### Usage Example:

```go
type User struct {
    RoleID int `validation:"exists:roles"`
}
```

In this example, the `RoleID` field must reference an existing ID in the `roles` table.
