package db

import (
	"context"
	"database/sql"
	"github.com/getevo/evo/v2"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func Register() {
	db = evo.GetDBO()
	db.Row()
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
func Set(key string, value interface{}) *gorm.DB {
	return db.Set(key, value)
}

// Get get value with key from current db instance's context
func Get(key string) (interface{}, bool) {
	return db.Get(key)
}

// InstanceSet store value with key into current db instance's context
func InstanceSet(key string, value interface{}) *gorm.DB {
	return db.InstanceSet(key, value)
}

// InstanceGet get value with key from current db instance's context
func InstanceGet(key string) (interface{}, bool) {
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
func SetupJoinTable(model interface{}, field string, joinTable interface{}) error {
	return db.SetupJoinTable(model, field, joinTable)
}

// Use use plugin
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
func Create(value interface{}) (tx *gorm.DB) {
	return db.Create(value)
}

// CreateInBatches inserts value in batches of batchSize
func CreateInBatches(value interface{}, batchSize int) (tx *gorm.DB) {
	return db.CreateInBatches(value, batchSize)
}

// Save updates value in database. If value doesn't contain a matching primary key, value is inserted.
func Save(value interface{}) (tx *gorm.DB) {
	return db.Save(value)
}

// First finds the first record ordered by primary key, matching given conditions conds
func First(dest interface{}, conds ...interface{}) (tx *gorm.DB) {
	return db.First(dest, conds...)
}

// Take finds the first record returned by the database in no specified order, matching given conditions conds
func Take(dest interface{}, conds ...interface{}) (tx *gorm.DB) {
	return db.Take(dest, conds...)
}

// Last finds the last record ordered by primary key, matching given conditions conds
func Last(dest interface{}, conds ...interface{}) (tx *gorm.DB) {
	return db.Last(dest, conds...)
}

// Find finds all records matching given conditions conds
func Find(dest interface{}, conds ...interface{}) (tx *gorm.DB) {
	return db.Find(dest, conds...)
}

// FindInBatches finds all records in batches of batchSize
func FindInBatches(dest interface{}, batchSize int, fc func(tx *gorm.DB, batch int) error) *gorm.DB {
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
func FirstOrInit(dest interface{}, conds ...interface{}) (tx *gorm.DB) {
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
func FirstOrCreate(dest interface{}, conds ...interface{}) (tx *gorm.DB) {
	return db.FirstOrCreate(dest, conds...)
}

// Update updates column with value using callbacks. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func Update(column string, value interface{}) (tx *gorm.DB) {
	return db.Update(column, value)
}

// Updates updates attributes using callbacks. values must be a struct or map. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func Updates(values interface{}) (tx *gorm.DB) {
	return db.Updates(values)
}

func UpdateColumn(column string, value interface{}) (tx *gorm.DB) {
	return db.UpdateColumn(column, value)
}

func UpdateColumns(values interface{}) (tx *gorm.DB) {
	return db.UpdateColumns(values)
}

// Delete deletes value matching given conditions. If value contains primary key it is included in the conditions. If
// value includes a deleted_at field, then Delete performs a soft delete instead by setting deleted_at with the current
// time if null.
func Delete(value interface{}, conds ...interface{}) (tx *gorm.DB) {
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
func Scan(dest interface{}) (tx *gorm.DB) {
	return db.Scan(dest)
}

// Pluck queries a single column from a model, returning in the slice dest. E.g.:
//
//	var ages []int64
//	db.Model(&users).Pluck("age", &ages)
func Pluck(column string, dest interface{}) (tx *gorm.DB) {
	return db.Pluck(column, dest)
}

func ScanRows(rows *sql.Rows, dest interface{}) error {
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
func Exec(sql string, values ...interface{}) (tx *gorm.DB) {
	return db.Exec(sql, values...)
}
