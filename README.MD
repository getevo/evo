# EVO Framework

EVO Framework is a backend development solution designed to facilitate efficient development using the Go programming language. It is built with a focus on modularity and follows the MVC (Model-View-Controller) architectural pattern. The core of EVO Framework is highly extensible, allowing for seamless extension or replacement of its main modules.

## Key Features

- **Modularity**: EVO Framework promotes modularity, enabling developers to structure their codebase in a modular manner.
- **MVC Structure**: Following the widely adopted MVC pattern, EVO Framework separates concerns and improves code organization.
- **Comprehensive Toolset**: EVO Framework provides a rich set of tools, eliminating the need for developers to deal with low-level libraries and technologies.
- **Enhanced Readability**: By leveraging the EVO Framework, your code becomes more readable and clear, enhancing collaboration and maintainability.

With EVO Framework, you can focus on your programming logic and rapidly develop robust backend solutions without getting bogged down by intricate implementation details.

#### For detailed usage instructions and examples, refer to the [Table of Contents](docs/TOC.MD)

## Getting Started

To get started with EVO Framework, follow these steps:

1. Install EVO Framework by running the following command:
   ```shell
   $ go get github.com/getevo/evo/v2
   ```
2. Create Minimum Configuration File
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
   Server: 192.168.1.102:3306
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
3. Initialize the EVO Framework and start building your application:
```go
package main

import (
   "github.com/getevo/evo/v2"
)

func main() {
    // initialize evo
    evo.Setup()
	
    //your code goes here ...
    evo.Run()
}

```
