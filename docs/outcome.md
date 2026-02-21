# Outcome — HTTP Response Library

`lib/outcome` provides a fluent, type-safe API for building HTTP responses. It is the standard way to return data from EVO handlers.

## Import

```go
import "github.com/getevo/evo/v2/lib/outcome"
```

## Basic usage

Return an outcome value from any handler function:

```go
evo.Get("/users/:id", func(r *evo.Request) any {
    user, err := getUserByID(r.Params("id"))
    if err != nil {
        return outcome.NotFound("user not found")
    }
    return outcome.OK(user)
})
```

## Response constructors

### 2xx — Success

| Function | Status | Default body |
|---|---|---|
| `OK(data...)` | 200 | `"OK"` |
| `Created(data...)` | 201 | `"Created"` |
| `Accepted(data...)` | 202 | `"Accepted"` |
| `NoContent(data...)` | 204 | _(empty)_ |

```go
// 200 with JSON body
return outcome.OK(map[string]any{"id": 1, "name": "Alice"})

// 200 plain text
return outcome.OK("operation complete")

// 201 with created resource
return outcome.Created(newUser)

// 204 no body
return outcome.NoContent()
```

### 4xx — Client errors

| Function | Status |
|---|---|
| `BadRequest(data...)` | 400 |
| `Unauthorized(data...)` | 401 |
| `Forbidden(data...)` | 403 |
| `NotFound(data...)` | 404 |
| `NotAcceptable(data...)` | 406 |
| `RequestTimeout(data...)` | 408 |
| `Conflict(data...)` | 409 |
| `UnprocessableEntity(data...)` | 422 |
| `TooManyRequests(data...)` | 429 |
| `UnavailableForLegalReasons(data...)` | 451 |

```go
return outcome.BadRequest("invalid input")
return outcome.Unauthorized("token expired")
return outcome.Forbidden("insufficient permissions")
return outcome.NotFound(map[string]string{"error": "user not found"})
return outcome.Conflict("duplicate email")
return outcome.UnprocessableEntity(validationErrors)
return outcome.TooManyRequests("slow down")
```

### 5xx — Server errors

| Function | Status |
|---|---|
| `InternalServerError(data...)` | 500 |
| `ServiceUnavailable(data...)` | 503 |
| `GatewayTimeout(data...)` | 504 |

```go
return outcome.InternalServerError("unexpected error")
return outcome.ServiceUnavailable("maintenance mode")
return outcome.GatewayTimeout("upstream timed out")
```

### Content type helpers

```go
return outcome.Text("plain text response")     // text/plain
return outcome.Html("<h1>Hello</h1>")          // text/html
return outcome.Json(myStruct)                  // application/json
```

### Redirects

```go
return outcome.Redirect("/new-path")                          // 307 Temporary
return outcome.RedirectTemporary("/new-path")                 // 307
return outcome.RedirectPermanent("/canonical")                // 301
return outcome.Redirect("/login", fiber.StatusFound)          // custom code
```

## Builder methods

All constructors return `*Response`. Chain methods to customize:

### `.Status(code int)`

Override the status code:

```go
return outcome.OK(data).Status(206) // 206 Partial Content
```

### `.Header(key, value string)`

Add or replace a response header:

```go
return outcome.OK(data).
    Header("X-Request-ID", requestID).
    Header("X-API-Version", "v2")
```

### `.Content(input any)`

Replace the response body. Strings and `[]byte` are sent as-is; structs/maps/slices are JSON-encoded.

```go
return outcome.OK().Content(myStruct)
```

### `.Cookie(key, value, params...)`

Add a cookie. Complex values (maps, structs, slices) are JSON+base64 encoded.

```go
// Simple string
return outcome.OK(data).Cookie("session", sessionToken)

// With expiry
return outcome.OK(data).Cookie("session", token, 24*time.Hour)

// With expiry time
return outcome.OK(data).Cookie("session", token, time.Now().Add(24*time.Hour))

// Complex value (auto JSON+base64 encoded)
return outcome.OK(data).Cookie("prefs", userPreferences, 30*24*time.Hour)
```

### `.RawCookie(cookie Cookie)`

Add a fully configured cookie:

```go
return outcome.OK(data).RawCookie(outcome.Cookie{
    Name:     "session",
    Value:    token,
    Path:     "/",
    Secure:   true,
    HTTPOnly: true,
    SameSite: "Strict",
    Expires:  time.Now().Add(24 * time.Hour),
})
```

### `.Redirect(to, code...)`

Add a redirect (can be chained after setting other properties):

```go
return outcome.OK().Redirect("/dashboard")
return outcome.OK().Redirect("/dashboard", 302)
```

### `.Error(value, code...)`

Add an error message and set status (default 400):

```go
return outcome.OK().Error("validation failed", 422)
```

### `.SetCacheControl(duration, directives...)`

Set `Cache-Control` header:

```go
// Cache for 1 hour
return outcome.OK(data).SetCacheControl(time.Hour)

// Public cache with revalidation
return outcome.OK(data).SetCacheControl(5*time.Minute, "public", "must-revalidate")

// No caching
return outcome.OK(data).SetCacheControl(0, "no-store")
```

### `.Filename(name string)`

Set `Content-Disposition: attachment` for file downloads:

```go
return outcome.OK(fileBytes).
    Header("Content-Type", "application/pdf").
    Filename("report.pdf")
```

### `.ShowInBrowser()`

Set `Content-Disposition: inline` to display in the browser:

```go
return outcome.OK(imageBytes).
    Header("Content-Type", "image/png").
    ShowInBrowser()
```

## Handler examples

### JSON API endpoint

```go
evo.Get("/api/users", func(r *evo.Request) any {
    var users []User
    if err := db.Find(&users).Error; err != nil {
        return outcome.InternalServerError("failed to load users")
    }
    return outcome.OK(users)
})
```

### Create resource

```go
evo.Post("/api/users", func(r *evo.Request) any {
    var input CreateUserInput
    if err := r.BodyParser(&input); err != nil {
        return outcome.BadRequest("invalid request body")
    }
    if errs := validation.Struct(input); len(errs) > 0 {
        return outcome.UnprocessableEntity(errs)
    }
    user, err := createUser(input)
    if err != nil {
        return outcome.InternalServerError(err.Error())
    }
    return outcome.Created(user).
        Header("Location", "/api/users/"+strconv.Itoa(user.ID))
})
```

### Authentication

```go
evo.Post("/auth/login", func(r *evo.Request) any {
    token, err := authenticate(r.FormValue("email"), r.FormValue("password"))
    if err != nil {
        return outcome.Unauthorized("invalid credentials")
    }
    return outcome.OK(map[string]string{"token": token}).
        Cookie("session", token, 24*time.Hour).
        Header("X-Auth-Token", token)
})
```

### File download

```go
evo.Get("/reports/:id/pdf", func(r *evo.Request) any {
    data, err := generatePDF(r.Params("id"))
    if err != nil {
        return outcome.NotFound("report not found")
    }
    return outcome.OK(data).
        Header("Content-Type", "application/pdf").
        Filename("report-"+r.Params("id")+".pdf").
        SetCacheControl(10 * time.Minute)
})
```

### Paginated list

```go
evo.Get("/api/articles", func(r *evo.Request) any {
    page := r.QueryInt("page", 1)
    limit := r.QueryInt("limit", 20)

    var articles []Article
    var total int64
    db.Model(&Article{}).Count(&total)
    db.Offset((page - 1) * limit).Limit(limit).Find(&articles)

    return outcome.OK(map[string]any{
        "data":  articles,
        "total": total,
        "page":  page,
        "limit": limit,
    })
})
```

### Redirect with cookie cleanup

```go
evo.Post("/auth/logout", func(r *evo.Request) any {
    return outcome.RedirectTemporary("/login").
        Cookie("session", "", time.Now().Add(-time.Hour)) // expire cookie
})
```

## `HTTPSerializer` interface

Implement `HTTPSerializer` to make your own type returnable from handlers:

```go
type HTTPSerializer interface {
    GetResponse() Response
}

// Example: custom API response wrapper
type APIResponse struct {
    Success bool   `json:"success"`
    Data    any    `json:"data,omitempty"`
    Message string `json:"message,omitempty"`
}

func (a APIResponse) GetResponse() outcome.Response {
    code := 200
    if !a.Success {
        code = 400
    }
    data, _ := json.Marshal(a)
    return outcome.Response{
        StatusCode:  code,
        ContentType: "application/json",
        Data:        data,
    }
}

// Use in handler
evo.Get("/api/data", func(r *evo.Request) any {
    return APIResponse{Success: true, Data: result}
})
```

## `Cookie` type

```go
type Cookie struct {
    Name     string    `json:"name"`
    Value    string    `json:"value"`
    Path     string    `json:"path"`
    Domain   string    `json:"domain"`
    Expires  time.Time `json:"expires"`
    Secure   bool      `json:"secure"`
    HTTPOnly bool      `json:"http_only"`
    SameSite string    `json:"same_site"`  // Strict | Lax | None
}
```

## `Response` struct

```go
type Response struct {
    ContentType string
    Data        interface{}       // []byte, string, or any (auto JSON-marshaled)
    StatusCode  int
    Headers     map[string]string
    RedirectURL string
    Cookies     []*Cookie
    Errors      []string
}
```

Access raw data:

```go
resp := outcome.OK(myData)
bytes := resp.GetData() // []byte
```

## See Also

- [Web Server](webserver.md)
- [Validation](validation.md)
- [Errors](../lib/errors/README.md)
