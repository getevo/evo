# Date Library

The Date library provides a convenient wrapper around Go's standard time.Time type with additional functionality for date manipulation, formatting, and calculations. It simplifies working with dates and times in your applications.

## Installation

```go
import "github.com/getevo/evo/v2/lib/date"
```

## Features

- **Date Creation**: Multiple ways to create dates from different sources (string, time.Time, Unix timestamp)
- **Natural Language Calculations**: Calculate relative dates using expressions like "next week" or "2 days after"
- **Time Differences**: Calculate durations between dates in various ways
- **Formatting Options**: Format dates using both Go's standard format and strftime syntax
- **Timestamp Conversion**: Easy conversion to Unix timestamps
- **Midnight Calculation**: Get the midnight (start of day) for any date

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/date"
)

func main() {
    // Get current date/time
    now := date.Now()
    fmt.Println("Current time:", now.Format("2006-01-02 15:04:05"))
    
    // Get midnight of today
    midnight := date.Now().Midnight()
    fmt.Println("Midnight:", midnight.Format("2006-01-02 15:04:05"))
    
    // Parse a date string
    d, err := date.FromString("2025-08-03 10:26")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Parsed date:", d.Format("2006-01-02 15:04:05"))
}
```

### Date Calculations

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/date"
)

func main() {
    // Get current date
    now := date.Now()
    
    // Calculate relative dates
    tomorrow, _ := now.Calculate("tomorrow")
    fmt.Println("Tomorrow:", tomorrow.Format("2006-01-02"))
    
    nextWeek, _ := now.Calculate("next week start")
    fmt.Println("Start of next week:", nextWeek.Format("2006-01-02"))
    
    twoMonthsLater, _ := now.Calculate("2 months")
    fmt.Println("Two months from now:", twoMonthsLater.Format("2006-01-02"))
    
    // Calculate time differences
    duration, _ := now.DiffExpr("2 days")
    fmt.Printf("48 hours in seconds: %d\n", duration.Seconds())
}
```

## How It Works

The Date library wraps Go's time.Time type and provides additional methods for common date operations. It uses the dateparse library for flexible string parsing and the strftime library for formatting options.

The core of the library is the Date struct, which contains a time.Time value. All operations are performed on this value, making it easy to chain methods together.

The Calculate method is particularly powerful, allowing for natural language expressions to calculate relative dates. It supports expressions like:
- "tomorrow", "yesterday", "today"
- "next week", "last month"
- "2 days after", "3 months before"
- "next year start", "this month start"

This makes it easy to perform common date calculations without having to manually manipulate the date components.
