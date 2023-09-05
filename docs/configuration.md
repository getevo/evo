# EVO Framework Configuration

The EVO Framework comes with a flexible configuration system that supports YAML files on the filesystem ("yml")as the default storage, and a database storage option called "database". Developers can also add custom storage interfaces based on their specific needs.

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

To retrieve all configuration keys as **`map[string]generic.Value`**:
```go
var m = settings.All()
```

To use a specific driver:
```go 
// Retrieve the configuration from the YAML file only.
var c = settings.Use("yml").Get("SECTION.KEY").Int()

// Retrieve the configuration from the database only.
var c = settings.Use("database").Get("SECTION.KEY").Int()

```

Please note that the **settings.Get** function returns a generic value. To cast it to other types or use it, please refer to the **[generic](generic.md)** section for further instructions.


To set single configuration key:
```go
// following code set configuration on default driver
var  err = settings.Set("SECTION.KEY","MY VALUE")
```

To set multiple keys at once:
```go
// following code set configuration on default driver
var  err = settings.SetMulti(data map[string]interface{}{
	"SECTION.KEY1":"value goes here",
	"SECTION.KEY2":"value goes here",
})
```

To use a specific driver:
```go
// following code set configuration on default driver
var  err = settings.Use("database").Set("SECTION.KEY","MY VALUE")
```

### Register configuration
To register configuration keys and use them later, you need to perform the registration step. Registering configuration keys enhances code readability and provides documentation for each configuration key. Here are examples of how to introduce configuration keys to the system:

- Registering a set of configuration keys using a config struct:
```go
type DatabaseConfig struct {
	Enabled            bool          `description:"Enabled database" default:"false" json:"enabled" yaml:"enabled"`
	Type               string        `description:"Database engine" default:"sqlite" json:"type" yaml:"type"`
	Username           string        `description:"Username" default:"root" json:"username" yaml:"username"`
	Password           string        `description:"Password" default:"" json:"password" yaml:"password"`
	Server             string        `description:"Server" default:"127.0.0.1:3306" json:"server" yaml:"server"`
	Cache              string        `description:"Enabled query cache" default:"false" json:"cache" yaml:"cache"`
	Debug              int           `description:"Debug level (1-4)" default:"3" params:"{\"min\":1,\"max\":4}" json:"debug" yaml:"debug"`
	Database           string        `description:"Database Name" default:"" json:"database" yaml:"database"`
	SSLMode            string        `description:"SSL Mode (required by some DBMS)" default:"false" json:"ssl-mode" yaml:"ssl-mode"`
	Params             string        `description:"Extra connection string parameters" default:"" json:"params" yaml:"params"`
	MaxOpenConns       int           `description:"Max pool connections" default:"100" json:"max-open-connections" yaml:"max-open-connections"`
	MaxIdleConns       int           `description:"Max idle connections in pool" default:"10" json:"max-idle-connections" yaml:"max-idle-connections"`
	ConnMaxLifTime     time.Duration `description:"Max connection lifetime" default:"1h" json:"connection-max-lifetime" yaml:"connection-max-lifetime"`
	SlowQueryThreshold time.Duration `description:"Slow query threshold" default:"500ms" json:"slow_query_threshold" yaml:"slow-query-threshold"`
}

// This code will register all keys of the DatabaseConfig struct under the Database section
// Note that this function will take description and default tags to fill the Value and Description attributes of correspondence settings.Setting struct
settings.Register("Database", &config)
```

- Registering a custom key:
```go
settings.Register(
	settings.Setting{
		Domain:      "CACHE",
		Name:        "REDIS_ADDRESS",
		Title:       "Redis server(s) address",
		Description: "Redis servers address. Separate using a comma if cluster.",
		Type:        "text",
		ReadOnly:    false,
		Visible:     true,
	},
	settings.Setting{
		Domain:      "CACHE",
		Name:        "REDIS_PREFIX",
		Title:       "Redis key prefix",
		Description: "Set a prefix for keys to prevent conjunction of keys in case of multiple applications running on the same instance of Redis",
		Type:        "text",
		ReadOnly:    false,
		Visible:     true,
	},
)
```
By registering configuration keys, you make them available for use in your application and provide additional information about their purpose, default values, and other details.


## Configuration Drivers

- To get list of available drivers:
```go
var drivers = settings.Drivers()
```

- To switch to specific driver:
```go
var drivers = settings.Use("name of driver")
```


- To get single driver:
```go
var driver = settings.Driver("name of driver")
```

- To set default driver:
```go
settings.SetDefaultDriver("name of driver")
```

## Create and Add driver
To create a new driver, you need to implement the **`settings.Interface`** with the following methods:

```go
type Interface interface {
    Name() string                             // Returns the driver name
    Get(key string) generic.Value             // Retrieves a single value
    Has(key string) (bool, generic.Value)     // Checks if a key exists
    All() map[string]generic.Value            // Returns all configuration values
    Set(key string, value interface{}) error  // Sets the value of a key
    SetMulti(data map[string]interface{}) error  // Sets multiple keys at once
    Register(settings ...interface{}) error  // Registers a new key to be used in the future
    Init(params ...string) error             // Initializes the driver (called during application initialization)
}
```

Once you have implemented the interface, create an instance of your struct and add it as a driver:
```go
driver := YourDriver{} // Create an instance of your driver struct
settings.AddDriver(driver) // Add your driver to the configuration system
```

Your driver will now be available for use in the EVO Framework.

---
#### [< Table of Contents](https://github.com/getevo/evo#table-of-contents)