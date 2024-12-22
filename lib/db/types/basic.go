package types

import (
	"time"
)

// CreatedAt represents the timestamp of when an entity or object was created.
// It uses GORM's autoCreateTime for automatic timestamp management.
type CreatedAt struct {
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP()" json:"created_at"`
}

// UpdatedAt represents the timestamp when an entity was last updated.
// It uses GORM's autoUpdateTime for automatic timestamp management.
type UpdatedAt struct {
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP();ON_UPDATE:CURRENT_TIMESTAMP()" json:"updated_at"`
}

// SoftDelete provides soft delete functionality in GORM.
// It includes a `Deleted` flag and a `DeletedAt` timestamp.
type SoftDelete struct {
	Deleted   bool       `gorm:"column:deleted;default:0;index" json:"deleted"`
	DeletedAt *time.Time `gorm:"column:deleted_at;nullable" json:"deleted_at"`
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
		now := time.Now()
		o.DeletedAt = &now
	} else {
		o.DeletedAt = nil
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
