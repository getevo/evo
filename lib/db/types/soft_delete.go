package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// SoftDeletedAt is a nullable time type used as the DeletedAt field in
// types.SoftDelete. It implements GORM's QueryClauses, UpdateClauses, and
// DeleteClauses interfaces so that:
//
//   - SELECT/UPDATE queries automatically gain WHERE deleted_at IS NULL.
//   - db.Delete() is converted to UPDATE SET deleted_at=NOW(), deleted=1
//     instead of issuing a hard DELETE.
//
// Use db.Unscoped() to bypass the filter and see or hard-delete records.
type SoftDeletedAt sql.NullTime

// Scan implements the sql.Scanner interface.
func (n *SoftDeletedAt) Scan(value interface{}) error {
	return (*sql.NullTime)(n).Scan(value)
}

// Value implements the driver.Valuer interface.
func (n SoftDeletedAt) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

// MarshalJSON implements json.Marshaler.
func (n SoftDeletedAt) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *SoftDeletedAt) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &n.Time)
	if err == nil {
		n.Valid = true
	}
	return err
}

// GormDataType returns the GORM column data type.
func (SoftDeletedAt) GormDataType() string {
	return "datetime"
}

// QueryClauses implements schema.QueryClausesInterface.
// GORM calls this during schema parsing and registers the returned clauses
// so that all SELECT statements gain WHERE deleted_at IS NULL automatically.
func (SoftDeletedAt) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{softDeleteQueryClause{Field: f}}
}

// UpdateClauses implements schema.UpdateClausesInterface.
// Ensures the WHERE deleted_at IS NULL filter is also applied on UPDATE.
func (SoftDeletedAt) UpdateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{softDeleteUpdateClause{Field: f}}
}

// DeleteClauses implements schema.DeleteClausesInterface.
// Converts DELETE into UPDATE SET deleted_at=NOW(), deleted=1.
func (SoftDeletedAt) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{softDeleteDeleteClause{Field: f}}
}

// NullTime returns the underlying sql.NullTime.
func (n SoftDeletedAt) NullTime() sql.NullTime {
	return sql.NullTime(n)
}

// TimeValue returns the time.Time value and whether it is valid (non-NULL).
func (n SoftDeletedAt) TimeValue() (time.Time, bool) {
	return n.Time, n.Valid
}

// ---- query clause -------------------------------------------------------

type softDeleteQueryClause struct {
	Field *schema.Field
}

func (softDeleteQueryClause) Name() string               { return "" }
func (softDeleteQueryClause) Build(clause.Builder)        {}
func (softDeleteQueryClause) MergeClause(*clause.Clause)  {}

func (sd softDeleteQueryClause) ModifyStatement(stmt *gorm.Statement) {
	if _, ok := stmt.Clauses["soft_delete_enabled"]; ok || stmt.Statement.Unscoped {
		return
	}

	// Wrap existing OR conditions in AND to avoid logic errors when combined
	// with the soft-delete predicate (mirrors gorm.SoftDeleteQueryClause).
	if c, ok := stmt.Clauses["WHERE"]; ok {
		if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) >= 1 {
			for _, expr := range where.Exprs {
				if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
					where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
					c.Expression = where
					stmt.Clauses["WHERE"] = c
					break
				}
			}
		}
	}

	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: sd.Field.DBName}, Value: nil},
	}})
	stmt.Clauses["soft_delete_enabled"] = clause.Clause{}
}

// ---- update clause -------------------------------------------------------

type softDeleteUpdateClause struct {
	Field *schema.Field
}

func (softDeleteUpdateClause) Name() string               { return "" }
func (softDeleteUpdateClause) Build(clause.Builder)       {}
func (softDeleteUpdateClause) MergeClause(*clause.Clause) {}

func (sd softDeleteUpdateClause) ModifyStatement(stmt *gorm.Statement) {
	if stmt.SQL.Len() == 0 && !stmt.Statement.Unscoped {
		softDeleteQueryClause{Field: sd.Field}.ModifyStatement(stmt)
	}
}

// ---- delete clause -------------------------------------------------------

type softDeleteDeleteClause struct {
	Field *schema.Field
}

func (softDeleteDeleteClause) Name() string               { return "" }
func (softDeleteDeleteClause) Build(clause.Builder)       {}
func (softDeleteDeleteClause) MergeClause(*clause.Clause) {}

func (sd softDeleteDeleteClause) ModifyStatement(stmt *gorm.Statement) {
	if stmt.SQL.Len() != 0 || stmt.Statement.Unscoped {
		return
	}

	curTime := stmt.DB.NowFunc()
	nowSDA := SoftDeletedAt{Valid: true, Time: curTime}

	// Build a single SET clause with both assignments.
	// NOTE: In GORM v1.30+ clause.Set.MergeClause replaces rather than
	// appends, so we must pass both assignments in one AddClause call.
	set := clause.Set{{Column: clause.Column{Name: sd.Field.DBName}, Value: nowSDA}}
	if deletedField := sd.Field.Schema.LookUpField("Deleted"); deletedField != nil {
		set = append(set, clause.Assignment{Column: clause.Column{Name: deletedField.DBName}, Value: true})
		stmt.SetColumn(deletedField.DBName, true, true)
	}
	stmt.AddClause(set)
	stmt.SetColumn(sd.Field.DBName, nowSDA, true)

	if stmt.Schema != nil {
		_, queryValues := schema.GetIdentityFieldValuesMap(stmt.Context, stmt.ReflectValue, stmt.Schema.PrimaryFields)
		column, values := schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)

		if len(values) > 0 {
			stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
		}

		if stmt.ReflectValue.CanAddr() && stmt.Dest != stmt.Model && stmt.Model != nil {
			_, queryValues = schema.GetIdentityFieldValuesMap(stmt.Context, reflect.ValueOf(stmt.Model), stmt.Schema.PrimaryFields)
			column, values = schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)

			if len(values) > 0 {
				stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
			}
		}
	}

	softDeleteQueryClause{Field: sd.Field}.ModifyStatement(stmt)
	stmt.AddClauseIfNotExists(clause.Update{})
	stmt.Build(stmt.DB.Callback().Update().Clauses...)
}
