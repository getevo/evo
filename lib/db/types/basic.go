package types

import (
	"time"
)

// CreatedAt represents the timestamp of when an entity or object was created.
// It uses GORM's autoCreateTime for automatic timestamp management.
type CreatedAt struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// Time returns the underlying time.Time value
func (c CreatedAt) Time() time.Time {
	return c.CreatedAt
}

// Set sets the CreatedAt timestamp
func (c *CreatedAt) Set(t time.Time) {
	c.CreatedAt = t
}

// After reports whether the CreatedAt time is after u
func (c CreatedAt) After(u time.Time) bool {
	return c.CreatedAt.After(u)
}

// Before reports whether the CreatedAt time is before u
func (c CreatedAt) Before(u time.Time) bool {
	return c.CreatedAt.Before(u)
}

// Equal reports whether the CreatedAt time is equal to u
func (c CreatedAt) Equal(u time.Time) bool {
	return c.CreatedAt.Equal(u)
}

// Add returns the time t+d
func (c CreatedAt) Add(d time.Duration) time.Time {
	return c.CreatedAt.Add(d)
}

// Sub returns the duration t-u
func (c CreatedAt) Sub(u time.Time) time.Duration {
	return c.CreatedAt.Sub(u)
}

// IsZero reports whether the CreatedAt time represents the zero time instant
func (c CreatedAt) IsZero() bool {
	return c.CreatedAt.IsZero()
}

// Unix returns the local time corresponding to the given Unix time
func (c CreatedAt) Unix() int64 {
	return c.CreatedAt.Unix()
}

// Format returns a textual representation of the time value formatted according to the layout
func (c CreatedAt) Format(layout string) string {
	return c.CreatedAt.Format(layout)
}

// NewCreatedAt creates a new CreatedAt with the given time
func NewCreatedAt(t time.Time) CreatedAt {
	return CreatedAt{CreatedAt: t}
}

// NowCreatedAt creates a new CreatedAt with the current time
func NowCreatedAt() CreatedAt {
	return CreatedAt{CreatedAt: time.Now()}
}

// UpdatedAt represents the timestamp when an entity was last updated.
// It uses GORM's autoUpdateTime for automatic timestamp management.
type UpdatedAt struct {
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// Time returns the underlying time.Time value
func (u UpdatedAt) Time() time.Time {
	return u.UpdatedAt
}

// Set sets the UpdatedAt timestamp
func (u *UpdatedAt) Set(t time.Time) {
	u.UpdatedAt = t
}

// After reports whether the UpdatedAt time is after t
func (u UpdatedAt) After(t time.Time) bool {
	return u.UpdatedAt.After(t)
}

// Before reports whether the UpdatedAt time is before t
func (u UpdatedAt) Before(t time.Time) bool {
	return u.UpdatedAt.Before(t)
}

// Equal reports whether the UpdatedAt time is equal to t
func (u UpdatedAt) Equal(t time.Time) bool {
	return u.UpdatedAt.Equal(t)
}

// Add returns the time t+d
func (u UpdatedAt) Add(d time.Duration) time.Time {
	return u.UpdatedAt.Add(d)
}

// Sub returns the duration t-u
func (u UpdatedAt) Sub(t time.Time) time.Duration {
	return u.UpdatedAt.Sub(t)
}

// IsZero reports whether the UpdatedAt time represents the zero time instant
func (u UpdatedAt) IsZero() bool {
	return u.UpdatedAt.IsZero()
}

// Unix returns the local time corresponding to the given Unix time
func (u UpdatedAt) Unix() int64 {
	return u.UpdatedAt.Unix()
}

// Format returns a textual representation of the time value formatted according to the layout
func (u UpdatedAt) Format(layout string) string {
	return u.UpdatedAt.Format(layout)
}

// NewUpdatedAt creates a new UpdatedAt with the given time
func NewUpdatedAt(t time.Time) UpdatedAt {
	return UpdatedAt{UpdatedAt: t}
}

// NowUpdatedAt creates a new UpdatedAt with the current time
func NowUpdatedAt() UpdatedAt {
	return UpdatedAt{UpdatedAt: time.Now()}
}

// SoftDelete provides soft delete functionality in GORM.
// It includes a `Deleted` flag and a `DeletedAt` timestamp.
//
// GORM integration: embedding SoftDelete in a model causes GORM to intercept
// db.Delete() calls and run UPDATE SET deleted=1, deleted_at=NOW() instead of
// a hard DELETE. Queries automatically filter with WHERE deleted=0. Use
// db.Unscoped() to bypass the filter and see or hard-delete records.
type SoftDelete struct {
	Deleted   bool          `gorm:"column:deleted;default:0;index" json:"deleted"`
	DeletedAt SoftDeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// IsDeleted checks if the entity has been marked as deleted.
// Returns true if the `Deleted` field is set to true.
func (o *SoftDelete) IsDeleted() bool {
	return o.Deleted
}

// SetDeleted updates the `Deleted` and `DeletedAt` fields of the SoftDelete object.
// If `v` is true, it sets `Deleted` to true and `DeletedAt` to the current time.
// If `v` is false, it resets `Deleted` to false and clears `DeletedAt`.
func (o *SoftDelete) SetDeleted(v bool) {
	o.Deleted = v
	if v {
		o.DeletedAt = SoftDeletedAt{Valid: true, Time: time.Now()}
	} else {
		o.DeletedAt = SoftDeletedAt{Valid: false}
	}
}

// Restore resets the Deleted status and clears the DeletedAt timestamp.
func (o *SoftDelete) Restore() {
	o.SetDeleted(false)
}

// Delete set the Deleted status and the DeletedAt timestamp.
func (o *SoftDelete) Delete() {
	o.SetDeleted(true)
}

// Time returns the underlying time.Time value, or zero time if DeletedAt is not set.
func (o SoftDelete) Time() time.Time {
	if !o.DeletedAt.Valid {
		return time.Time{}
	}
	return o.DeletedAt.Time
}

// Set sets the DeletedAt timestamp.
func (o *SoftDelete) Set(t time.Time) {
	o.DeletedAt = SoftDeletedAt{Valid: true, Time: t}
}

// After reports whether the DeletedAt time is after u (returns false if DeletedAt is not set).
func (o SoftDelete) After(u time.Time) bool {
	if !o.DeletedAt.Valid {
		return false
	}
	return o.DeletedAt.Time.After(u)
}

// Before reports whether the DeletedAt time is before u (returns false if DeletedAt is not set).
func (o SoftDelete) Before(u time.Time) bool {
	if !o.DeletedAt.Valid {
		return false
	}
	return o.DeletedAt.Time.Before(u)
}

// Equal reports whether the DeletedAt time is equal to u (returns false if DeletedAt is not set).
func (o SoftDelete) Equal(u time.Time) bool {
	if !o.DeletedAt.Valid {
		return false
	}
	return o.DeletedAt.Time.Equal(u)
}

// IsZero reports whether the DeletedAt is not set or represents the zero time instant.
func (o SoftDelete) IsZero() bool {
	return !o.DeletedAt.Valid || o.DeletedAt.Time.IsZero()
}

// Unix returns the Unix timestamp (returns 0 if DeletedAt is not set).
func (o SoftDelete) Unix() int64 {
	if !o.DeletedAt.Valid {
		return 0
	}
	return o.DeletedAt.Time.Unix()
}

// Format returns a textual representation of the time value formatted according to the layout
// (returns empty string if DeletedAt is not set).
func (o SoftDelete) Format(layout string) string {
	if !o.DeletedAt.Valid {
		return ""
	}
	return o.DeletedAt.Time.Format(layout)
}
