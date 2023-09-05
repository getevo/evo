# CURL
An uncomplicated HTTP request-handling library designed with the purpose of providing a user-friendly wrapper for the native GoLang HTTP client.
## Usage

To use this package, import it in your Go code:
```go
import "github.com/getevo/evo/v2/lib/curl"
```


### Examples:


- [Simple Request ](#simple-request)
- [Custom Method Request ](#custom-method-request)
- [Get Response Body ](#get-response-body)
- [Simple Post Request with Response Parser ](#simple-post-request-with-response-parser)
- [Debug Single Request ](#debug-single-request)
- [Debug All Requests ](#debug-all-requests)
- [Set Object as JSON Body ](#set-object-as-json-body)
- [Set Object as XML Body ](#set-object-as-xml-body)
- [Set RAW Body ](#set-raw-body)
- [Set Form Value ](#set-form-value)
- [Set Query String Value ](#set-query-string-value)
- [Set Cache on Results ](#set-cache-on-results)
- [Set Header ](#set-header)
- [Force Headers ](#force-headers)
- [Set Header from Struct ](#set-header-from-struct)
- [Request with Context ](#request-with-context)
- [Request with Custom HTTP Client ](#request-with-custom-http-client)
- [Timeout ](#timeout)
- [BasicAuth ](#basicauth)
- [Download Progress ](#download-progress)
- [Upload File ](#upload-file)
- [Request Cost ](#request-cost)
- [Access to Underlay http.Request and http.Response ](#access-to-underlay-httprequest-and-httpresponse)


#### Simple Request
This Go code performs an HTTP GET request retrieve the client's public IP address in JSON format. If there's an error during the request, it panics. Otherwise, it prints the server's response (the IP address in JSON) to the console.

**Supported Methods**
- **GET**
- **POST**
- **PUT**
- **PATCH**
- **DELETE**
- **HEAD**
- **OPTIONS**
```go
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.String())
```


#### Custom Method Request
This Go code performs an HTTP with custom method request to retrieve the client's public IP address in JSON format. If there's an error during the request, it panics. Otherwise, it prints the server's response (the IP address in JSON) to the console.
```go
    var resp, err = curl.Do("GET","https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.String())
```

### Get Response Body
You may get response body using one of ToString, String, ToBytes, Bytes functions.
```go
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.ToString()) // return response as string
    fmt.Println(resp.String()) // return response as string
    fmt.Println(resp.ToBytes()) // return response as []bytes
    fmt.Println(resp.Bytes()) // return response as []bytes
```

#### Simple Post Request with response parser
This code retrieves the public IP address of the client from the specified API, parses it into a Go struct.
##### Parsers:
- **ToJSON**
- **ToXML**
```go
    var result struct{
	    IP string `json:"ip"`
    }
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    err := resp.ToJSON(&result)
    if err != nil {
        panic(err)
    }
	evo.Dump(result)
```


#### Debug Single Request
This Go code sends an HTTP GET request to retrieve the client's public IP address in JSON format. If there's an error during the request, it panics. Otherwise, it prints the server's response, including headers and content, to the console.
```go
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Debug All Requests
In this Go code, curl.Debug is set to true, which enables debug mode for all subsequent curl library calls. This means that additional debug information, such as request and response details, will be displayed while making HTTP requests.
```go
    curl.Debug = true
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
```


#### Set object as json body
This Go code takes a predefined struct, converts it into a JSON format, and then uses it as the request body in an HTTP POST call to a specified URL.
```go
    var data = struct{
	    Param1 string
		Param2 int
    }{
	    Param1:"Hello World",
		Param2: 100
    }
    var resp, err = curl.Post("https://postman-echo.com/post",curl.BodyJSON(data))
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Set object as XML body
This Go code takes a predefined struct, converts it into XML format, and then uses it as the request body in an HTTP POST call to a specified URL.
```go
    var data = struct{
	    Param1 string
		Param2 int
    }{
	    Param1:"Hello World",
		Param2: 100
    }
    var resp, err = curl.Post("https://postman-echo.com/post",curl.BodyXML(data))
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Set RAW body
This Go code sends an HTTP POST request with a raw request body containing the text "hello world!". curl.BodyRaw accepts string, []byte, io.Reader, io.ReadCloser.
```go
    //curl.BodyRaw accepts string, []byte, io.Reader, io.ReadCloser
    var resp, err = curl.Post("https://postman-echo.com/post",curl.BodyRaw("hello world!"))
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Set Form Value
This Go code makes an HTTP POST request and includes two parameters in the request body: "param1" with the value "hello world" and "param2" with the value "100".
```go
    var resp, err = curl.Post("https://postman-echo.com/post",curl.Param{
        "param1":"hello world",
        "param2":"100"
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Set Query String Value
This Go code sends an HTTP GET request to "https://postman-echo.com/get" with query parameters "param1" and "param2" set to "hello world" and "100," respectively.
```go
    var resp, err = curl.Get("https://postman-echo.com/get",curl.QueryParam{
        "param1":"hello world",
        "param2":"100"
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Set cache on results
Enabling caching in the Curl library allows it to store previous responses for specific URLs and methods for a limited duration. When the same request is made again, the library will retrieve and return the cached response.
```go
    var resp, err = curl.Post("https://postman-echo.com/post",curl.BodyRaw(time.Now().String()),curl.Cache{Duration: 10 * time.Second})
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.String())
```


#### Set Header
This Go code sends an HTTP POST request with a custom user-agent header "CURL/GO".

#### Accepts:
-  **curl.Header**
- **http.Header**
```go
    // alternatively http.Header is allowed
    var resp, err = curl.Post("https://postman-echo.com/post",curl.Header{
	    "X-User-Agent":"CURL/GO"
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Force Headers
This Go code force headers on request.
```go
    var resp, err = curl.Post("https://postman-echo.com/post",curl.ReservedHeader{
	    "User-Agent":"CURL/GO",
		"Content-Type":"text/html"
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Set Header from Struct
This Go code snippet demonstrates how to use the curl library to make an HTTP POST request to the "https://postman-echo.com/post" URL with a custom user-agent header.
```go
    var data = struct{
        Param1 string
     Param2 int
    }{
        Param1:"Hello World",
        Param2: 100
    }
    var resp, err = curl.Post("https://postman-echo.com/post",curl.HeaderFromStruct(data))
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Request with context
This Go code demonstrates how to use a context with curl library package to manage the execution of the HTTP request.
```go
    // Create a context with a timeout of 5 seconds
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel() // Ensure the context is canceled when done
    var resp, err = curl.Post("https://postman-echo.com/post",ctx)
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Request with custom http client
This Go code creates a custom HTTP client with specific settings, including a 10-second timeout for HTTP requests, maintaining up to 10 idle connections, a 30-second idle connection timeout, enabling HTTP keep-alives, and a 5-second TLS handshake timeout.
```go
    client := &http.Client{
        Timeout: 10 * time.Second, // Set a timeout for HTTP requests
        Transport: &http.Transport{
            MaxIdleConns:        10, // Maximum idle connections to keep alive
            IdleConnTimeout:     30 * time.Second, // Idle connection timeout
            DisableKeepAlives:   false, // Enable HTTP keep-alives
            TLSHandshakeTimeout: 5 * time.Second, // TLS handshake timeout
        },
    }
    var resp, err = curl.Post("https://postman-echo.com/post",client)
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```

#### Timeout
This Go code sends an HTTP GET request with 1 second timeout.
```go
    var resp, err = curl.Get("https://speed.hetzner.de/10GB.bin",1*time.Duration)
    if err != nil {
    panic(err)
    }
    fmt.Println(resp.Dump())
```

#### BasicAuth
This Go code sends an HTTP POST request with basic authentication credentials.
```go
    var resp, err = curl.Get("http://httpbin.org/basic-auth/user/passwd",curl.BasicAuth{Username:"user",Password:"passwd"})
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Download Progress
This Go code defines a custom download progress function that prints the current and total bytes downloaded and the download progress percentage. It then uses the curl library to download a 10GB file from "https://speed.hetzner.de/10GB.bin," displaying the progress using the custom function.

```go
    var progress curl.DownloadProgress = func(current, total int64) {
        fmt.Println(current, "of", total, (float64(current)/float64(total))*100, "%")
    }

    var resp, err = curl.Get("https://speed.hetzner.de/1GB.bin", progress)
    if err != nil {
        panic(err)
    }
    err = resp.ToFile("./1GB.bin")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Upload File
This Go code defines a custom upload progress function that calculates and displays the current and total bytes uploaded and the upload progress percentage.
```go
    var progress curl.UploadProgress = func(current, total int64) {
        fmt.Println(current, "of", total, (float64(current)/float64(total))*100, "%")
    }
    file, err := os.Open("./test.jpg")
    if err != nil {
        panic(err)
    }
    defer file.Close()
    resp, err := curl.Post("https://postman-echo.com/post", progress, []curl.FileUpload{
        curl.FileUpload{FieldName: "image", FileName: "./test.jpg", File: file},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Dump())
```


#### Request Cost
Calculates the time spent to run an HTTP request and receive a valid response
```go
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Cost().Milliseconds() ,"ms")
```


#### Access to underlay http.Request and http.Response
These actions can be useful when you need to access lower-level HTTP details in your code, beyond what the curl library provides in its high-level response object.

```go
    var resp, err = curl.Get("https://api.ipify.org?format=json")
    if err != nil {
        panic(err)
    }
    var request *http.Request = resp.Request()
    var response *http.Response = resp.Response()
```

#### [< Table of Contents](https://github.com/getevo/evo#table-of-contents)