package model

import "time"

// CreatedAt represents the timestamp of when an entity or object was created.
// It is used to track the creation time of various entities.
type CreatedAt struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// UpdatedAt represents the timestamp when an entity was last updated.
// It is used to keep track of the latest modification of the entity.
type UpdatedAt struct {
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// DeletedAt struct represents the soft delete functionality in GORM.
// Embedding this struct causes GORM to intercept db.Delete() calls and run
// UPDATE SET deleted=1, deleted_at=NOW() instead of a hard DELETE.
// Queries automatically filter with WHERE deleted=0.
// Use db.Unscoped() to bypass the filter.
//
// Prefer types.SoftDelete from lib/db/types for new models.
type DeletedAt struct {
	Deleted   bool       `gorm:"column:deleted;index:deleted" json:"deleted"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

// IsDeleted returns true if the Deleted field of the DeletedAt object is set to true, indicating that the object has been deleted. Otherwise, it returns false.
func (o *DeletedAt) IsDeleted() bool {
	return o.Deleted
}

// Delete updates the `Deleted` and `DeletedAt` fields of the `DeletedAt` object.
// If `v` is true, `Deleted` is set to `v` and `DeletedAt` is set to the current time.
// If `v` is false, `Deleted` is set to `v` and `DeletedAt` is set to `nil`.
func (o *DeletedAt) Delete(v bool) {
	o.Deleted = v
	if v {
		var now = time.Now()
		o.DeletedAt = &now
	} else {
		o.DeletedAt = nil
	}
}
