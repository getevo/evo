# Using Webserver
EVO utilizes **[gofiber](https://github.com/gofiber/fiber)** as the web server framework, which, in turn, leverages **[fasthttp](https://github.com/valyala/fasthttp)** as a high-performance HTTP server. EVO enhances the fiber context by adding additional details and information to the original context. EVO extends the fiber context to enrich it with more detailed information. By wrapping the fiber context, EVO adds additional context-specific data and functionality, enhancing the capabilities of the original context.

### Basic Routing
```go
evo.Get("/api/*",func(request *evo.Request) interface{} {  
	return "hello world" 
})


evo.Get("/api/:name",func(request *evo.Request) interface{} {  
	return "Hello "+request.Param("name").String()
})

//Post request
evo.Post("/api/post",func(request *evo.Request) interface{} {
    return "Hello "+request.FormValue("name").String()
})


```

### Middleware & Next
```go
evo.Use("/api",func(request *evo.Request) interface{} {  
	return request.Next()
})
```

### Input/Output
#### Access url parameters
```go
// curl "/path/100/9999999999999999999999/hello-world"
evo.Get("/path/:integer/:int64/:text",func(request *evo.Request) interface{} {
    var i = request.Param("p1").Int()
    var j = request.Param("int64").Int64()
    var y = request.Param("text").String()
	return true
})

```
#### Access query string params
```go
// curl "/path?p1=true&int64=19999999999999999999&text=hello"
evo.Get("/path",func(request *evo.Request) interface{} {
    var i = request.Query("p1").Int()
    var j = request.Query("int64").Int64()
    var y = request.Query("text").String()
    return true
})

```
#### Access to headers
```go
// curl "/path" -H "x-header-key: value"
// note: headers are case-insensitive
evo.Get("/path",func(request *evo.Request) interface{} {
    var i = request.Header("x-header-key")
    return true
})
```
#### Read/Write cookies

```go
// Access to cookies
evo.Get("/path",func(request *evo.Request) interface{} {
	//Read cookie
    var i = request.Cookie("key")
	
	//Set key
    request.SetCookie("key","value")
	
	//Set object as value
	request.SetCookie("key",map[string]interface{}{"hello":"world"})
	
	//Set expiration duration
	request.SetCookie("key","value",24*time.Hour)
	
	// Set advanced cookie
    request.SetRawCookie(&outcome.Cookie{
        Name     :"key",
        Value    :"value",
        Path     :"/",
        Domain   :"",
        Expires: time.Now().Add(1*time.Hour),
        Secure  : true,
        HTTPOnly : true,
        SameSite : "",
    })
	
	
	// Clear cookie
	request.ClearCookie("key","key2","key3")
	
    return true
})

```


### Access to request Body
#### Parse Form/JSON body
The BodyParser function is designed to automatically detect the request type by assessing the Content-Type header of the request. It then unmarshals the request body into the provided struct based on the detected request type.
This method supports application/json, application/x-www-form-urlencoded and multipart/form-data.
```go
type MyStruct struct{
    Key1 string `json:"key1" form:"key1"`
    Key2 string `json:"key2" form:"key2"`
}
evo.Post("/path",func(request *evo.Request) interface{} {
    var body MyStruct
    err := request.BodyParser(&body)
	
    return true
})
```

##### Access JSON body attributes without struct
When you use the `ParseJsonBody` function, it returns an instance of `gjson.Result`, which provides a convenient way to query the JSON body. To understand how to query the JSON body using gjson, you can refer to the **[gjson documentation](https://github.com/tidwall/gjson)** for detailed information and examples.
```go

evo.Post("/path",func(request *evo.Request) interface{} {
    var body MyStruct
	var text = request.ParseJsonBody().Get("key1").String()
    var integer = request.ParseJsonBody().Get("key2").Int()
    return true
})
```


#### Access to single form value
```go
evo.Post("/path",func(request *evo.Request) interface{} {
    var i = request.FormValue("key").Int()
    return true
})
```

