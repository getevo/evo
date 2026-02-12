# DB Library

The DB library provides a comprehensive wrapper around GORM (Go Object Relational Mapper) for database operations in the EVO Framework. It simplifies database interactions and provides additional functionality for schema management, migrations, and more.

## Installation

```go
import "github.com/getevo/evo/v2/lib/db"
```

## Features

The DB library offers a wide range of features:

- **Connection Management**: Functions for managing database connections
- **CRUD Operations**: Simple functions for creating, reading, updating, and deleting records
- **Query Building**: Methods for constructing complex database queries
- **Transaction Support**: Functions for handling database transactions
- **Schema Management**: Tools for managing database schemas and migrations
- **Model Registration**: Ability to register and manage database models

## Subdirectories

The DB library includes several subdirectories for specialized functionality:

- **entity**: Provides base entity structures and functionality
- **schema**: Tools for schema management and migrations (DB-agnostic)
- **types**: Custom data types for database interactions

## Usage Examples

### Basic CRUD Operations

```go
package main

import (
    "github.com/getevo/evo/v2/lib/db"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

func main() {
    // Register models
    db.UseModel(&User{})
    
    // Create a new user
    user := User{Name: "John Doe", Age: 30}
    db.Create(&user)
    
    // Find a user
    var foundUser User
    db.First(&foundUser, user.ID)
    
    // Update a user
    db.Model(&foundUser).Update("Age", 31)
    
    // Delete a user
    db.Delete(&foundUser)
}
```

### Using Transactions

```go
package main

import (
    "github.com/getevo/evo/v2/lib/db"
)

func main() {
    // Start a transaction
    err := db.Transaction(func(tx *gorm.DB) error {
        // Perform operations within the transaction
        if err := tx.Create(&User{Name: "User 1"}).Error; err != nil {
            // Return error will rollback the transaction
            return err
        }
        
        if err := tx.Create(&User{Name: "User 2"}).Error; err != nil {
            return err
        }
        
        // Return nil will commit the transaction
        return nil
    })
    
    if err != nil {
        // Handle error
    }
}
```

### Schema Migrations

```go
package main

import (
    "github.com/getevo/evo/v2/lib/db"
)

func main() {
    // Register models
    db.UseModel(&User{})
    
    // Get migration script
    scripts := db.GetMigrationScript()
    
    // Or perform migration directly
    err := db.DoMigration()
    if err != nil {
        // Handle error
    }
}
```

## Advanced Query Building

```go
package main

import (
    "github.com/getevo/evo/v2/lib/db"
)

func main() {
    var users []User
    
    // Complex query with conditions, ordering, and limits
    db.Where("age > ?", 18).
       Where("name LIKE ?", "%Doe%").
       Order("age DESC").
       Limit(10).
       Find(&users)
       
    // Using scopes for reusable query parts
    db.Scopes(ActiveUsers, AgeGreaterThan(18)).Find(&users)
}

// Scope example
func ActiveUsers(d *gorm.DB) *gorm.DB {
    return d.Where("active = ?", true)
}

func AgeGreaterThan(age int) func(*gorm.DB) *gorm.DB {
    return func(d *gorm.DB) *gorm.DB {
        return d.Where("age > ?", age)
    }
}
```

## Related Libraries

- **entity**: Base entity structures and functionality
- **schema**: Schema management and migrations
- **types**: Custom data types for database interactions