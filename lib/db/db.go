package db

import (
	"context"
	"database/sql"
	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/db/schema/ddl"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Enabled = false

var (
	db *gorm.DB
)

func Register(obj *gorm.DB) {
	if obj != nil {
		Enabled = true
	}
	db = obj
}

// Session create new db session
func Session(config *gorm.Session) *gorm.DB {
	return db.Session(config)
}

// WithContext change current instance db's context to ctx
func WithContext(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

// Debug start debug mode
func Debug() (tx *gorm.DB) {
	return db.Debug()
}

// Set store value with key into current db instance's context
func Set(key string, value any) *gorm.DB {
	return db.Set(key, value)
}

// Get value with key from current db instance's context
func Get(key string) (any, bool) {
	return db.Get(key)
}

// InstanceSet store value with key into current db instance's context
func InstanceSet(key string, value any) *gorm.DB {
	return db.InstanceSet(key, value)
}

// InstanceGet get value with key from current db instance's context
func InstanceGet(key string) (any, bool) {
	return db.InstanceGet(key)
}

// AddError add error to db
func AddError(err error) error {
	return db.AddError(err)
}

// DB returns `*sql.DB`
func DB() (*sql.DB, error) {
	return db.DB()
}

// SetupJoinTable setup join table schema
func SetupJoinTable(model any, field string, joinTable any) error {
	return db.SetupJoinTable(model, field, joinTable)
}

// Use plugin
func Use(plugin gorm.Plugin) error {
	return db.Use(plugin)
}

// ToSQL for generate SQL string.
//
//	db.ToSQL(func(tx *gorm.DB) *gorm.DB {
//			return tx.Model(&User{}).Where(&User{Name: "foo", Age: 20})
//				.Limit(10).Offset(5)
//				.Order("name ASC")
//				.First(&User{})
//	})
func ToSQL(queryFn func(tx *gorm.DB) *gorm.DB) string {
	return db.ToSQL(queryFn)
}

// Create inserts value, returning the inserted data's primary key in value's id
func Create(value any) (tx *gorm.DB) {
	return db.Create(value)
}

// CreateInBatches inserts value in batches of batchSize
func CreateInBatches(value any, batchSize int) (tx *gorm.DB) {
	return db.CreateInBatches(value, batchSize)
}

// Save updates value in database. If value doesn't contain a matching primary key, value is inserted.
func Save(value any) (tx *gorm.DB) {
	return db.Save(value)
}

// First finds the first record ordered by primary key, matching given conditions conds
func First(dest any, conds ...any) (tx *gorm.DB) {
	return db.First(dest, conds...)
}

// Take finds the first record returned by the database in no specified order, matching given conditions conds
func Take(dest any, conds ...any) (tx *gorm.DB) {
	return db.Take(dest, conds...)
}

// Last finds the last record ordered by primary key, matching given conditions conds
func Last(dest any, conds ...any) (tx *gorm.DB) {
	return db.Last(dest, conds...)
}

// Find finds all records matching given conditions conds
func Find(dest any, conds ...any) (tx *gorm.DB) {
	return db.Find(dest, conds...)
}

// FindInBatches finds all records in batches of batchSize
func FindInBatches(dest any, batchSize int, fc func(tx *gorm.DB, batch int) error) *gorm.DB {
	return db.FindInBatches(dest, batchSize, fc)
}

// FirstOrInit finds the first matching record, otherwise if not found initializes a new instance with given conds.
// Each conds must be a struct or map.
//
// FirstOrInit never modifies the database. It is often used with Assign and Attrs.
//
//	// assign an email if the record is not found
//	db.Where(User{Name: "non_existing"}).Attrs(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//
//	// assign email regardless of if record is found
//	db.Where(User{Name: "jinzhu"}).Assign(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "jinzhu", Age: 20, Email: "fake@fake.org"}
func FirstOrInit(dest any, conds ...any) (tx *gorm.DB) {
	return db.FirstOrInit(dest, conds...)
}

// FirstOrCreate finds the first matching record, otherwise if not found creates a new instance with given conds.
// Each conds must be a struct or map.
//
// Using FirstOrCreate in conjunction with Assign will result in an update to the database even if the record exists.
//
//	// assign an email if the record is not found
//	result := db.Where(User{Name: "non_existing"}).Attrs(User{Email: "fake@fake.org"}).FirstOrCreate(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//	// result.RowsAffected -> 1
//
//	// assign email regardless of if record is found
//	result := db.Where(User{Name: "jinzhu"}).Assign(User{Email: "fake@fake.org"}).FirstOrCreate(&user)
//	// user -> User{Name: "jinzhu", Age: 20, Email: "fake@fake.org"}
//	// result.RowsAffected -> 1
func FirstOrCreate(dest any, conds ...any) (tx *gorm.DB) {
	return db.FirstOrCreate(dest, conds...)
}

// Update updates column with value using callbacks. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func Update(column string, value any) (tx *gorm.DB) {
	return db.Update(column, value)
}

// Updates updates attributes using callbacks. values must be a struct or map. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func Updates(values any) (tx *gorm.DB) {
	return db.Updates(values)
}

func UpdateColumn(column string, value any) (tx *gorm.DB) {
	return db.UpdateColumn(column, value)
}

func UpdateColumns(values any) (tx *gorm.DB) {
	return db.UpdateColumns(values)
}

// Delete deletes value matching given conditions. If value contains primary key it is included in the conditions. If
// value includes a deleted_at field, then Delete performs a soft delete instead by setting deleted_at with the current
// time if null.
func Delete(value any, conds ...any) (tx *gorm.DB) {
	return db.Delete(value, conds...)
}

func Count(count *int64) (tx *gorm.DB) {
	return db.Count(count)
}

func Row() *sql.Row {
	return db.Row()
}

func Rows() (*sql.Rows, error) {
	return db.Rows()
}

// Scan scans selected value to the struct dest
func Scan(dest any) (tx *gorm.DB) {
	return db.Scan(dest)
}

// Pluck queries a single column from a model, returning in the slice dest. E.g.:
//
//	var ages []int64
//	db.Model(&users).Pluck("age", &ages)
func Pluck(column string, dest any) (tx *gorm.DB) {
	return db.Pluck(column, dest)
}

func ScanRows(rows *sql.Rows, dest any) error {
	return db.ScanRows(rows, dest)
}

// Connection uses a db connection to execute an arbitrary number of commands in fc. When finished, the connection is
// returned to the connection pool.
func Connection(fc func(tx *gorm.DB) error) (err error) {
	return db.Connection(fc)
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit. Transaction executes an
// arbitrary number of commands in fc within a transaction. On success the changes are committed; if an error occurs
// they are rolled back.
func Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) (err error) {
	return db.Transaction(fc, opts...)
}

// Begin begins a transaction with any transaction options opts
func Begin(opts ...*sql.TxOptions) *gorm.DB {
	return db.Begin(opts...)
}

// Commit commits the changes in a transaction
func Commit() *gorm.DB {
	return db.Commit()
}

// Rollback rollbacks the changes in a transaction
func Rollback() *gorm.DB {
	return db.Rollback()
}

func SavePoint(name string) *gorm.DB {
	return db.SavePoint(name)
}

func RollbackTo(name string) *gorm.DB {
	return db.RollbackTo(name)
}

// Exec executes raw sql
func Exec(sql string, values ...any) (tx *gorm.DB) {
	return db.Exec(sql, values...)
}

// Model specify the model you would like to run db operations
//
//	// update all user's name to `hello`
//	db.Model(&User{}).Update("name", "hello")
//	// if user's primary key is non-blank, will use it as condition, then will only update that user's name to `hello`
//	db.Model(&user).Update("name", "hello")
func Model(value any) (tx *gorm.DB) {
	return db.Model(value)
}

// Clauses Add clauses
//
// This supports both standard clauses (clause.OrderBy, clause.Limit, clause.Where) and more
// advanced techniques like specifying lock strength and optimizer hints. See the
// [docs] for more depth.
//
//	// add a simple limit clause
//	db.Clauses(clause.Limit{Limit: 1}).Find(&User{})
//	// tell the optimizer to use the `idx_user_name` index
//	db.Clauses(hints.UseIndex("idx_user_name")).Find(&User{})
//	// specify the lock strength to UPDATE
//	db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&users)
//
// [docs]: https://gorm.io/docs/sql_builder.html#Clauses
func Clauses(conds ...clause.Expression) (tx *gorm.DB) {
	return db.Clauses(conds...)
}

// Table specify the table you would like to run db operations
//
//	// Get a user
//	db.Table("users").take(&result)
func Table(name string, args ...any) (tx *gorm.DB) {
	return db.Table(name, args...)
}

// Distinct specify distinct fields that you want querying
//
//	// Select distinct names of users
//	db.Distinct("name").Find(&results)
//	// Select distinct name/age pairs from users
//	db.Distinct("name", "age").Find(&results)
func Distinct(args ...any) (tx *gorm.DB) {
	return db.Distinct(args...)
}

// Select specify fields that you want when querying, creating, updating
//
// Use Select when you only want a subset of the fields. By default, GORM will select all fields.
// Select accepts both string arguments and arrays.
//
//	// Select name and age of user using multiple arguments
//	db.Select("name", "age").Find(&users)
//	// Select name and age of user using an array
//	db.Select([]string{"name", "age"}).Find(&users)
func Select(query any, args ...any) (tx *gorm.DB) {
	return db.Select(query, args...)
}

// Omit specify fields that you want to ignore when creating, updating and querying
func Omit(columns ...string) (tx *gorm.DB) {
	return db.Omit(columns...)
}

// Where add conditions
//
// See the [docs] for details on the various formats that where clauses can take. By default, where clauses chain with AND.
//
//	// Find the first user with name jinzhu
//	db.Where("name = ?", "jinzhu").First(&user)
//	// Find the first user with name jinzhu and age 20
//	db.Where(&User{Name: "jinzhu", Age: 20}).First(&user)
//	// Find the first user with name jinzhu and age not equal to 20
//	db.Where("name = ?", "jinzhu").Where("age <> ?", "20").First(&user)
//
// [docs]: https://gorm.io/docs/query.html#Conditions
func Where(query any, args ...any) (tx *gorm.DB) {
	return db.Where(query, args...)
}

// Not add NOT conditions
//
// Not works similarly to where, and has the same syntax.
//
//	// Find the first user with name not equal to jinzhu
//	db.Not("name = ?", "jinzhu").First(&user)
func Not(query any, args ...any) (tx *gorm.DB) {
	return db.Not(query, args...)
}

// Or add OR conditions
//
// Or is used to chain together queries with an OR.
//
//	// Find the first user with name equal to jinzhu or john
//	db.Where("name = ?", "jinzhu").Or("name = ?", "john").First(&user)
func Or(query any, args ...any) (tx *gorm.DB) {
	return db.Or(query, args...)
}

// Joins specify Joins conditions
//
//	db.Joins("Account").Find(&user)
//	db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
//	db.Joins("Account", DB.Select("id").Where("user_id = users.id AND name = ?", "someName").Model(&Account{}))
func Joins(query string, args ...any) (tx *gorm.DB) {
	return db.Joins(query, args...)
}

// InnerJoins specify inner joins conditions
// db.InnerJoins("Account").Find(&user)
func InnerJoins(query string, args ...any) (tx *gorm.DB) {
	return db.InnerJoins(query, args...)
}

// Group specify the group method on the find
//
//	// Select the sum age of users with given names
//	db.Model(&User{}).Select("name, sum(age) as total").Group("name").Find(&results)
func Group(name string) (tx *gorm.DB) {
	return db.Group(name)
}

// Having specify HAVING conditions for GROUP BY
//
//	// Select the sum age of users with name jinzhu
//	db.Model(&User{}).Select("name, sum(age) as total").Group("name").Having("name = ?", "jinzhu").Find(&result)
func Having(query any, args ...any) (tx *gorm.DB) {
	return db.Having(query, args...)
}

// Order specify order when retrieving records from database
//
//	db.Order("name DESC")
//	db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func Order(value any) (tx *gorm.DB) {
	return db.Order(value)
}

// Limit specify the number of records to be retrieved
//
// Limit conditions can be cancelled by using `Limit(-1)`.
//
//	// retrieve 3 users
//	db.Limit(3).Find(&users)
//	// retrieve 3 users into users1, and all users into users2
//	db.Limit(3).Find(&users1).Limit(-1).Find(&users2)
func Limit(limit int) (tx *gorm.DB) {
	return db.Limit(limit)
}

// Offset specify the number of records to skip before starting to return the records
//
// Offset conditions can be cancelled by using `Offset(-1)`.
//
//	// select the third user
//	db.Offset(2).First(&user)
//	// select the first user by cancelling an earlier chained offset
//	db.Offset(5).Offset(-1).First(&user)
func Offset(offset int) (tx *gorm.DB) {
	return db.Offset(offset)
}

// Scopes pass current database connection to arguments `func(DB) DB`, which could be used to add conditions dynamically
//
//	func AmountGreaterThan1000(db *gorm.DB) *gorm.DB {
//	    return db.Where("amount > ?", 1000)
//	}
//
//	func OrderStatus(status []string) func  *gorm.DB {
//	    return func  *gorm.DB {
//	        return db.Scopes(AmountGreaterThan1000).Where("status in (?)", status)
//	    }
//	}
//
//	db.Scopes(AmountGreaterThan1000, OrderStatus([]string{"paid", "shipped"})).Find(&orders)
func Scopes(funcs ...func(*gorm.DB) *gorm.DB) (tx *gorm.DB) {
	return db.Scopes(funcs...)
}

// Preload preload associations with given conditions
//
//	// get all users, and preload all non-cancelled orders
//	db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
func Preload(query string, args ...any) (tx *gorm.DB) {
	return db.Preload(query, args...)
}

// Attrs provide attributes used in [FirstOrCreate] or [FirstOrInit]
//
// Attrs only adds attributes if the record is not found.
//
//	// assign an email if the record is not found
//	db.Where(User{Name: "non_existing"}).Attrs(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//
//	// assign an email if the record is not found, otherwise ignore provided email
//	db.Where(User{Name: "jinzhu"}).Attrs(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "jinzhu", Age: 20}
//
// [FirstOrCreate]: https://gorm.io/docs/advanced_query.html#FirstOrCreate
// [FirstOrInit]: https://gorm.io/docs/advanced_query.html#FirstOrInit
func Attrs(attrs ...any) (tx *gorm.DB) {
	return db.Attrs(attrs...)
}

// Assign provide attributes used in [FirstOrCreate] or [FirstOrInit]
//
// Assign adds attributes even if the record is found. If using FirstOrCreate, this means that
// records will be updated even if they are found.
//
//	// assign an email regardless of if the record is not found
//	db.Where(User{Name: "non_existing"}).Assign(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//
//	// assign email regardless of if record is found
//	db.Where(User{Name: "jinzhu"}).Assign(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "jinzhu", Age: 20, Email: "fake@fake.org"}
//
// [FirstOrCreate]: https://gorm.io/docs/advanced_query.html#FirstOrCreate
// [FirstOrInit]: https://gorm.io/docs/advanced_query.html#FirstOrInit
func Assign(attrs ...any) (tx *gorm.DB) {
	return Assign(attrs...)
}

func Unscoped() (tx *gorm.DB) {
	return db.Unscoped()
}

func Raw(sql string, values ...any) (tx *gorm.DB) {
	return db.Raw(sql, values...)
}

func UseModel(models ...any) {
	schema.UseModel(db, models...)
}

func GetMigrationScript() []string {
	return schema.GetMigrationScript(db)
}

func DoMigration() error {
	return schema.DoMigration(db)
}

func Models() []schema.Model {
	return schema.Models
}

func GetModel(name string) *schema.Model {
	return schema.Find(name)
}

func SetDefaultCollation(collation string) {
	ddl.DefaultCollation = collation
}

func SetDefaultCharset(charset string) {
	ddl.DefaultCharset = charset
}

func SetDefaultEngine(engine string) {
	ddl.DefaultEngine = engine
}

func IsEnabled() bool {
	return Enabled
}

var _onContext []func(v interface{}) *gorm.DB

func OnPrepareContext(fn func(v interface{}) *gorm.DB) {
	_onContext = append(_onContext, fn)
}

func GetContext() *gorm.DB {
	var dbo = db
	for _, fn := range _onContext {
		dbo = fn(dbo)
	}
	return dbo
}
