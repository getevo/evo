package types_test

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/getevo/evo/v2/lib/db/types"
)

type testUser struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"column:name"`
	types.SoftDelete
}

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSoftDelete_DeleteSetsFlags(t *testing.T) {
	db := openDB(t)

	db.Create(&testUser{Name: "Alice"})
	db.Create(&testUser{Name: "Bob"})

	// Soft-delete Alice
	db.Where("name = ?", "Alice").Delete(&testUser{})

	// Alice must NOT appear in normal queries
	var users []testUser
	db.Find(&users)
	if len(users) != 1 || users[0].Name != "Bob" {
		t.Fatalf("expected only Bob, got %+v", users)
	}

	// Alice must appear when unscoped
	db.Unscoped().Find(&users)
	if len(users) != 2 {
		t.Fatalf("expected 2 unscoped rows, got %d", len(users))
	}

	// Verify deleted=1 and deleted_at IS NOT NULL for Alice
	var alice testUser
	db.Unscoped().Where("name = ?", "Alice").First(&alice)
	if !alice.Deleted {
		t.Error("expected Deleted=true for Alice")
	}
	if !alice.DeletedAt.Valid {
		t.Error("expected DeletedAt.Valid=true for Alice")
	}
}

func TestSoftDelete_HardDeleteWithUnscoped(t *testing.T) {
	db := openDB(t)

	db.Create(&testUser{Name: "Carol"})
	db.Unscoped().Where("name = ?", "Carol").Delete(&testUser{})

	var users []testUser
	db.Unscoped().Find(&users)
	if len(users) != 0 {
		t.Fatalf("expected 0 rows after hard delete, got %d", len(users))
	}
}

func TestSoftDelete_QueryFiltersDeleted(t *testing.T) {
	db := openDB(t)

	db.Create(&testUser{Name: "Dave"})
	db.Create(&testUser{Name: "Eve"})
	db.Where("name = ?", "Dave").Delete(&testUser{})

	// Count should only include Eve
	var count int64
	db.Model(&testUser{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected count=1, got %d", count)
	}
}

func TestSoftDelete_SetDeletedHelper(t *testing.T) {
	var sd types.SoftDelete

	sd.SetDeleted(true)
	if !sd.Deleted {
		t.Error("expected Deleted=true after SetDeleted(true)")
	}
	if !sd.DeletedAt.Valid {
		t.Error("expected DeletedAt.Valid=true after SetDeleted(true)")
	}

	sd.SetDeleted(false)
	if sd.Deleted {
		t.Error("expected Deleted=false after SetDeleted(false)")
	}
	if sd.DeletedAt.Valid {
		t.Error("expected DeletedAt.Valid=false after SetDeleted(false)")
	}
}

func TestSoftDelete_TimeHelpers(t *testing.T) {
	var sd types.SoftDelete

	if !sd.IsZero() {
		t.Error("empty SoftDelete should IsZero()")
	}
	if sd.Unix() != 0 {
		t.Errorf("expected Unix()=0 when not set, got %d", sd.Unix())
	}
	if sd.Format("2006") != "" {
		t.Errorf("expected Format()='' when not set, got %q", sd.Format("2006"))
	}

	sd.SetDeleted(true)
	if sd.IsZero() {
		t.Error("set SoftDelete should not IsZero()")
	}
	if sd.Unix() == 0 {
		t.Error("expected non-zero Unix() after SetDeleted(true)")
	}
	year := sd.Format("2006")
	if !strings.HasPrefix(year, "20") {
		t.Errorf("unexpected year format: %q", year)
	}
}
