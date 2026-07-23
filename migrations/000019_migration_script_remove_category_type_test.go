package migrations

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func TestMigrateRemoveCategoryType(t *testing.T) {
	dsn := fmt.Sprintf("file:remove_category_type_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	sqls.SetDB(db)

	if err := db.Exec(`CREATE TABLE t_category (id INTEGER PRIMARY KEY, name TEXT NOT NULL, type TEXT NOT NULL DEFAULT 'normal')`).Error; err != nil {
		t.Fatalf("create legacy category table: %v", err)
	}
	if err := db.Exec(`CREATE INDEX idx_category_type ON t_category (type)`).Error; err != nil {
		t.Fatalf("create legacy category index: %v", err)
	}
	if err := db.Exec(`INSERT INTO t_category (id, name, type) VALUES (1, 'General', 'normal'), (2, 'Questions', 'qa')`).Error; err != nil {
		t.Fatalf("seed legacy categories: %v", err)
	}

	if err := migrate_remove_category_type(); err != nil {
		t.Fatalf("remove category type: %v", err)
	}
	if db.Migrator().HasColumn("t_category", "type") {
		t.Fatal("expected legacy category type column to be removed")
	}
	if db.Migrator().HasIndex("t_category", "idx_category_type") {
		t.Fatal("expected legacy category type index to be removed")
	}

	var count int64
	if err := db.Table("t_category").Count(&count).Error; err != nil {
		t.Fatalf("count migrated categories: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected both categories to remain, got %d", count)
	}
}
