# Backend Code Guideline for AI Agents

## 1. Framework
Use [`github.com/getevo/evo/v2`](https://github.com/getevo/evo) for backend development.

## 2. HTTP Requests
```go
import "github.com/getevo/evo/v2/lib/curl"
```

## 3. Settings
```go
import "github.com/getevo/evo/v2/lib/settings"
settings.Get("APP1.SETTING_NAME", "default").String()
settings.Get("APP2.SETTING_NAME", 5.6).Float64()
```

## 4. Database
```go
db.Create(Model{...})
```

## 5. Logging
```go
import "github.com/getevo/evo/v2/lib/log"
log.Info("message")
log.Error("error", err)
```

## 6. Middleware
```go
evo.Use("/api", func(request *evo.Request) interface{} {  
    return request.Next()
})
```

## 7. Project Structure
```
main.go
apps/
  app1/
    app.go         # Init & routing
    models.go      # DB models
    controller.go  # API logic
    serializer.go  # API/data structs
    functions.go   # Helpers
library/
  lib1/            # Common libs
submodule/
  sub1/            # Git submodules
static/            # Static files
config.yml
config.{env}.yml   # Env config
Dockerfile
DockerfileDev      # Dev Dockerfile
docs/              # Docs
```

## 8. API Endpoints
Use [`github.com/getevo/restify`](https://github.com/getevo/restify) for auto CRUD (create, update, delete, list, filter).

## 9. Validation & JSON Utility
[Validation Docs](https://github.com/getevo/evo/blob/master/docs/validation.md)
```go
import (
  "github.com/getevo/evo/v2/lib/validation"
  "github.com/getevo/json"
)

// Struct validation with field tags
type User struct {
  Name     string `json:"name" validation:"required,name"`
  Email    string `json:"email" validation:"required,email"`
  Password string `json:"password" validation:"required,password(medium)"`
  Age      int    `json:"age" validation:">=18"`
}

// Validate struct
errors := validation.Struct(user)
if len(errors) > 0 {
  // Handle validation errors
}

// Individual value validation
err := validation.Value("user@example.com", "email")

// Common validators:
// - required: Field must not be empty
// - email, url, domain: Format validation
// - >N, <N, >=N, <=N: Numeric comparisons
// - len>N, len<N: Length comparisons
// - password(level): Password complexity
// - unique: Database uniqueness check
// - regex(pattern): Pattern matching
```

## 10. Pagination
```go
import "github.com/getevo/pagination"

func (c Controller) ListOrders(request *evo.Request) any {
  var items []models.Order
  model := db.Model(&models.Order{})
  page, err := pagination.New(model, request, &items, pagination.Options{MaxSize: 20})
  if err != nil { log.Error(err) }
  return page
}
```

## 11. Error Handling with Try/Catch
```go
import (
  "github.com/getevo/evo/v2/lib/try"
  "github.com/getevo/evo/v2/lib/panics"
)

// Execute code that might panic
try.This(func() {
  // Risky code here
  result := riskyOperation()
}).Catch(func(recovered *panics.Recovered) {
  // Handle the panic
  log.Error("Operation failed", recovered.Value)
}).Finally(func() {
  // Cleanup code that runs regardless of panic
  cleanup()
})
```

## 12. Template Rendering
```go
import "github.com/getevo/evo/v2/lib/tpl"

// Render template with variables
template := "Hello, $name! Welcome to $app."
result := tpl.Render(template, map[string]interface{}{
  "name": "User",
  "app": "EVO Framework",
})
```

## 13. Command-line Arguments
```go
import "github.com/getevo/evo/v2/lib/args"

// Check if an argument exists
if args.Exists("--debug") {
  // Enable debug mode
}

// Get argument value
configPath := args.Get("--config")
```

## 14. Application Management
```go
import "github.com/getevo/evo/v2/lib/application"

// Define your application
type MyApp struct{}

func (app MyApp) Register() error {
  // Initialize your application
  return nil
}

func (app MyApp) Router() error {
  // Set up routes
  return nil
}

func (app MyApp) WhenReady() error {
  // Execute when all applications are ready
  return nil
}

func (app MyApp) Name() string {
  return "myapp"
}

// In main.go
app := application.GetInstance()
app.Register(MyApp{})
app.Run()
```

## 15. Version Handling
```go
import "github.com/getevo/evo/v2/lib/version"

// Compare versions
isNewer := version.Compare("2.0.0", "1.5.0", ">")

// Check version constraints
constraint := version.NewConstrainGroupFromString(">=1.0.0,<2.0.0")
matches := constraint.Match("1.5.0")
```

## 16. Storage Interface
```go
import "github.com/getevo/evo/v2/lib/storage"

// Create storage instance
fsStorage, _ := storage.NewStorageInstance("local", "fs:///path/to/directory")

// Get storage by name
driver := storage.GetStorage("local")

// File operations
driver.Write("example.txt", "Hello, World!")
content, _ := driver.ReadAllString("example.txt")
```

## 17. File and Path Handling
```go
import "github.com/getevo/evo/v2/lib/gpath"

// File operations
gpath.Write("example.txt", "Hello, World!")
content, _ := gpath.ReadFile("example.txt")

// Path utilities
parent := gpath.Parent("/path/to/file.txt")
exists := gpath.IsFileExist("config.yml")
```

See [Webserver Docs](https://github.com/getevo/evo/blob/master/docs/webserver.md) for full webserver features detail.

## 18. Minimum Config
```yaml
Database:
  Cache: "false"
  ConnMaxLifTime: 1h
  Database: "database"
  Debug: "3"
  Enabled: true
  MaxIdleConns: "10"
  MaxOpenConns: "100"
  Params: "parseTime=true"
  Password: "***"
  SSLMode: false
  Server: host:3306
  SlowQueryThreshold: 500ms
  Type: mysql
  Username: root

HTTP:
  BodyLimit: 10mb
  CaseSensitive: false
  CompressedFileSuffix: .evo.gz
  Concurrency: 1024
  DisableDefaultContentType: false
  DisableDefaultDate: false
  DisableHeaderNormalizing: false
  DisableKeepalive: false
  ETag: false
  GETOnly: false
  Host: 0.0.0.0
  IdleTimeout: 0
  Immutable: false
  Network: ""
  Port: 8080
  Prefork: false
  ProxyHeader: X-Forwarded-For
  ReadBufferSize: 10mb
  ReadTimeout: 1s
  ReduceMemoryUsage: false
  ServerHeader: EVO
  StrictRouting: false
  UnescapePath: false
  EnablePrintRoutes: false
  WriteBufferSize: 4kb
  WriteTimeout: 5s
```

