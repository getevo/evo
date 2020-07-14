---
layout: default
title: Request Context
nav_order: 5
---

EVO Request context is the upgraded version of [fiber.Context](https://github.com/gofiber/docs/blob/master/ctx.md) with extra features.

---
  The Request struct represents the Context which hold the HTTP request and
  response. It has methods for the request query string, parameters, body, HTTP
  headers and so on.
---

# Request

## Accepts

Checks, if the specified **extensions** or **content** **types** are acceptable.

Based on the request’s [Accept](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept) HTTP header.


```go
request.Accepts(types ...string)                 string
request.AcceptsCharsets(charsets ...string)      string
request.AcceptsEncodings(encodings ...string)    string
request.AcceptsLanguages(langs ...string)        string
```



```go
// Accept: text/*, application/json

app.Get("/", func(request *evo.Request) {
  request.Accepts("html")             // "html"
  request.Accepts("text/html")        // "text/html"
  request.Accepts("json", "text")     // "json"
  request.Accepts("application/json") // "application/json"
  request.Accepts("image/png")        // ""
  request.Accepts("png")              // ""
})
```


EVO provides similar functions for the other accept headers.

```go
// Accept-Charset: utf-8, iso-8859-1;q=0.2
// Accept-Encoding: gzip, compress;q=0.2
// Accept-Language: en;q=0.8, nl, ru

app.Get("/", func(request *evo.Request) {
  request.AcceptsCharsets("utf-16", "iso-8859-1") 
  // "iso-8859-1"

  request.AcceptsEncodings("compress", "br") 
  // "compress"

  request.AcceptsLanguages("pt", "nl", "ru") 
  // "nl"
})
```

## Append

Appends the specified **value** to the HTTP response header field.


If the header is **not** already set, it creates the header with the specified value.



```go
request.Append(field, values ...string)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Append("Link", "http://google.com", "http://localhost")
  // => Link: http://localhost, http://google.com

  request.Append("Link", "Test")
  // => Link: http://localhost, http://google.com, Test
})
```


## Attachment

Sets the HTTP response [Content-Disposition](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Disposition) header field to `attachment`.


```go
request.Attachment(file ...string)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Attachment()
  // => Content-Disposition: attachment

  request.Attachment("./upload/images/logo.png")
  // => Content-Disposition: attachment; filename="logo.png"
  // => Content-Type: image/png
})
```


## App

Returns the [\*App](app.md#new) reference so you could easily access all application settings.


```go
request.App() *App
```



```go
app.Get("/bodylimit", func(request *evo.Request) {
  bodylimit := request.App().Settings.BodyLimit
  request.Send(bodylimit)
})
```


## BaseURL

Returns base URL \(**protocol** + **host**\) as a `string`.


```go
request.BaseURL() string
```



```go
// GET https://example.com/page#chapter-1

app.Get("/", func(request *evo.Request) {
  request.BaseURL() // https://example.com
})
```


## Body

Contains the **raw body** submitted in a **POST** request.


```go
request.Body() string
```



```go
// curl -X POST http://localhost:8080 -d user=john

app.Post("/", func(request *evo.Request) {
  // Get raw body from POST request:
  request.Body() // user=john
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## BodyParser

Binds the request body to a struct. `BodyParser` supports decoding query parameters and the following content types based on the `Content-Type` header:

* `application/json`
* `application/xml`
* `application/x-www-form-urlencoded`
* `multipart/form-data`


```go
request.BodyParser(out interface{}) error
```



```go
// Field names should start with an uppercase letter
type Person struct {
    Name string `json:"name" xml:"name" form:"name" query:"name"`
    Pass string `json:"pass" xml:"pass" form:"pass" query:"pass"`
}

app.Post("/", func(request *evo.Request) {
        p := new(Person)

        if err := request.BodyParser(p); err != nil {
            log.Fatal(err)
        }

        log.Println(p.Name) // john
        log.Println(p.Pass) // doe
})
// Run tests with the following curl commands

// curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"pass\":\"doe\"}" localhost:3000

// curl -X POST -H "Content-Type: application/xml" --data "<login><name>john</name><pass>doe</pass></login>" localhost:3000

// curl -X POST -H "Content-Type: application/x-www-form-urlencoded" --data "name=john&pass=doe" localhost:3000

// curl -X POST -F name=john -F pass=doe http://localhost:3000

// curl -X POST "http://localhost:3000/?name=john&pass=doe"
```


## ClearCookie

Expire a client cookie \(_or all cookies if left empty\)_


```go
request.ClearCookie(key ...string)
```



```go
app.Get("/", func(request *evo.Request) {
  // Clears all cookies:
  request.ClearCookie()

  // Expire specific cookie by name:
  request.ClearCookie("user")

  // Expire multiple cookies by names:
  request.ClearCookie("token", "session", "track_id", "version")
})
```


## Context

Returns context.Context that carries a deadline, a cancellation signal, and other values across API boundaries.  
  
**Signature**

```go
request.Context() context.Context
```

## Cookie

Set cookie

**Signature**

```text
request.Cookie(*Cookie)
```

```go
type Cookie struct {
    Name     string
    Value    string
    Path     string
    Domain   string
    Expires  time.Time
    Secure   bool
    HTTPOnly bool
    SameSite string // lax, strict, none
}
```


```go
app.Get("/", func(request *evo.Request) {
  // Create cookie
  cookie := new(fiber.Cookie)
  cookie.Name = "john"
  cookie.Value = "doe"
  cookie.Expires = time.Now().Add(24 * time.Hour)

  // Set cookie
  request.Cookie(cookie)
})
```


## Cookies

Get cookie value by key.

**Signatures**

```go
request.Cookies(key string) string
```


```go
app.Get("/", func(request *evo.Request) {
  // Get cookie by key:
  request.Cookies("name") // "john"
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## Download

Transfers the file from path as an `attachment`.

Typically, browsers will prompt the user for download. By default, the [Content-Disposition](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Disposition) header `filename=` parameter is the file path \(_this typically appears in the browser dialog_\).

Override this default with the **filename** parameter.


```go
request.Download(path, filename ...string) error
```



```go
app.Get("/", func(request *evo.Request) {
  if err := request.Download("./files/report-12345.pdf"); err != nil {
    request.Next(err) // Pass err to EVO
  }
  // => Download report-12345.pdf

  if err := request.Download("./files/report-12345.pdf", "report.pdf"); err != nil {
    request.Next(err) // Pass err to EVO
  }
  // => Download report.pdf
})
```


## Fasthttp

You can still **access** and use all **Fasthttp** methods and properties.

**Signature**


Please read the [Fasthttp Documentation](https://pkg.go.dev/github.com/valyala/fasthttp?tab=doc) for more information.


**Example**

```go
app.Get("/", func(request *evo.Request) {
  request.Fasthttp.Request.Header.Method()
  // => []byte("GET")

  request.Fasthttp.Response.Write([]byte("Hello, World!"))
  // => "Hello, World!"
})
```

## Error

This contains the error information that thrown by a panic or passed via `Next` method.


```go
request.Error() error
```



```go
func main() {
  evo.Setup()
  evo.Post("/api/register", func (request *evo.Request) {
    if err := request.JSON(&User); err != nil {
      request.Next(err)
    }
  })
  evo.Get("/api/user", func (request *evo.Request) {
    if err := request.JSON(&User); err != nil {
      request.Next(err)
    }
  })
  evo.Put("/api/update", func (request *evo.Request) {
    if err := request.JSON(&User); err != nil {
      request.Next(err)
    }
  })
  evo.Use("/api", func(request *evo.Request) {
    request.Set("Content-Type", "application/json")
    request.Status(500).Send(request.Error())
  })
  evo.Run()
}
```


## Format

Performs content-negotiation on the [Accept](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept) HTTP header. It uses [Accepts](ctx.md#accepts) to select a proper format.


If the header is **not** specified or there is **no** proper format, **text/plain** is used.



```go
request.Format(body interface{})
```



```go
app.Get("/", func(request *evo.Request) {
  // Accept: text/plain
  request.Format("Hello, World!")
  // => Hello, World!

  // Accept: text/html
  request.Format("Hello, World!")
  // => <p>Hello, World!</p>

  // Accept: application/json
  request.Format("Hello, World!")
  // => "Hello, World!"
})
```


## FormFile

MultipartForm files can be retrieved by name, the **first** file from the given key is returned.


```go
request.FormFile(name string) (*multipart.FileHeader, error)
```



```go
app.Post("/", func(request *evo.Request) {
  // Get first file from form field "document":
  file, err := request.FormFile("document")

  // Check for errors:
  if err == nil {
    // Save file to root directory:
    request.SaveFile(file, fmt.Sprintf("./%s", file.Filename))
  }
})
```


## FormValue

Any form values can be retrieved by name, the **first** value from the given key is returned.


```go
request.FormValue(name string) string
```



```go
app.Post("/", func(request *evo.Request) {
  // Get first value from form field "name":
  request.FormValue("name")
  // => "john" or "" if not exist
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## Fresh

[https://expressjs.com/en/4x/api.html\#req.fresh](https://expressjs.com/en/4x/api.html#req.fresh)


Not implemented yet, pull requests are welcome!


## Get

Returns the HTTP request header specified by field.


The match is **case-insensitive**.



```go
request.Get(field string) string
```



```go
app.Get("/", func(request *evo.Request) {
  request.Get("Content-Type") // "text/plain"
  request.Get("CoNtEnT-TypE") // "text/plain"
  request.Get("something")    // ""
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## Hostname

Contains the hostname derived from the [Host](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Host) HTTP header.


```go
request.Hostname() string
```



```go
// GET http://google.com/search

app.Get("/", func(request *evo.Request) {
  request.Hostname() // "google.com"
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## IP

Returns the remote IP address of the request.


```go
request.IP() string
```



```go
app.Get("/", func(request *evo.Request) {
  request.IP() // "127.0.0.1"
})
```


## IPs

Returns an array of IP addresses specified in the [X-Forwarded-For](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For) request header.


```go
request.IPs() []string
```



```go
// X-Forwarded-For: proxy1, 127.0.0.1, proxy3

app.Get("/", func(request *evo.Request) {
  request.IPs() // ["proxy1", "127.0.0.1", "proxy3"]
})
```


## Is

Returns the matching **content type**, if the incoming request’s [Content-Type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type) HTTP header field matches the [MIME type](https://developer.mozilla.org/ru/docs/Web/HTTP/Basics_of_HTTP/MIME_types) specified by the type parameter.


If the request has **no** body, it returns **false**.



```go
request.Is(t string) bool
```



```go
// Content-Type: text/html; charset=utf-8

app.Get("/", func(request *evo.Request) {
  request.Is("html")  // true
  request.Is(".html") // true
  request.Is("json")  // false
})
```


## JSON

Converts any **interface** or **string** to JSON using [Jsoniter](https://github.com/json-iterator/go).


JSON also sets the content header to **application/json**.



```go
request.JSON(v interface{}) error
```



```go
type SomeStruct struct {
  Name string
  Age  uint8
}

app.Get("/json", func(request *evo.Request) {
  // Create data struct:
  data := SomeStruct{
    Name: "Grame",
    Age:  20,
  }

  if err := request.JSON(data); err != nil {
    request.Status(500).Send(err)
    return
  }
  // => Content-Type: application/json
  // => "{"Name": "Grame", "Age": 20}"

  if err := request.JSON(map[string]interface{}{
    "name": "Grame",
    "age": 20,
  }); err != nil {
    request.Status(500).Send(err)
    return
  }
  // => Content-Type: application/json
  // => "{"name": "Grame", "age": 20}"
})
```


## JSONP

Sends a JSON response with JSONP support. This method is identical to [JSON](ctx.md#json), except that it opts-in to JSONP callback support. By default, the callback name is simply callback.

Override this by passing a **named string** in the method.


```go
request.JSONP(v interface{}, callback ...string) error
```



```go
type SomeStruct struct {
  name string
  age  uint8
}

app.Get("/", func(request *evo.Request) {
  // Create data struct:
  data := SomeStruct{
    name: "Grame",
    age:  20,
  }

  request.JSONP(data)
  // => callback({"name": "Grame", "age": 20})

  request.JSONP(data, "customFunc")
  // => customFunc({"name": "Grame", "age": 20})
})
```


## Links

Joins the links followed by the property to populate the response’s [Link](https://developer.mozilla.org/ru/docs/Web/HTTP/Headers/Link) HTTP header field.


```go
request.Links(link ...string)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Link(
    "http://api.example.com/users?page=2", "next",
    "http://api.example.com/users?page=5", "last",
  )
  // Link: <http://api.example.com/users?page=2>; rel="next",
  //       <http://api.example.com/users?page=5>; rel="last"
})
```


## Locals

Method that stores string variables scoped to the request and therefore available only to the routes that match the request.


This is useful, if you want to pass some **specific** data to the next middleware.



```go
request.Locals(key string, value ...interface{}) interface{}
```



```go
app.Get("/", func(request *evo.Request) {
  request.Locals("user", "admin")
  request.Next()
})

app.Get("/admin", func(request *evo.Request) {
  if request.Locals("user") == "admin" {
    request.Status(200).Send("Welcome, admin!")
  } else {
    request.SendStatus(403)
    // => 403 Forbidden
  }
})
```


## Location

Sets the response [Location](https://developer.mozilla.org/ru/docs/Web/HTTP/Headers/Location) HTTP header to the specified path parameter.


```go
request.Location(path string)
```



```go
app.Post("/", func(request *evo.Request) {
  request.Location("http://example.com")
  request.Location("/foo/bar")
})
```


## Method

Contains a string corresponding to the HTTP method of the request: `GET`, `POST`, `PUT` and so on.  
Optionally, you could override the method by passing a string.


```go
request.Method(override ...string) string
```



```go
app.Post("/", func(request *evo.Request) {
  request.Method() // "POST"
})
```


## MultipartForm

To access multipart form entries, you can parse the binary with `MultipartForm()`. This returns a `map[string][]string`, so given a key the value will be a string slice.


```go
request.MultipartForm() (*multipart.Form, error)
```



```go
app.Post("/", func(request *evo.Request) {
  // Parse the multipart form:
  if form, err := request.MultipartForm(); err == nil {
    // => *multipart.Form

    if token := form.Value["token"]; len(token) > 0 {
      // Get key value:
      fmt.Println(token[0])
    }

    // Get all files from "documents" key:
    files := form.File["documents"]
    // => []*multipart.FileHeader

    // Loop through files:
    for _, file := range files {
      fmt.Println(file.Filename, file.Size, file.Header["Content-Type"][0])
      // => "tutorial.pdf" 360641 "application/pdf"

      // Save the files to disk:
      request.SaveFile(file, fmt.Sprintf("./%s", file.Filename))
    }
  }
})
```


## Next

When **Next** is called, it executes the next method in the stack that matches the current route. You can pass an error struct within the method for custom error handling.


```go
request.Next(err ...error)
```



```go
app.Get("/", func(request *evo.Request) {
  fmt.Println("1st route!")
  request.Next()
})

app.Get("*", func(request *evo.Request) {
  fmt.Println("2nd route!")
  request.Next(fmt.Errorf("Some error"))
})

app.Get("/", func(request *evo.Request) {
  fmt.Println(request.Error()) // => "Some error"
  fmt.Println("3rd route!")
  request.Send("Hello, World!")
})
```


## OriginalURL

Contains the original request URL.


```go
request.OriginalURL() string
```



```go
// GET http://example.com/search?q=something

app.Get("/", func(request *evo.Request) {
  request.OriginalURL() // "/search?q=something"
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## Params

Method can be used to get the route parameters.


Defaults to empty string \(`""`\), if the param **doesn't** exist.



```go
request.Params(param string) string
```



```go
// GET http://example.com/user/fenny

app.Get("/user/:name", func(request *evo.Request) {
  request.Params("name") // "fenny"
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)\_\_

## Path

Contains the path part of the request URL. Optionally, you could override the path by passing a string.


```go
request.Path(override ...string) string
```



```go
// GET http://example.com/users?sort=desc

app.Get("/users", func(request *evo.Request) {
  request.Path() // "/users"
})
```


## Protocol

Contains the request protocol string: `http` or `https` for **TLS** requests.


```go
request.Protocol() string
```



```go
// GET http://example.com

app.Get("/", func(request *evo.Request) {
  request.Protocol() // "http"
})
```


## Query

This property is an object containing a property for each query string parameter in the route.


If there is **no** query string, it returns an **empty string**.



```go
request.Query(parameter string) string
```



```go
// GET http://example.com/shoes?order=desc&brand=nike

app.Get("/", func(request *evo.Request) {
  request.Query("order") // "desc"
  request.Query("brand") // "nike"
})
```


> _Returned value is only valid within the handler. Do not store any references.  
> Make copies or use the_ [_**`Immutable`**_](app.md#settings) _setting instead._ [_Read more..._](./#zero-allocation)

## Range

An struct containg the type and a slice of ranges will be returned.


```go
request.Range(int size)
```



```go
// Range: bytes=500-700, 700-900
app.Get("/", func(request *evo.Request) {
  b := request.Range(1000)
  if b.Type == "bytes" {
      for r := range r.Ranges {
      fmt.Println(r)
      // [500, 700]
    }
  }
})
```


## Redirect

Redirects to the URL derived from the specified path, with specified status, a positive integer that corresponds to an HTTP status code.


If **not** specified, status defaults to **302 Found**.



```go
request.Redirect(path string, status ...int)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Redirect("/foo/bar")
  request.Redirect("../login")
  request.Redirect("http://example.com")
  request.Redirect("http://example.com", 301)
})
```



## Route

Contains the matched [Route](https://pkg.go.dev/github.com/gofiber/fiber?tab=doc#Route) struct.


```go
request.Route() *Route
```



```go
// http://localhost:8080/hello

request.Get("/hello", func(request *evo.Request) {
  r := request.Route()
  fmt.Println(r.Method, r.Path, r.Params, r.Regexp, r.Handler)
})

request.Post("/:api?", func(request *evo.Request) {
  request.Route()
  // => {GET /hello [] nil 0x7b49e0}
})
```


## SaveFile

Method is used to save **any** multipart file to disk.


```go
request.SaveFile(fh *multipart.FileHeader, path string)
```



```go
request.Post("/", func(request *evo.Request) {
  // Parse the multipart form:
  if form, err := request.MultipartForm(); err == nil {
    // => *multipart.Form

    // Get all files from "documents" key:
    files := form.File["documents"]
    // => []*multipart.FileHeader

    // Loop through files:
    for _, file := range files {
      fmt.Println(file.Filename, file.Size, file.Header["Content-Type"][0])
      // => "tutorial.pdf" 360641 "application/pdf"

      // Save the files to disk:
      request.SaveFile(file, fmt.Sprintf("./%s", file.Filename))
    }
  }
})
```


## Secure

A boolean property, that is `true` , if a **TLS** connection is established.


```go
request.Secure() bool
```



```go
// Secure() method is equivalent to:
request.Protocol() == "https"
```


## Send

Sets the HTTP response body. The **Send** body can be of any type.



```go
request.Send(body ...interface{})
```



```go
request.Get("/", func(request *evo.Request) {
  request.Send("Hello, World!")         // => "Hello, World!"
  request.Send([]byte("Hello, World!")) // => "Hello, World!"
  request.Send(123)                     // => 123
})
```


EVO also provides `SendBytes` ,`SendString` and `SendStream` methods for raw inputs.


Use this, if you **don't need** type assertion, recommended for **faster** performance.



```go
request.SendBytes(b []byte)
request.SendString(s string)
request.SendStream(r io.Reader, s ...int)
```



```go
app.Get("/", func(request *evo.Request) {
  request.SendByte([]byte("Hello, World!"))
  // => "Hello, World!"

  request.SendString("Hello, World!")
  // => "Hello, World!"
  
  request.SendStream(bytes.NewReader([]byte("Hello, World!")))
  // => "Hello, World!"
})
```


## SendFile

Transfers the file from the given path. Sets the [Content-Type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type) response HTTP header field based on the **filenames** extension.


Method use **gzipping** by default, set it to **true** to disable.



```go
request.SendFile(path string, compress ...bool) error
```



```go
app.Get("/not-found", func(request *evo.Request) {
  if err := request.SendFile("./public/404.html"); err != nil {
    request.Next(err) // pass err to ErrorHandler
  }

  // Enable compression
  if err := request.SendFile("./static/index.html", true); err != nil {
    request.Next(err) // pass err to ErrorHandler
  }
})
```


## SendStatus

Sets the status code and the correct status message in the body, if the response body is **empty**.


```go
request.SendStatus(status int)
```



```go
app.Get("/not-found", func(request *evo.Request) {
  request.SendStatus(415)
  // => 415 "Unsupported Media Type"

  request.Send("Hello, World!")
  request.SendStatus(415)
  // => 415 "Hello, World!"
})
```


## Set

Sets the response’s HTTP header field to the specified `key`, `value`.


```go
request.Set(field, value string)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Set("Content-Type", "text/plain")
  // => "Content-type: text/plain"
})
```


## Stale

[https://expressjs.com/en/4x/api.html\#req.fresh](https://expressjs.com/en/4x/api.html#req.fresh)


Not implemented yet, pull requests are welcome!


## Status

Sets the HTTP status for the response.


Method is a **chainable**.



```go
request.Status(status int)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Status(200)
  request.Status(400).Send("Bad Request")
  request.Status(404).SendFile("./public/gopher.png")
})
```


## Subdomains

An array of subdomains in the domain name of the request.

The application property subdomain offset, which defaults to `2`, is used for determining the beginning of the subdomain segments.


```go
request.Subdomains(offset ...int) []string
```



```go
// Host: "tobi.ferrets.example.com"

app.Get("/", func(request *evo.Request) {
  request.Subdomains()  // ["ferrets", "tobi"]
  request.Subdomains(1) // ["tobi"]
})
```


## Type

Sets the [Content-Type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type) HTTP header to the MIME type listed [here](https://github.com/nginx/nginx/blob/master/conf/mime.types) specified by the file **extension**.


```go
request.Type(t string) string
```



```go
app.Get("/", func(request *evo.Request) {
  request.Type(".html") // => "text/html"
  request.Type("html")  // => "text/html"
  request.Type("json")  // => "application/json"
  request.Type("png")   // => "image/png"
})
```


## Vary

Adds the given header field to the [Vary](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Vary) response header. This will append the header, if not already listed, otherwise leaves it listed in the current location.


Multiple fields are **allowed**.



```go
request.Vary(field ...string)
```



```go
app.Get("/", func(request *evo.Request) {
  request.Vary("Origin")     // => Vary: Origin
  request.Vary("User-Agent") // => Vary: Origin, User-Agent

  // No duplicates
  request.Vary("Origin") // => Vary: Origin, User-Agent

  request.Vary("Accept-Encoding", "Accept")
  // => Vary: Origin, User-Agent, Accept-Encoding, Accept
})
```


## Write

Appends **any** input to the HTTP body response.


```go
request.Write(body ...interface{})
```



```go
app.Get("/", func(request *evo.Request) {
  request.Write("Hello, ")         // => "Hello, "
  request.Write([]byte("World! ")) // => "Hello, World! "
  request.Write(123)               // => "Hello, World! 123"
})
```


## XHR

A Boolean property, that is `true`, if the request’s [X-Requested-With](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers) header field is [XMLHttpRequest](https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest), indicating that the request was issued by a client library \(such as [jQuery](https://api.jquery.com/jQuery.ajax/)\).


```go
request.XHR() bool
```



```go
// X-Requested-With: XMLHttpRequest

app.Get("/", func(request *evo.Request) {
  request.XHR() // true
})
```

