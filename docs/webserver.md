# Web Server

EVO uses **[Fiber v3](https://github.com/gofiber/fiber)** as its web framework, which in turn uses **[fasthttp](https://github.com/valyala/fasthttp)** under the hood. EVO wraps the Fiber context into a `*Request` object that adds extra helpers, structured responses, and integration with other EVO subsystems.

---

## Setup & Run

```go
package main

import "github.com/getevo/evo/v2"

func main() {
    if err := evo.Setup(); err != nil {
        panic(err)
    }

    evo.Get("/hello", func(r *evo.Request) any {
        return "hello world"
    })

    if err := evo.Run(); err != nil {
        panic(err)
    }
}
```

---

## HTTP Methods

All route registration functions accept one or more `Handler` functions:

```go
type Handler func(request *evo.Request) any
```

```go
evo.Get(path, handlers...)
evo.Post(path, handlers...)
evo.Put(path, handlers...)
evo.Patch(path, handlers...)
evo.Delete(path, handlers...)
evo.Head(path, handlers...)
evo.Options(path, handlers...)
evo.Trace(path, handlers...)
evo.Connect(path, handlers...)
evo.All(path, handlers...)     // all HTTP methods
```

All functions return `fiber.Router` for further chaining (e.g. `.Name("my-route")`).

---

## Middleware

```go
type Middleware func(request *evo.Request) error
```

```go
// Apply middleware to a path prefix
evo.Use("/api", func(r *evo.Request) error {
    token := r.Header("Authorization")
    if token == "" {
        return r.Context.SendStatus(401)
    }
    return r.Next()
})

// Catch-all handler (set before evo.Run)
evo.Any = func(r *evo.Request) error {
    return r.Context.SendStatus(404)
}
```

---

## Route Groups

```go
api := evo.Group("/api/v1")

api.Get("/users", listUsers)
api.Post("/users", createUser)

// Group with middleware
admin := evo.Group("/admin", authMiddleware)
admin.Get("/dashboard", dashboard)

// Nested groups
v2 := api.Group("/v2")
v2.Get("/users", listUsersV2)

// Named groups (for URL generation)
api.Name("api.")
api.Get("/orders", listOrders) // route name: "api.GET/api/v1/orders"
```

---

## Route Parameters

```go
// Single named parameter
evo.Get("/users/:id", func(r *evo.Request) any {
    id := r.Param("id").Int()
    return id
})

// Multiple parameters
evo.Get("/users/:userId/orders/:orderId", func(r *evo.Request) any {
    userId  := r.Param("userId").Int64()
    orderId := r.Param("orderId").String()
    return nil
})

// Wildcard
evo.Get("/files/*", func(r *evo.Request) any {
    path := r.Param("*")
    return path
})

// All route params as a map
evo.Get("/a/:x/b/:y", func(r *evo.Request) any {
    params := r.Params() // map[string]string{"x":"...", "y":"..."}
    return params
})
```

---

## Query String

```go
// GET /search?q=fiber&page=2&active=true
evo.Get("/search", func(r *evo.Request) any {
    q      := r.Query("q").String()
    page   := r.Query("page").Int()
    active := r.Query("active").Bool()

    // All query params at once
    all := r.Queries() // map[string]string
    return all
})
```

---

## Request Headers

```go
evo.Get("/path", func(r *evo.Request) any {
    ct    := r.Header("Content-Type")
    token := r.Header("Authorization")

    // Quick existence check (Fiber v3)
    if r.HasHeader("X-Request-ID") { ... }

    // All request headers
    headers := r.ReqHeaders() // map[string]string

    return nil
})
```

---

## Request Body

### Auto-detect & bind (recommended — Fiber v3 Bind API)

```go
type CreateUserDTO struct {
    Name  string `json:"name" form:"name" query:"name"`
    Email string `json:"email" form:"email"`
    Age   int    `json:"age"  form:"age"`
}

evo.Post("/users", func(r *evo.Request) any {
    var dto CreateUserDTO
    if err := r.Bind().Body(&dto); err != nil {
        return err
    }
    return dto
})
```

`Bind()` automatically selects the decoder based on `Content-Type` (JSON, XML, form, multipart). You can also bind multiple sources in one chain:

```go
if err := r.Bind().Body(&body).Query(&q).Header(&h); err != nil { ... }
```

### BodyParser (legacy — auto-detects JSON / form / XML)

```go
evo.Post("/path", func(r *evo.Request) any {
    var body MyStruct
    if err := r.BodyParser(&body); err != nil {
        return err
    }
    return body
})
```

### Raw body

```go
evo.Post("/raw", func(r *evo.Request) any {
    raw := r.Body()    // string
    b   := r.BodyRaw() // []byte — undecoded (no decompression)
    _ = b
    return raw
})
```

### JSON body without struct (gjson)

```go
evo.Post("/path", func(r *evo.Request) any {
    name := r.ParseJsonBody().Get("user.name").String()
    age  := r.ParseJsonBody().Get("user.age").Int()
    return name + " " + strconv.Itoa(int(age))
})
```

### Single form value

```go
evo.Post("/path", func(r *evo.Request) any {
    name := r.FormValue("name").String()
    age  := r.FormValue("age").Int()
    return name
})
```

### Body existence check

```go
if !r.HasBody() {
    return r.Context.SendStatus(400)
}
```

### File upload

```go
evo.Post("/upload", func(r *evo.Request) any {
    file, err := r.FormFile("avatar")
    if err != nil {
        return err
    }
    return r.SaveFile(file, "./uploads/"+file.Filename)
})
```

---

## URL Information

```go
evo.Get("/info", func(r *evo.Request) any {
    full     := r.FullURL()      // https://example.com/info?x=1
    base     := r.BaseURL()      // https://example.com
    path     := r.Path()         // /info
    original := r.OriginalURL()  // /info?x=1
    host     := r.Hostname()     // example.com
    proto    := r.Protocol()     // https
    port     := r.Port()         // 443
    method   := r.Method()       // GET
    secure   := r.IsSecure()     // true
    local    := r.IsFromLocal()  // false
    xhr      := r.XHR()          // false
    referer  := r.Referer()      // "https://other.com"
    ips      := r.IPs()          // []string
    ip       := r.IP()           // "203.0.113.5"
    _ = full; _ = base; _ = path; _ = original; _ = host
    _ = proto; _ = port; _ = method; _ = secure; _ = local
    _ = xhr; _ = referer; _ = ips; _ = ip
    return nil
})
```

---

## Content-Type Helpers

```go
mt  := r.MediaType()  // "application/json"
cs  := r.Charset()    // "utf-8"
ok  := r.Is("json")   // true if Content-Type is application/json
```

---

## Sending Responses

### JSON (most common)

```go
evo.Get("/users/:id", func(r *evo.Request) any {
    user := getUser(r.Param("id").Int())
    return user  // auto-wrapped in {"success":true,"data":{...}}
})

// Or send directly without the wrapper:
r.JSON(user)
```

### XML

```go
r.XML(myStruct) // Content-Type: application/xml
```

### Plain text / HTML

```go
r.SendString("hello")
r.SendHTML("<h1>hello</h1>")
```

### Auto content-negotiation

```go
// Chooses format based on Accept header (JSON, XML, text, …)
r.AutoFormat(myData)
```

### Status codes

```go
r.SendStatus(204)
r.Status(201).JSON(created)
```

### File download

```go
r.SendFile("./report.pdf")
r.Download("./report.pdf", "monthly-report.pdf")
```

### Streaming

```go
evo.Get("/stream", func(r *evo.Request) any {
    pr, pw := io.Pipe()
    go func() {
        defer pw.Close()
        for i := 0; i < 5; i++ {
            fmt.Fprintf(pw, "chunk %d\n", i)
        }
    }()
    return r.SendStream(pr)
})
```

### Redirect

```go
r.Redirect("/new-path")          // 303 See Other (Fiber v3 default)
r.Redirect("/new-path", 301)     // Permanent redirect
r.Redirect("/new-path", 302)     // Temporary redirect

// Package-level redirects (registered at startup)
evo.Redirect("/old", "/new")                // 307
evo.RedirectPermanent("/old", "/new")       // 301
evo.RedirectTemporary("/old", "/new")       // 302
```

### Raw write

```go
r.Write([]byte("raw bytes"))
r.Write("string")
r.Write(42)
```

### Drop connection (DDoS mitigation)

```go
r.Drop() // closes TCP connection without sending any response
```

---

## Response Headers

```go
r.Set("X-Custom-Header", "value")
r.SetHeader("X-Custom-Header", "value") // alias
r.AppendHeader("Vary", "Accept-Encoding")
r.Type("json")             // sets Content-Type by extension
r.Vary("Accept-Language")
r.Links("http://api.example.com/users?page=2; rel=\"next\"")
r.Attachment("report.pdf") // Content-Disposition: attachment

// All response headers
headers := r.RespHeaders() // map[string]string
```

---

## Cookies

```go
evo.Get("/cookies", func(r *evo.Request) any {
    // Read
    val := r.Cookie("session")

    // Set simple value
    r.SetCookie("session", "abc123")

    // Set with expiry
    r.SetCookie("session", "abc123", 24*time.Hour)

    // Set complex value (JSON-encoded + base64)
    r.SetCookie("prefs", map[string]any{"theme": "dark"})

    // Full control
    r.SetRawCookie(&outcome.Cookie{
        Name:     "session",
        Value:    "abc123",
        Path:     "/",
        Domain:   "example.com",
        Expires:  time.Now().Add(24 * time.Hour),
        Secure:   true,
        HTTPOnly: true,
        SameSite: "Strict",
    })

    // Clear
    r.ClearCookie("session")

    _ = val
    return nil
})
```

---

## Locals (per-request storage)

```go
// Set in middleware
evo.Use("/", func(r *evo.Request) error {
    r.Context.Locals("userID", 42)
    return r.Next()
})

// Read in handler
evo.Get("/profile", func(r *evo.Request) any {
    id := r.Var("userID").Int() // generic.Value wrapper
    return id
})
```

---

## Static Files

```go
// Serve ./public at /
evo.Static("/", "./public")

// Serve ./assets at /static with options
evo.Static("/static", "./assets", static.Config{
    Browse:    false,  // disable directory listing
    Compress:  true,   // compress responses
    ByteRange: true,   // enable range requests
    MaxAge:    3600,   // Cache-Control max-age in seconds
})
```

---

## Named Routes & URL Generation

```go
evo.Get("/users/:id", listUser).Name("users.show")

// In a handler, generate a URL for a named route
url := r.Route("users.show", "id", 42) // "/users/42"
```

---

## Request Context Propagation

```go
// Attach values to a request for downstream handlers
evo.Use("/", func(r *evo.Request) error {
    r.Context.Locals("requestID", uuid.New().String())
    return r.Next()
})
```

---

## Structured Responses via `outcome` package

```go
import "github.com/getevo/evo/v2/lib/outcome"

evo.Get("/orders/:id", func(r *evo.Request) any {
    order, err := db.FindOrder(r.Param("id").Int())
    if err != nil {
        return outcome.NotFound("order not found")
    }
    return outcome.OK(order)
})
```

See [outcome.md](outcome.md) for the full list of status helpers.

---

## Graceful Shutdown

```go
// Default — waits indefinitely for connections to drain
evo.Shutdown()

// With timeout
evo.ShutdownWithTimeout(30 * time.Second)

// With context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
evo.ShutdownWithContext(ctx)
```

---

## Route Introspection

```go
// All routes (including middleware)
routes := evo.GetRoutes()

// Only non-middleware routes
routes := evo.GetRoutes(true)

for _, r := range routes {
    fmt.Println(r.Method, r.Path, r.Name)
}
```

---

## Access the Underlying Fiber App

```go
fiberApp := evo.GetFiber() // *fiber.App
```

---

## Configuration

HTTP server settings are controlled via `settings.yml` (or environment variables) under the `[HTTP]` section:

| Key | Default | Description |
|-----|---------|-------------|
| `Host` | `""` | Listen address |
| `Port` | `8080` | Listen port |
| `Prefork` | `false` | Enable prefork (multi-process) |
| `ServerHeader` | `""` | Server header value; also used as the trusted proxy header for IP detection |
| `StrictRouting` | `false` | Treat `/foo` and `/foo/` as different |
| `CaseSensitive` | `false` | Case-sensitive routing |
| `BodyLimit` | `4194304` | Max request body size (bytes) |
| `Concurrency` | `262144` | Max concurrent connections |
| `ReadTimeout` | `0` | Read deadline per connection |
| `WriteTimeout` | `0` | Write deadline per connection |
| `IdleTimeout` | `0` | Keep-alive idle timeout |
| `ReadBufferSize` | `4096` | Per-connection read buffer |
| `WriteBufferSize` | `4096` | Per-connection write buffer |
| `GETOnly` | `false` | Accept only GET requests |
| `DisableKeepalive` | `false` | Disable keep-alive |
| `ReduceMemoryUsage` | `false` | Reduce memory at the cost of CPU |
| `EnablePrintRoutes` | `false` | Print all routes on startup |
