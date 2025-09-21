# Getting Started with EVO Framework

This guide provides step-by-step instructions to help you get started with the EVO Framework. By following these instructions, you'll set up a basic EVO application and learn the core concepts of the framework.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Creating Your First EVO Application](#creating-your-first-evo-application)
4. [Understanding the Project Structure](#understanding-the-project-structure)
5. [Routing and Request Handling](#routing-and-request-handling)
6. [Working with Databases](#working-with-databases)
7. [Validation](#validation)
8. [Error Handling](#error-handling)
9. [Configuration Management](#configuration-management)
10. [Next Steps](#next-steps)

## Prerequisites

Before you begin, ensure you have the following installed:

- Go 1.16 or later
- Git
- A code editor (VS Code, GoLand, etc.)
- MySQL, PostgreSQL, or SQLite (if you plan to use a database)

## Installation

1. **Create a new Go module for your project:**

```bash
mkdir my-evo-app
cd my-evo-app
go mod init github.com/yourusername/my-evo-app
```

2. **Add EVO Framework as a dependency:**

```bash
go get github.com/getevo/evo/v2
```

3. **Create a main.go file:**

```bash
touch main.go
```

## Creating Your First EVO Application

Open the `main.go` file in your editor and add the following code:

```go
package main

import (
    "github.com/getevo/evo/v2"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Define a simple route
    evo.Get("/", func(request *evo.Request) interface{} {
        return "Hello, EVO Framework!"
    })
    
    // Run the server
    evo.Run()
}
```

Run your application:

```bash
go run main.go
```

Open your browser and navigate to `http://localhost:8080`. You should see "Hello, EVO Framework!" displayed.

## Understanding the Project Structure

For a well-organized EVO application, consider the following project structure:

```
my-evo-app/
├── config/             # Configuration files
│   └── settings.json   # Main configuration file
├── controllers/        # Request handlers
├── models/             # Data models
├── migrations/         # Database migrations
├── public/             # Static assets
├── views/              # Templates
├── main.go             # Application entry point
└── go.mod              # Go module file
```

Let's create this structure:

```bash
mkdir -p config controllers models migrations public views
```

Create a basic configuration file in `config/settings.json`:

```json
{
    "app": {
        "name": "My EVO App",
        "debug": true,
        "port": 8080
    },
    "database": {
        "driver": "mysql",
        "host": "localhost",
        "port": 3306,
        "username": "root",
        "password": "",
        "database": "evo_app",
        "charset": "utf8mb4"
    }
}
```

## Routing and Request Handling

EVO Framework uses a simple and intuitive routing system. Let's create a more structured application with controllers:

1. **Create a controller file** in `controllers/home_controller.go`:

```go
package controllers

import (
    "github.com/getevo/evo/v2"
)

// HomeController handles home-related routes
type HomeController struct{}

// Index is the handler for the home page
func (c *HomeController) Index(request *evo.Request) interface{} {
    return "Welcome to My EVO App!"
}

// About is the handler for the about page
func (c *HomeController) About(request *evo.Request) interface{} {
    return "About My EVO App"
}
```

2. **Update your main.go file** to use the controller:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/yourusername/my-evo-app/controllers"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Create a new instance of HomeController
    home := &controllers.HomeController{}
    
    // Define routes using the controller
    evo.Get("/", home.Index)
    evo.Get("/about", home.About)
    
    // Run the server
    evo.Run()
}
```

3. **Run your application** and visit `http://localhost:8080` and `http://localhost:8080/about` to see the different responses.

## Working with Databases

EVO Framework provides a powerful database layer. Let's set up a simple model and database operations:

1. **Create a model** in `models/user.go`:

```go
package models

import (
    "github.com/getevo/evo/v2/lib/db"
)

// User represents a user in the system
type User struct {
    ID        uint   `json:"id" gorm:"primaryKey"`
    Name      string `json:"name" validation:"required"`
    Email     string `json:"email" validation:"required,email"`
    CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`
}
```

2. **Create a migration** in `migrations/001_create_users_table.go`:

```go
package migrations

import (
    "github.com/getevo/evo/v2/lib/db"
    "github.com/yourusername/my-evo-app/models"
)

// CreateUsersTable creates the users table
func CreateUsersTable() {
    db.AutoMigrate(&models.User{})
}
```

3. **Update your main.go file** to run the migration:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/yourusername/my-evo-app/controllers"
    "github.com/yourusername/my-evo-app/migrations"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Run migrations
    migrations.CreateUsersTable()
    
    // Create a new instance of HomeController
    home := &controllers.HomeController{}
    
    // Define routes using the controller
    evo.Get("/", home.Index)
    evo.Get("/about", home.About)
    
    // Run the server
    evo.Run()
}
```

4. **Create a users controller** in `controllers/users_controller.go`:

```go
package controllers

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/validation"
    "github.com/yourusername/my-evo-app/models"
)

// UsersController handles user-related routes
type UsersController struct{}

// List returns all users
func (c *UsersController) List(request *evo.Request) interface{} {
    var users []models.User
    db.Find(&users)
    return users
}

// Get returns a specific user
func (c *UsersController) Get(request *evo.Request) interface{} {
    id := request.Param("id").Int()
    var user models.User
    
    if db.First(&user, id).Error != nil {
        return request.Status(404).JSON(map[string]string{
            "error": "User not found",
        })
    }
    
    return user
}

// Create creates a new user
func (c *UsersController) Create(request *evo.Request) interface{} {
    var user models.User
    request.BodyParser(&user)
    
    errors := validation.Struct(user)
    if len(errors) > 0 {
        return request.Status(400).JSON(map[string]interface{}{
            "errors": errors,
        })
    }
    
    db.Create(&user)
    return user
}
```

5. **Update your main.go file** to add the user routes:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/yourusername/my-evo-app/controllers"
    "github.com/yourusername/my-evo-app/migrations"
)

func main() {
    // Initialize EVO
    evo.Setup()
    
    // Run migrations
    migrations.CreateUsersTable()
    
    // Create controller instances
    home := &controllers.HomeController{}
    users := &controllers.UsersController{}
    
    // Define routes
    evo.Get("/", home.Index)
    evo.Get("/about", home.About)
    
    // User routes
    evo.Get("/users", users.List)
    evo.Get("/users/:id", users.Get)
    evo.Post("/users", users.Create)
    
    // Run the server
    evo.Run()
}
```

## Validation

EVO Framework includes a powerful validation library. We've already used it in our User model with the `validation` tags. Let's see how to use it more extensively:

```go
// In controllers/users_controller.go

// Update updates an existing user
func (c *UsersController) Update(request *evo.Request) interface{} {
    id := request.Param("id").Int()
    var user models.User
    
    if db.First(&user, id).Error != nil {
        return request.Status(404).JSON(map[string]string{
            "error": "User not found",
        })
    }
    
    // Parse the request body into the user
    request.BodyParser(&user)
    
    // Validate the user
    errors := validation.Struct(user)
    if len(errors) > 0 {
        return request.Status(400).JSON(map[string]interface{}{
            "errors": errors,
        })
    }
    
    // Save the updated user
    db.Save(&user)
    return user
}
```

Add the new route to your main.go file:

```go
evo.Put("/users/:id", users.Update)
```

## Error Handling

EVO Framework provides structured error handling with the `try` library. Let's see how to use it:

```go
// In controllers/users_controller.go

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/panics"
    "github.com/getevo/evo/v2/lib/try"
    "github.com/getevo/evo/v2/lib/validation"
    "github.com/yourusername/my-evo-app/models"
)

// Delete deletes a user
func (c *UsersController) Delete(request *evo.Request) interface{} {
    id := request.Param("id").Int()
    
    var result interface{}
    
    try.This(func() {
        var user models.User
        
        if db.First(&user, id).Error != nil {
            panic("User not found")
        }
        
        db.Delete(&user)
        result = map[string]string{
            "message": "User deleted successfully",
        }
    }).Catch(func(recovered *panics.Recovered) {
        result = request.Status(404).JSON(map[string]string{
            "error": recovered.Value.(string),
        })
    })
    
    return result
}
```

Add the new route to your main.go file:

```go
evo.Delete("/users/:id", users.Delete)
```

## Configuration Management

EVO Framework provides a settings library for configuration management. Let's see how to use it:

1. **Update your main.go file** to load settings from a file:

```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/settings"
    "github.com/yourusername/my-evo-app/controllers"
    "github.com/yourusername/my-evo-app/migrations"
)

func main() {
    // Load settings from file
    settings.Load("config/settings.json")
    
    // Initialize EVO
    evo.Setup()
    
    // Run migrations
    migrations.CreateUsersTable()
    
    // Create controller instances
    home := &controllers.HomeController{}
    users := &controllers.UsersController{}
    
    // Define routes
    evo.Get("/", home.Index)
    evo.Get("/about", home.About)
    
    // User routes
    evo.Get("/users", users.List)
    evo.Get("/users/:id", users.Get)
    evo.Post("/users", users.Create)
    evo.Put("/users/:id", users.Update)
    evo.Delete("/users/:id", users.Delete)
    
    // Run the server with the configured port
    port := settings.Get("app.port").Int()
    evo.Run(port)
}
```

2. **Access settings in your code**:

```go
// In controllers/home_controller.go

// Index is the handler for the home page
func (c *HomeController) Index(request *evo.Request) interface{} {
    appName := settings.Get("app.name").String()
    return "Welcome to " + appName + "!"
}
```

## Next Steps

Congratulations! You've created a basic EVO Framework application with routing, database operations, validation, error handling, and configuration management. Here are some next steps to explore:

1. **Templates**: Learn how to use the template system to render HTML views.
2. **Middleware**: Add middleware for authentication, logging, and other cross-cutting concerns.
3. **File Uploads**: Implement file upload functionality using the storage library.
4. **API Documentation**: Create interactive API documentation using Swagger or ReDoc.
5. **Testing**: Write tests for your controllers and models.

For more information, refer to the following resources:

- [EVO Framework Documentation](https://github.com/getevo/evo/tree/master/docs)
- [Library Documentation](https://github.com/getevo/evo/tree/master/lib)
- [Example Applications](https://github.com/getevo/evo/tree/master/dev/app)