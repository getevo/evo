# Database
By enabling the database in the **`config.yml`**, you can start using your database connection extensively in your applications. The EVO Framework currently supports the following database engines:

- MySQL
- TiDB
- Microsoft SQL
- SQLite

Here is an example of the database configuration with detailed settings:

```yaml
#config.yml
Database:
    # If enabled, it will cache the constructed queries to save processing time
    Cache: "false"

    # Indicates the maximum duration for which a connection can remain idle without being closed by the server
    ConnMaxLifTime: 1h

    # Specifies the name of the database
    Database: "service_test"

    # Sets the verbosity level of the logging: 1:silent, 2:warn, 3:error, 4:info
    Debug: 3

    # Enables the database connection
    Enabled: true

    # Specifies the maximum number of idle connections allowed
    MaxIdleConns: "10"

    # Specifies the maximum number of concurrent connections allowed
    MaxOpenConns: "100"

    # Additional parameters to be included in the connection string
    Params: "charset=utf8mb4&parseTime=True&loc=Local"

    # Specifies the password for the database
    Password: ""

    # Enables support for SSL
    SSLMode: "false"

    # Specifies the server address or SQLite file path
    Server: 127.0.01:3306

    # Defines the threshold duration for query execution. If a query exceeds this duration, a warning will be issued.
    SlowQueryThreshold: 500ms

    # Specifies the type of database: mysql, mssql, sqlite
    Type: mysql

    # Specifies the username for the database
    Username: root
```

By configuring the database settings according to your requirements, you can utilize the database connection seamlessly within your EVO Framework applications.

### Accessing the database by creating a new instance
To access the database, you can obtain a database object, which is an instance of **`gorm.DB`**, by requesting the database instance and saving it in a variable:

```go
var user User
var db = evo.GetDBO()
db.Find(&user)
```

### Accessing the database using the db library
Alternatively, you can access the database object directly using the **db** library, which is a wrapper around the **gorm.DB** instance:
```go
var user User
db.Find(&user)
```

For further information on how to use **gorm**, you can refer to the **[gorm documentation](https://gorm.io/docs)**. The documentation provides detailed guidance on using the gorm library for database operations.


