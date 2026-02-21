package main

import (
	"time"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db/schema"
)

// User demonstrates basic model with primary key, indexes, and constraints
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username  string    `gorm:"uniqueIndex:idx_username;size:50;not null" json:"username"`
	FirstName string    `gorm:"size:100" json:"first_name"`
	LastName  string    `gorm:"size:100" json:"last_name"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	IsActive  bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// UserProfile demonstrates foreign key relationships and advanced field types
type UserProfile struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;FK:users.id" json:"user_id"`
	Bio         string    `gorm:"type:text" json:"bio"`
	Website     string    `gorm:"size:255" json:"website"`
	Avatar      string    `gorm:"size:500" json:"avatar"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Phone       string    `gorm:"size:20" json:"phone"`
	Address     string    `gorm:"size:500" json:"address"`
	Country     string    `gorm:"size:100;default:'US'" json:"country"`
	Timezone    string    `gorm:"size:50;default:'UTC'" json:"timezone"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Post demonstrates full-text search, various indexes, and enums
type Post struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;index;FK:users.id" json:"user_id"`
	Title       string    `gorm:"size:255;not null;FULLTEXT" json:"title"`
	Content     string    `gorm:"type:text;FULLTEXT" json:"content"`
	Slug        string    `gorm:"uniqueIndex;size:255;not null" json:"slug"`
	Status      string    `gorm:"type:enum('draft','published','archived');default:'draft';not null" json:"status"`
	PublishedAt *time.Time `gorm:"index" json:"published_at"`
	ViewCount   int       `gorm:"default:0;not null" json:"view_count"`
	Tags        string    `gorm:"size:500" json:"tags"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Comment demonstrates self-referencing foreign keys and composite indexes
type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    uint      `gorm:"not null;index:idx_post_user,priority:1;FK:posts.id" json:"post_id"`
	UserID    uint      `gorm:"not null;index:idx_post_user,priority:2;FK:users.id" json:"user_id"`
	ParentID  *uint     `gorm:"index;FK:comments.id" json:"parent_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsApproved bool     `gorm:"default:false;not null;index" json:"is_approved"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Category demonstrates custom table options and charset/collation
type Category struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex;CHARSET:utf8mb4;COLLATE:utf8mb4_unicode_ci" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *uint     `gorm:"index;FK:categories.id" json:"parent_id"`
	SortOrder   int       `gorm:"default:0;not null" json:"sort_order"`
	IsActive    bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Implement custom table configuration
func (Category) TableEngine() string {
	return "InnoDB"
}

func (Category) TableCharset() string {
	return "utf8mb4"
}

func (Category) TableCollation() string {
	return "utf8mb4_unicode_ci"
}

// Tag demonstrates many-to-many relationships and custom column types
type Tag struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Color     string    `gorm:"size:7;default:'#000000'" json:"color"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// PostTag demonstrates junction table for many-to-many relationships
type PostTag struct {
	PostID uint `gorm:"primaryKey;FK:posts.id" json:"post_id"`
	TagID  uint `gorm:"primaryKey;FK:tags.id" json:"tag_id"`
}

// Settings demonstrates various data types and constraints
type Settings struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex;FK:users.id" json:"user_id"`
	Theme       string    `gorm:"size:20;default:'light'" json:"theme"`
	Language    string    `gorm:"size:5;default:'en'" json:"language"`
	EmailNotif  bool      `gorm:"default:true;not null" json:"email_notifications"`
	PushNotif   bool      `gorm:"default:true;not null" json:"push_notifications"`
	Preferences string    `gorm:"type:json" json:"preferences"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Log demonstrates versioned migrations
type Log struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    *uint     `gorm:"index;FK:users.id" json:"user_id"`
	Action    string    `gorm:"size:50;not null;index" json:"action"`
	Resource  string    `gorm:"size:50;not null;index" json:"resource"`
	Details   string    `gorm:"type:text" json:"details"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// Implement versioned migration for Log table
func (Log) Migration(currentVersion string) []schema.Migration {
	// For now, return empty migrations to avoid SQLite AFTER syntax error
	// TODO: Implement database-specific migration queries
	return []schema.Migration{}
}

// SessionData demonstrates NULLABLE fields and complex constraints
type SessionData struct {
	ID        string    `gorm:"primaryKey;size:128" json:"id"`
	UserID    *uint     `gorm:"NULLABLE;index;FK:users.id" json:"user_id"`
	Data      string    `gorm:"type:text" json:"data"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// InitializeModels registers all models for migration
func InitializeModels() {
	if db := evo.GetDBO(); db != nil {
		schema.UseModel(db,
			User{},
			UserProfile{},
			Post{},
			Comment{},
			Category{},
			Tag{},
			PostTag{},
			Settings{},
			Log{},
			SessionData{},
		)
	}
}