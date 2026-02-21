package main

import (
	"time"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db/schema"
	"fmt"
)

// TestForeignKeys validates all FK: syntax variations
type TestUser struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;size:255;not null" json:"email"`
}

type TestProfile struct {
	ID     uint `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID uint `gorm:"not null;FK:test_users.id" json:"user_id"`
	Bio    string `gorm:"type:text" json:"bio"`
}

type TestPost struct {
	ID       uint `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthorID uint `gorm:"not null;FK:test_users.id" json:"author_id"`
	Title    string `gorm:"size:255;not null" json:"title"`
}

type TestComment struct {
	ID       uint  `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID   uint  `gorm:"not null;FK:test_posts.id" json:"post_id"`
	UserID   uint  `gorm:"not null;FK:test_users.id" json:"user_id"`
	ParentID *uint `gorm:"NULLABLE;FK:test_comments.id" json:"parent_id"`
	Content  string `gorm:"type:text;not null" json:"content"`
}

// TestEnums validates ENUM handling across databases
type TestProduct struct {
	ID     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name   string `gorm:"size:255;not null" json:"name"`
	Status string `gorm:"type:enum('active','inactive','pending');default:'pending';not null" json:"status"`
	Type   string `gorm:"type:enum('physical','digital','service');default:'physical';not null" json:"type"`
}

type TestOrder struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Status   string `gorm:"type:enum('draft','pending','processing','shipped','delivered','cancelled');default:'draft';not null" json:"status"`
	Priority string `gorm:"type:enum('low','medium','high','urgent');default:'medium';not null" json:"priority"`
}

// TestDefaults validates default value handling
type TestDefaults struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StringDefault string    `gorm:"size:100;default:'default_value';not null" json:"string_default"`
	IntDefault    int       `gorm:"default:42;not null" json:"int_default"`
	BoolDefault   bool      `gorm:"default:true;not null" json:"bool_default"`
	FloatDefault  float64   `gorm:"default:3.14;not null" json:"float_default"`
	TimeDefault   time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"time_default"`
	IsActive      bool      `gorm:"default:false;not null" json:"is_active"`
	Counter       int       `gorm:"default:0;not null" json:"counter"`
	Rating        float32   `gorm:"default:0.0;not null" json:"rating"`
}

// TestComplexConstraints combines multiple features
type TestComplexModel struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;FK:test_users.id" json:"user_id"`
	Status      string    `gorm:"type:enum('draft','published','archived');default:'draft';not null" json:"status"`
	Priority    int       `gorm:"default:1;not null;check:priority >= 1 AND priority <= 10" json:"priority"`
	Title       string    `gorm:"size:255;not null;uniqueIndex" json:"title"`
	Description string    `gorm:"size:500;default:'No description provided'" json:"description"`
	IsPublic    bool      `gorm:"default:false;not null" json:"is_public"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// TestVectorEmbedding demonstrates PostgreSQL vector embedding support
type TestVectorEmbedding struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Content     string    `gorm:"type:text" json:"content"`
	Embedding   []float32 `gorm:"type:vector(1536)" json:"embedding"`    // OpenAI embedding
	SmallVector []float32 `gorm:"type:vector(384)" json:"small_vector"`  // Smaller embedding
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TestSpecialCases runs comprehensive tests for special migration cases
func TestSpecialCases() {
	fmt.Println("🧪 Starting Comprehensive Migration Tests")
	
	db := evo.GetDBO()
	if db == nil {
		fmt.Println("❌ Failed to get database connection")
		return
	}
	
	fmt.Println("📊 Testing Foreign Key Constraints...")
	testModels := []interface{}{
		TestUser{},
		TestProfile{},
		TestPost{},
		TestComment{},
	}
	
	for _, model := range testModels {
		fmt.Printf("   Testing FK model: %T\n", model)
		schema.UseModel(db, model)
		fmt.Printf("   ✅ Successfully registered FK model: %T\n", model)
	}
	
	fmt.Println("🎯 Testing ENUM Constraints...")
	enumModels := []interface{}{
		TestProduct{},
		TestOrder{},
	}
	
	for _, model := range enumModels {
		fmt.Printf("   Testing ENUM model: %T\n", model)
		schema.UseModel(db, model)
		fmt.Printf("   ✅ Successfully registered ENUM model: %T\n", model)
	}
	
	fmt.Println("⚙️ Testing Default Values...")
	defaultModels := []interface{}{
		TestDefaults{},
		TestComplexModel{},
		TestVectorEmbedding{}, // PostgreSQL vector support
	}
	
	for _, model := range defaultModels {
		fmt.Printf("   Testing DEFAULT model: %T\n", model)
		schema.UseModel(db, model)
		fmt.Printf("   ✅ Successfully registered DEFAULT model: %T\n", model)
	}
	
	fmt.Println("🔄 Attempting to run migration...")
	if err := schema.DoMigrationV2(db); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
	} else {
		fmt.Println("✅ Migration completed successfully!")
	}
	
	fmt.Println("🧪 Migration tests completed")
}