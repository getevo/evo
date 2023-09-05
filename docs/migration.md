# Database Migration
Database migration poses a fundamental challenge for developers. Synchronizing data structures and versioning represents a significant source of frustration for both developers and database administrators. Evo has introduced a straightforward solution, making it incredibly easy for developers to keep their required database structures aligned with the demands of their applications.

> The EVO migration system conducts a routine check of the database structure during each startup. It then compares this structure with the application's requirements and endeavors to apply any necessary patches if discrepancies are detected.

## Warning
> Altering column in a database may potentially lead to data loss or corruption of text data. Collation determines how string comparison and sorting operations are performed in a database, including how data are interpreted and ordered.
Before altering the collation of a column, it's crucial to:
> - **Back up your data:** Always make a full backup of your database before making any significant changes like altering collations.
> - **Understand the implications:** Carefully assess how the change in collation may impact your data and application. Test the changes in a controlled environment if possible.
> - **Plan and script the changes:** Develop a well-thought-out plan for altering collations, including any necessary data conversion or migration steps.
> - **Consider data migration:** In some cases, it may be safer to export the data, create a new column with the desired collation, and then import the data into the new column.
> - **Involve database administrators:** If you're not experienced with database administration, consider consulting with a DBA or database expert who can help you make informed decisions and execute the changes safely.

## Compatibility
>Whether you're working with MySQL, MariaDB, or TiDB, EVO's migrator can help streamline the process of managing and evolving your database schema to meet the needs of your application.
## Usage

To use this package, import it in your Go code:

```go
import (
    "fmt"
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/db"
    "github.com/getevo/evo/v2/lib/db/schema"
    "github.com/getevo/evo/v2/lib/log"
    "testing"
    "time"
)

// Define the model
type Model struct {
    Identifier    string `gorm:"column:identifier;primaryKey;size:255" `
    Name          string `gorm:"column:name;"`
    Type          string `gorm:"column:type;type:enum('user','admin','developer')"`
    Invoker       string `gorm:"column:invoker;size:512;index"`
    Price         float64 `gorm:"column:price;precision:2;scale:2"`
    NullableField *string `gorm:"column:nullable_field;size:512"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

// Set model table name
func (Model) TableName() string {
    return "my_model"
}

// create versioned schema with roll back on DML queries.
func (Model) Migration(currentVersion string) []schema.Migration {
    if version.Compare(currentVersion, "1", "<") {
        db.Exec("-- SOME QUERY HERE")
    }
    return []schema.Migration{
        {"0.0.1", "ALTER TABLE my_model AUTO_INCREMENT = " + fmt.Sprint(time.Now().Unix())},
    }
}
func main() {
    evo.Setup()
	
    // introduce model to list of our models
    db.UseModel(&Model{})

    // get and print migration queries
    var queries []string = db.GetMigrationScript()
    for _,query := range queries{
	    	fmt.Println(query)
    }
	
    // run migration 
    var err = db.DoMigration()
    log.Error(err)

}
```

This Go code setting up a database migration process. Here's a breakdown of what the code does:
 
- It defines a Go struct called Model which to represent a database table named "my_model." This struct has several fields with tags indicating their properties, such as column names and types.
- The TableName method is defined for the Model struct, specifying the name of the database table associated with this model.
- The Migration method is also defined for the Model struct, returning a slice of schema.Migration structs. Each of these structs represents a database migration step with a version and SQL query to be executed when migrating the database.

## Model attributes
### Column name
The column attributes specify the name of a column within a table. If you do not explicitly define a column name, it will default to using the attribute name converted into snake case.
```go
    type Model struct{
	    Column string `gorm:"column:struct"`
    }
```
---
### Column type
You can also specify the SQL data type for a column. If you do not set a specific type, the migration script will use a default type based on the data type of the attribute within the struct.
```go
    type Model struct{
	    Column string `gorm:"type:varchar(255)"`
    }
```
> **BEWARE: Changing type should be done cautiously and with a clear understanding of the potential risks involved. Data integrity and consistency are paramount when working with databases.**
---
### Column size
You can also specify the SQL data size for a column.
```go
    type Model struct{
	    Column string `gorm:"size:50"`
    }
```
> **BEWARE: Changing size should be done cautiously and with a clear understanding of the potential risks involved. Data integrity and consistency are paramount when working with databases.**

---

### Column precision and scale
You can also specify the SQL data precision and scale for a column.
```go
    type Model struct{
	    Salary float64 `gorm:"type:DECIMAL;precision:2;scale:5"`
    }
```
In this example, 5 is the precision and 2 is the scale. The precision represents the number of significant digits that are stored for values, and the scale represents the number of digits that can be stored following the decimal point.

Standard SQL requires that DECIMAL(5,2) be able to store any value with five digits and two decimals, so values that can be stored in the salary column range from -999.99 to 999.99.
> **BEWARE: Changing precision and scale should be done cautiously and with a clear understanding of the potential risks involved. Data integrity and consistency are paramount when working with databases.**
---

### Index
You can configure a column to be indexed, and there are various examples demonstrating different methods of creating an index.
In this Model struct:

- The ID field is marked as the primary key using gorm:"primaryKey".
- The Column field is indexed with an auto-generated index name.
- The Column2 field is indexed with a specific custom index name, "column_2_index."
- The Column3 and Column4 fields are indexed together as a composite index with a custom name, "my_idx."
- The Column5 field has a unique constraint.
- The Column6 and Column7 fields have a unique constraint on both columns with a custom unique index name, "my_unique_idx."
```go
    type Model struct{
       //Set column as primary column
       ID int `gorm:"primaryKey`

       //index single column with auto naming
       Column string `gorm:"index`

       //index single column with specific name
       Column2 string `gorm:"index:column_2_index`

       //index multiple columns
       Column3 string `gorm:"index:my_idx`
       Column4 int `gorm:"index:my_idx`

       //unique value column
       Column5 string `gorm:"unique`

       //unique multiple columns
       Column6 string `gorm:"unique:my_unique_idx`
       Column7 int `gorm:"unique:my_unique_idx`
    }
```
---
### AutoIncrement
The `autoIncrement` attribute is specified to enable auto-increment functionality for this column, typically used for auto-generating unique identifiers.
```go
    type Model struct{
       //Set column as primary column
       ID int `gorm:"primaryKey;autoIncrement"`
    }
```
---

### Charset
Using the `charset` tag in GORM, you can specify your preferred character set for specific text columns in your database schema. 
```go
    type Model struct{
       Name string `gorm:"charset:utf8mb4"`
    }
```
> **BEWARE: Changing collations should be done cautiously and with a clear understanding of the potential risks involved. Data integrity and consistency are paramount when working with databases.**
---

### Collate
Using the `collate` tag in GORM, you can specify your preferred collation for specific text columns in your database schema.
```go
    type Model struct{
       Name string `gorm:"collate:utf8mb4_unicode"`
    }
```
> **BEWARE: Changing collations should be done cautiously and with a clear understanding of the potential risks involved. Data integrity and consistency are paramount when working with databases.**
---

### Default
Using the `default` tag in GORM, you can specify your preferred default value for specific column in your database schema.
```go
    type Model struct{
       Name string `gorm:"default:'hello world'"`
    }
```
---

### Comment
Using the `comment` tag in GORM, you can specify your preferred comment for specific column in your database schema.
```go
    type Model struct{
       Name string `gorm:"comment:'hello world'"`
    }
```
---

### Nullable
Setting a Go variable type to a pointer for a field within a GORM model will make the corresponding column in the database nullable. When a column is nullable, it means that it can contain NULL values, indicating the absence of a value. This is particularly useful when you want to allow certain fields in your database to have missing or unknown values.
```go
    type Model struct {
        ID   int     `gorm:"primaryKey"`
        Name *string `gorm:"column:name"`
    }
```

Also  By including the "null" option in the GORM tag for a field, you explicitly indicate that the field is nullable in the database schema. This can be particularly useful when you want to provide clarity in your code about the nullability of specific fields.
```go
type Model struct {
    ID   int     `gorm:"primaryKey"`
    Name string `gorm:"column:name;null"`
}
```
---

### Timestamps
Migrator handles date and time migrations seamlessly. When using the time.Time type for a field in your Go model, the Evo migrator will automatically set the data type for that column to "timestamp" in the database schema.

Here's a basic example:
```go
type Model struct {
    ID        int
    Date time.Time `gorm:"column:date"`
}
```
It's common to have timestamp fields like `created_at`, `updated_at`, and `deleted_at` for tracking activities in database records. Migrator simplifies this process by providing these predefined date fields that you can copy and paste directly into your Go struct. Here's an example of how you can use them:

```go
type Model struct {
    ID        int
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}
```



---
### Table options

In addition to the struct definition, there are methods implemented for customizing the table's name, charset, and collation. The `TableName()`, `TableCharset()`, and `TableCollation()` methods allow developers to specify the name of the database table, the character set, and the collation used for that table. This level of customization can be valuable when tailoring the database schema to specific project requirements, especially when dealing with internationalization and character encoding considerations.

```go
type Model struct {
    ID        int
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

// Set Table Name
func (Model)TableName() string{
	return "table_name"
}

// Set Table Charset
func (Model)TableCharset() string{
    return "utf8"
}

// Set Table Collation
func (Model)TableCollation() string{
    return "utf8mb4_bin"
}
```
---
### Change Defaults
In this Go code snippet, database settings are being customized using the db package. These settings are applied globally and will affect all database operations in the application.
```go
    // Change default table engine
    db.SetDefaultEngine("MEMORY")

    // Change default charset
    db.SetDefaultCharset("utf8mb4")
    
    // Change default collation
    db.SetDefaultCollation("utf8mb4_bin")
```
---
### Custom migrations
The migrator offers valuable capabilities beyond just migrating database schema changes. It can also handle data migration and transformation as needed, making it a powerful tool for ensuring data consistency and structure during application updates. Additionally, the migrator can manage versioning of the database schema, which is useful for tracking and applying incremental changes over time.

**It's essential to be aware that the migrator uses table comments to store version information.** If you plan to use table comments in your database for other purposes, such as documentation or annotations, the migrator's versioning mechanism may overwrite or interfere with your existing comments. To avoid conflicts, consider carefully how you use table comments and whether they may be affected by the migrator's versioning approach. Clear documentation and communication within your development team can help ensure that everyone understands how table comments are being utilized in your database schema.

```go

// Define the model
type Model struct {
    Identifier    string `gorm:"column:identifier;primaryKey;size:255" `
    Name          string `gorm:"column:name;"`
    Type          string `gorm:"column:type;type:enum('user','admin','developer')"`
    Invoker       string `gorm:"column:invoker;size:512;index"`
    Price         float64 `gorm:"column:price;precision:2;scale:2"`
    NullableField *string `gorm:"column:nullable_field;size:512"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}


// create versioned schema with roll back on DML queries.
func (Model) Migration(currentVersion string) []schema.Migration {
    if version.Compare(currentVersion, "1", "<") {
        db.Exec("-- SOME QUERY HERE")
		// do some stuff here
    }
    return []schema.Migration{
        {"0.0.1", "ALTER TABLE my_model AUTO_INCREMENT = " + fmt.Sprint(time.Now().Unix())},
    }
}
```
---
#### [< Table of Contents](https://github.com/getevo/evo#table-of-contents)