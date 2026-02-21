package main

import (
	"fmt"
	"time"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db/schema"
)

// TestChangeModel1 - Initial version with basic fields
type TestChangeModel1 struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:255;not null" json:"email"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TestChangeModel2 - Enhanced version to test schema changes
type TestChangeModel2 struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:150;not null" json:"name"`        // Changed size from 100 to 150
	Email       string    `gorm:"size:255;not null;uniqueIndex" json:"email"` // Added unique index
	Phone       string    `gorm:"size:20" json:"phone"`                 // New column
	IsActive    bool      `gorm:"default:true;not null" json:"is_active"` // New column
	CategoryID  *uint     `gorm:"FK:test_categories.id" json:"category_id"` // New FK with explicit column
	StatusID    *uint     `gorm:"FK:test_statuses" json:"status_id"`    // New FK using primary key
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"` // Added index
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`     // New column
}

// TestChangeModel3 - Further changes to test removals and modifications
type TestChangeModel3 struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FullName    string    `gorm:"size:200;not null" json:"full_name"`   // Renamed from Name, changed size
	Email       string    `gorm:"size:255;not null" json:"email"`       // Removed unique index
	Phone       string    `gorm:"size:25" json:"phone"`                 // Changed size from 20 to 25
	IsActive    bool      `gorm:"default:false;not null" json:"is_active"` // Changed default
	CategoryID  *uint     `gorm:"FK:test_categories.id;index" json:"category_id"` // Added index to FK
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`     // Removed index
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	// Removed StatusID FK
}

// Helper tables for FK testing
type TestCategory struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"size:50;not null" json:"name"`
}

type TestStatus struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"size:20;not null" json:"code"`
	Name string `gorm:"size:50;not null" json:"name"`
}

// TestSchemaChanges tests various schema change scenarios
func TestSchemaChanges() {
	fmt.Println("🔄 Testing Schema Changes")
	fmt.Println("========================")
	
	db := evo.GetDBO()
	if db == nil {
		fmt.Println("❌ Failed to get database connection")
		return
	}
	
	// Register helper tables first
	schema.UseModel(db, TestCategory{}, TestStatus{})
	
	fmt.Println("📊 Phase 1: Initial Schema (TestChangeModel1)")
	schema.UseModel(db, TestChangeModel1{})
	
	if err := schema.DoMigrationV2(db); err != nil {
		fmt.Printf("❌ Phase 1 migration failed: %v\n", err)
		return
	}
	fmt.Println("✅ Phase 1 completed - Basic table created")
	
	// Clear models and register new version
	schema.ClearModels()
	schema.UseModel(db, TestCategory{}, TestStatus{})
	
	fmt.Println("\n📊 Phase 2: Schema Evolution (TestChangeModel2)")
	schema.UseModel(db, TestChangeModel2{})
	
	if err := schema.DoMigrationV2(db); err != nil {
		fmt.Printf("❌ Phase 2 migration failed: %v\n", err)
		return
	}
	fmt.Println("✅ Phase 2 completed - Added columns, indexes, and FKs")
	
	// Clear models and register final version
	schema.ClearModels()
	schema.UseModel(db, TestCategory{}, TestStatus{})
	
	fmt.Println("\n📊 Phase 3: Further Changes (TestChangeModel3)")
	schema.UseModel(db, TestChangeModel3{})
	
	if err := schema.DoMigrationV2(db); err != nil {
		fmt.Printf("❌ Phase 3 migration failed: %v\n", err)
		return
	}
	fmt.Println("✅ Phase 3 completed - Modified columns, removed indexes/FKs")
	
	fmt.Println("\n✅ Schema change testing completed!")
}

// TestForeignKeyFormats tests both FK formats
func TestForeignKeyFormats() {
	fmt.Println("\n🔗 Testing Foreign Key Formats")
	fmt.Println("==============================")
	
	// TestExplicitFK uses FK:table.column format
	type TestExplicitFK struct {
		ID     uint `gorm:"primaryKey;autoIncrement" json:"id"`
		UserID uint `gorm:"not null;FK:users.id" json:"user_id"`
		Name   string `gorm:"size:100" json:"name"`
	}
	
	// TestImplicitFK uses FK:table format (should use primary key)
	type TestImplicitFK struct {
		ID         uint `gorm:"primaryKey;autoIncrement" json:"id"`
		CategoryID uint `gorm:"not null;FK:categories" json:"category_id"`
		Title      string `gorm:"size:100" json:"title"`
	}
	
	db := evo.GetDBO()
	if db == nil {
		fmt.Println("❌ Failed to get database connection")
		return
	}
	
	fmt.Println("   Testing FK:table.column format...")
	schema.UseModel(db, TestExplicitFK{})
	
	fmt.Println("   Testing FK:table format...")
	schema.UseModel(db, TestImplicitFK{})
	
	if err := schema.DoMigrationV2(db); err != nil {
		fmt.Printf("❌ FK format test failed: %v\n", err)
		return
	}
	
	fmt.Println("✅ Both FK formats working correctly")
}