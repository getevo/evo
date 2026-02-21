# EVO Framework

EVO Framework is a backend development solution designed to facilitate efficient development using the Go programming language. It is built with a focus on modularity and follows the MVC (Model-View-Controller) architectural pattern. The core of EVO Framework is highly extensible, allowing for seamless extension or replacement of its main modules.

## Key Features

- **Modularity**: EVO Framework promotes modularity, enabling developers to structure their codebase in a modular manner.
- **MVC Structure**: Following the widely adopted MVC pattern, EVO Framework separates concerns and improves code organization.
- **Comprehensive Toolset**: EVO Framework provides a rich set of tools, eliminating the need for developers to deal with low-level libraries and technologies.
- **Enhanced Readability**: By leveraging the EVO Framework, your code becomes more readable and clear, enhancing collaboration and maintainability.

With EVO Framework, you can focus on your programming logic and rapidly develop robust backend solutions without getting bogged down by intricate implementation details.

## Table of Contents
- [Getting Started](#getting-started)
- [Configuration and Settings](docs/configuration.md)
- [Database](docs/database.md)
  - [MySQL Driver](docs/mysql.md)
  - [PostgreSQL Driver](docs/pgsql.md)
  - [Database Migration](docs/migration.md)
- [Health Checks](docs/health-checks.md)
- [HTTP Responses (Outcome)](docs/outcome.md)
- [Modules](#libraries)
- [Web Server](docs/webserver.md)
- [Storage Interface](lib/storage/README.md)
- [Cache Interface](lib/memo/README.md)
- [Messaging Bus Interface](lib/pubsub/README.md)
- [Date](lib/date/README.md)
- [Internationalization](lib/text/README.md)
- [TPL Templates](docs/tpl.md)
- [CURL](docs/curl.md)
- [Build](lib/build/README.md)
- [Args](docs/args.md)
- [Logging](docs/log.md)
  - [File Logger](docs/file_logger.md)
- [Concurrency Patterns](lib/async/README.md)
- [STract configuration language](lib/stract/README.md)
- [Local Files](lib/gpath/README.md)
- [Dot Notation](docs/dot.md)
- [Reflections](docs/reflections.md)
- [Validations](docs/validation.md)
- [GORM based pagination](https://github.com/getevo/pagination)

## Libraries
EVO Framework includes a comprehensive set of libraries in the `lib` directory:

- [Application](lib/application/README.md) - Application management
- [Args](lib/args/README.md) - Command-line argument handling
- [Async](lib/async/README.md) - Asynchronous operations
- [Build](lib/build/README.md) - Build utilities
- [Connectors](lib/connectors/README.md) - External system connectors
- [Curl](lib/curl/README.md) - HTTP client functionality
- [Date](lib/date/README.md) - Date and time utilities
- [DB](lib/db/README.md) - Database operations
- [Dot](lib/dot/README.md) - Dot notation for accessing nested data
- [Errors](lib/errors/README.md) - Error handling
- [Frm](lib/frm/README.md) - Form handling
- [Generic](lib/generic/README.md) - Generic data structures
- [Gpath](lib/gpath/README.md) - Path manipulation
- [Is](lib/is/README.md) - Type checking utilities
- [JSON](lib/json/README.md) - JSON handling
- [Log](lib/log/README.md) - Logging system
- [Memo](lib/memo/README.md) - Memoization and caching
- [Model](lib/model/README.md) - Data modeling
- [Outcome](lib/outcome/README.md) - Function result handling
- [Panics](lib/panics/README.md) - Panic handling
- [Ptr](lib/ptr/README.md) - Pointer utilities
- [Pubsub](lib/pubsub/README.md) - Publish-subscribe pattern
- [Reflections](lib/reflections/README.md) - Reflection utilities
- [Scheduler](lib/scheduler/README.md) - Task scheduling
- [Serializer](lib/serializer/README.md) - Data serialization
- [Settings](lib/settings/README.md) - Configuration management
- [Storage](lib/storage/README.md) - Storage interfaces
- [Stract](lib/stract/README.md) - Structured data handling
- [Text](lib/text/README.md) - Text manipulation
- [Tpl](lib/tpl/README.md) - Template system
- [Try](lib/try/README.md) - Error handling with try/catch pattern
- [Validation](lib/validation/README.md) - Data validation
- [Version](lib/version/README.md) - Version management

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
3. Initialize the EVO Framework and start building your application:

**MySQL:**
```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/mysql"
)

func main() {
    evo.Setup(mysql.Driver{})
    evo.Run()
}
```

**PostgreSQL:**
```go
package main

import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/pgsql"
)

func main() {
    evo.Setup(pgsql.Driver{})
    evo.Run()
}
```

