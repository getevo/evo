package main

import (
	"fmt"
	"log"
	"time"
	
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db/schema"
	"gorm.io/gorm"
)

// TestMigration runs comprehensive migration tests
func TestMigration() {
	fmt.Println("🚀 Starting Migration Test Suite")
	fmt.Println("================================")
	
	// Get database info
	db := evo.GetDBO()
	if db == nil {
		log.Fatal("❌ Database connection not available")
	}
	
	// Test database detection
	fmt.Println("📊 Testing Database Detection...")
	engine, err := schema.NewMigrationEngine(db)
	if err != nil {
		log.Fatal("❌ Failed to create migration engine:", err)
	}
	
	fmt.Printf("✅ Database detected: %s\n", engine.GetDatabaseType())
	fmt.Printf("📋 Database info: %s\n", engine.GetDatabaseInfo())
	
	// Run migration
	fmt.Println("\n📦 Running Migration...")
	startTime := time.Now()
	
	err = schema.DoMigrationV2(db)
	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}
	
	duration := time.Since(startTime)
	fmt.Printf("✅ Migration completed in %v\n", duration)
	
	// Verify tables were created
	fmt.Println("\n🔍 Verifying Table Creation...")
	if err := verifyTables(db); err != nil {
		log.Fatal("❌ Table verification failed:", err)
	}
	
	// Test data insertion
	fmt.Println("\n💾 Testing Data Insertion...")
	if err := testDataInsertion(db); err != nil {
		log.Fatal("❌ Data insertion test failed:", err)
	}
	
	// Test foreign key constraints
	fmt.Println("\n🔗 Testing Foreign Key Constraints...")
	if err := testForeignKeys(db); err != nil {
		log.Fatal("❌ Foreign key test failed:", err)
	}
	
	// Test indexes
	fmt.Println("\n📈 Testing Indexes...")
	if err := testIndexes(db); err != nil {
		log.Fatal("❌ Index test failed:", err)
	}
	
	fmt.Println("\n🎉 All tests passed successfully!")
	fmt.Println("Migration system is working correctly.")
}

// verifyTables checks that all expected tables exist
func verifyTables(db *gorm.DB) error {
	expectedTables := []string{
		"users",
		"user_profiles", 
		"posts",
		"comments",
		"categories",
		"tags",
		"post_tags",
		"settings",
		"logs",
		"session_data",
	}
	
	// Detect database type to use appropriate query
	dialectName := db.Dialector.Name()
	
	for _, tableName := range expectedTables {
		var count int64
		var err error
		
		switch dialectName {
		case "sqlite":
			// SQLite uses sqlite_master table
			err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", tableName).Scan(&count).Error
		case "postgres":
			// PostgreSQL uses information_schema.tables
			err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tableName).Scan(&count).Error
		default: // MySQL, MariaDB
			err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tableName).Scan(&count).Error
		}
		
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", tableName, err)
		}
		
		if count == 0 {
			return fmt.Errorf("table %s was not created", tableName)
		}
		
		fmt.Printf("  ✅ Table '%s' exists\n", tableName)
	}
	
	return nil
}

// testDataInsertion tests basic CRUD operations
func testDataInsertion(db *gorm.DB) error {
	// Use unique email/username for each test run
	timestamp := time.Now().Unix()
	
	// Test User creation
	user := User{
		Email:     fmt.Sprintf("test%d@example.com", timestamp),
		Username:  fmt.Sprintf("testuser%d", timestamp),
		FirstName: "Test",
		LastName:  "User",
		Password:  "hashed_password",
		IsActive:  true,
	}
	
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Printf("  ✅ User created with ID: %d\n", user.ID)
	
	// Test UserProfile creation with foreign key
	profile := UserProfile{
		UserID:  user.ID,
		Bio:     "Test user bio",
		Website: "https://example.com",
		Country: "US",
	}
	
	if err := db.Create(&profile).Error; err != nil {
		return fmt.Errorf("failed to create user profile: %w", err)
	}
	fmt.Printf("  ✅ User profile created with ID: %d\n", profile.ID)
	
	// Test Category creation
	category := Category{
		Name:        fmt.Sprintf("Technology%d", timestamp),
		Description: "Tech related posts",
		IsActive:    true,
	}
	
	if err := db.Create(&category).Error; err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	fmt.Printf("  ✅ Category created with ID: %d\n", category.ID)
	
	// Test Post creation
	now := time.Now()
	post := Post{
		UserID:      user.ID,
		Title:       fmt.Sprintf("Test Post %d", timestamp),
		Content:     "This is a test post content",
		Slug:        fmt.Sprintf("test-post-%d", timestamp),
		Status:      "published",
		PublishedAt: &now,
		ViewCount:   0,
	}
	
	if err := db.Create(&post).Error; err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	fmt.Printf("  ✅ Post created with ID: %d\n", post.ID)
	
	// Test Comment creation with foreign keys
	comment := Comment{
		PostID:     post.ID,
		UserID:     user.ID,
		Content:    "Great post!",
		IsApproved: true,
	}
	
	if err := db.Create(&comment).Error; err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	fmt.Printf("  ✅ Comment created with ID: %d\n", comment.ID)
	
	return nil
}

// testForeignKeys tests foreign key constraint enforcement
func testForeignKeys(db *gorm.DB) error {
	// Try to create a profile with non-existent user ID
	invalidProfile := UserProfile{
		UserID:  99999, // Non-existent user ID
		Bio:     "Invalid profile",
		Country: "US",
	}
	
	err := db.Create(&invalidProfile).Error
	if err == nil {
		return fmt.Errorf("expected foreign key constraint error but got none")
	}
	
	fmt.Printf("  ✅ Foreign key constraint properly enforced: %v\n", err)
	
	// Test valid foreign key
	var user User
	if err := db.First(&user).Error; err != nil {
		return fmt.Errorf("failed to find user for FK test: %w", err)
	}
	
	validProfile := UserProfile{
		UserID:  user.ID,
		Bio:     "Valid profile",
		Country: "US",
	}
	
	if err := db.Create(&validProfile).Error; err != nil {
		return fmt.Errorf("failed to create valid profile: %w", err)
	}
	
	fmt.Printf("  ✅ Valid foreign key accepted\n")
	return nil
}

// testIndexes tests that indexes were created correctly
func testIndexes(db *gorm.DB) error {
	timestamp := time.Now().Unix()
	uniqueEmail := fmt.Sprintf("unique%d@example.com", timestamp)
	
	// Test unique constraint on email
	user1 := User{
		Email:    uniqueEmail,
		Username: fmt.Sprintf("user%d_1", timestamp),
		Password: "password",
	}
	
	if err := db.Create(&user1).Error; err != nil {
		return fmt.Errorf("failed to create first user: %w", err)
	}
	
	// Try to create another user with same email
	user2 := User{
		Email:    uniqueEmail, // Same email
		Username: fmt.Sprintf("user%d_2", timestamp),
		Password: "password",
	}
	
	err := db.Create(&user2).Error
	if err == nil {
		return fmt.Errorf("expected unique constraint error but got none")
	}
	
	fmt.Printf("  ✅ Unique index properly enforced on email\n")
	
	// Test unique constraint on username
	user3 := User{
		Email:    fmt.Sprintf("another%d@example.com", timestamp),
		Username: fmt.Sprintf("user%d_1", timestamp), // Same username as user1
		Password: "password",
	}
	
	err = db.Create(&user3).Error
	if err == nil {
		return fmt.Errorf("expected unique constraint error for username but got none")
	}
	
	fmt.Printf("  ✅ Unique index properly enforced on username\n")
	
	return nil
}

