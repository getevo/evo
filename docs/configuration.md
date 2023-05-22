# EVO Framework Configuration

The EVO Framework comes with a flexible configuration system that supports YAML files on the filesystem ("yml") as the default storage, and a database storage option called "database". Developers can also add custom storage interfaces based on their specific needs.

To run the EVO Framework, you need to configure the following settings in a local file called `config.yml`:

```yaml
#config.yml
Database:
   Cache: "false"
   ConnMaxLifTime: 1h
   Database: "database"
   Debug: "3"
   Enabled: "false"
   MaxIdleConns: "10"
   MaxOpenConns: "100"
   Params: ""
   Password: "password"
   SSLMode: "false"
   Server: 127.0.0.1:3306
   SlowQueryThreshold: 500ms
   Type: mysql
   Username: root
HTTP:
   BodyLimit: 1kb
   CaseSensitive: "false"
   CompressedFileSuffix: .evo.gz
   Concurrency: "1024"
   DisableDefaultContentType: "false"
   DisableDefaultDate: "false"
   DisableHeaderNormalizing: "false"
   DisableKeepalive: "false"
   ETag: "false"
   GETOnly: "false"
   Host: 0.0.0.0
   IdleTimeout: "0"
   Immutable: "false"
   Network: ""
   Port: "8080"
   Prefork: "false"
   ProxyHeader: X-Forwarded-For
   ReadBufferSize: 8kb
   ReadTimeout: 1s
   ReduceMemoryUsage: "false"
   ServerHeader: EVO
   StrictRouting: "false"
   UnescapePath: "false"
   EnablePrintRoutes: false
   WriteBufferSize: 4kb
   WriteTimeout: 5s
```

If you have configured the database settings, the framework will use the database configuration as the default storage driver. However, please note that the filesystem configuration has higher priority. You can have a default configuration driver along with optional custom drivers. The configuration system always checks the filesystem first, then the database, and subsequently other custom drivers to fetch the requested configuration key. The configuration key is constructed by combining the section name and the key name with a dot separator.

Here's an example of how you can retrieve a configuration:
```go 
// Retrieve the configuration using all available drivers.
var c = settings.Get("SECTION.KEY").Int()
```

To use a specific driver:
```go 
// Retrieve the configuration from the YAML file only.
var c = settings.Use("yml").Get("SECTION.KEY").Int()

// Retrieve the configuration from the database only.
var c = settings.Use("database").Get("SECTION.KEY").Int()
```

Please note that the **settings.Get** function returns a generic value. To cast it to other types or use it, please refer to the **[generic](generic.md)** section for further instructions.


To set the configuration:
```go
// following code set configuration on default driver
var  err = settings.Set("SECTION.KEY","MY VALUE")
```

To use a specific driver:
```go
// following code set configuration on default driver
var  err = settings.Use("database").Set("SECTION.KEY","MY VALUE")
```

## Configuration Drivers

- To get list of available drivers:
```go
var drivers = settings.Drivers()
```



- To get single driver:
```go
var driver = settings.Driver("name of driver")
```
