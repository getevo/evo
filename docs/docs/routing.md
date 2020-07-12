---
layout: default
title: Routing
nav_order: 5
---

EVO inherit the ultimate power of [Fiber](https://github.com/gofiber/fiber) in background. So all the routing system is inherited from Fiber.

**Routing** refers to how an application's endpoints (URIs) respond to client requests.

## Paths

Route paths, in combination with a request method, define the endpoints at which requests can be made. Route paths can be **strings** or **string patterns**.

**Examples of route paths based on strings**

```go
// This route path will match requests to the root route, "/":
evo.Get("/", func(request *evo.Request) {
  request.Send("root")
})

// This route path will match requests to "/about":
evo.Get("/about", func(request *evo.Request) {
  request.Send("about")
})

// This route path will match requests to "/random.txt":
evo.Get("/random.txt", func(request *evo.Request) {
  request.Send("random.txt")
})
```

## Parameters

Route parameters are **named URL segments** that are used to capture the values specified at their position in the URL. The captured values can be retrieved using the [Params](https://fiber.wiki/context#params) function, with the name of the route parameter specified in the path as their respective keys.


Name of the route parameter must be made up of **characters** \(`[A-Za-z0-9_]`\).


**Example of define routes with route parameters**

```go
// Parameters
evo.Get("/user/:name/books/:title", func(request *evo.Request) {
  request.Write(request.Params("name"))
  request.Write(request.Params("title"))
})
// Wildcard
evo.Get("/user/*", func(request *evo.Request) {
  request.Send(request.Params("*"))
})
// Optional parameter
app.Get("/user/:name?", func(request *evo.Request) {
  request.Send(request.Params("name"))
})
```


 Since the hyphen \(`-`\) and the dot \(`.`\) are interpreted literally, they can be used along with route parameters for useful purposes.


```go
// http://localhost:3000/plantae/prunus.persica
evo.Get("/plantae/:genus.:species", func(request *evo.Request) {
  request.Params("genus")   // prunus
  request.Params("species") // persica
})
```

```go
// http://localhost:3000/flights/LAX-SFO
evo.Get("/flights/:from-:to", func(request *evo.Request) {
  request.Params("from")   // LAX
  request.Params("to")     // SFO
})
```

## Middleware

Functions, that are designed to make changes to the request or response, are called **middleware functions**. The [Next](https://github.com/gofiber/docs/tree/34729974f7d6c1d8363076e7e88cd71edc34a2ac/context/README.md#next) is a **Fiber** router function, when called, executes the **next** function that **matches** the current route.

**Example of a middleware function**

```go
app.Use(func(request *evo.Request) {
  // Set some security headers:
  request.Set("X-XSS-Protection", "1; mode=block")
  request.Set("X-Content-Type-Options", "nosniff")
  request.Set("X-Download-Options", "noopen")
  request.Set("Strict-Transport-Security", "max-age=5184000")
  request.Set("X-Frame-Options", "SAMEORIGIN")
  request.Set("X-DNS-Prefetch-Control", "off")

  // Go to next middleware:
  request.Next()
})

app.Get("/", func(request *evo.Request) {
  request.Send("Hello, World!")
})
```

`Use` method path is a **mount** or **prefix** path and limits middleware to only apply to any paths requested that begin with it. This means you cannot use `:params` on the `Use` method.

## Grouping

If you have many endpoints, you can organize your routes using `Group`

```go
func main() {
  evo.Setup()
  
  api := evo.Group("/api", cors())  // /api

  v1 := api.Group("/v1", mysql())   // /api/v1
  v1.Get("/list", handler)          // /api/v1/list
  v1.Get("/user", handler)          // /api/v1/user

  v2 := api.Group("/v2", mongodb()) // /api/v2
  v2.Get("/list", handler)          // /api/v2/list
  v2.Get("/user", handler)          // /api/v2/user
  
  evo.Start()
}
```
