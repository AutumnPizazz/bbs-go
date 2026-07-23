package migrations

import (
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestQuoteIdentifierForReservedConfigKey(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:quote_identifier_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if got := quoteIdentifier(db, "key"); got != `"key"` {
		t.Fatalf("expected SQLite identifier quoting, got %q", got)
	}
}

func TestRemoveTaskNavigationItems(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"title": "Topics", "url": "/topics"},
		map[string]interface{}{"title": "Tasks", "url": "/tasks"},
		map[string]interface{}{
			"title": "More",
			"children": []interface{}{
				map[string]interface{}{"title": "Tasks", "url": "/tasks"},
				map[string]interface{}{"title": "Links", "url": "/links"},
			},
		},
		map[string]interface{}{
			"title": "Retired",
			"children": []interface{}{
				map[string]interface{}{"title": "Tasks", "url": "/tasks"},
			},
		},
	}

	got := removeTaskNavigationItems(items)
	want := []interface{}{
		map[string]interface{}{"title": "Topics", "url": "/topics"},
		map[string]interface{}{
			"title": "More",
			"children": []interface{}{
				map[string]interface{}{"title": "Links", "url": "/links"},
			},
		},
		map[string]interface{}{"title": "Retired"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected filtered navigation: %#v", got)
	}
}

func TestDropColumnAfterLegacyIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:migration_index_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}
	if err := db.Exec(`CREATE TABLE "t_user" ("id" INTEGER PRIMARY KEY, "score" INTEGER NOT NULL)`).Error; err != nil {
		t.Fatalf("create test table: %v", err)
	}
	if err := db.Exec(`CREATE INDEX "idx_user_score" ON "t_user" ("score")`).Error; err != nil {
		t.Fatalf("create legacy index: %v", err)
	}

	if err := dropLegacyIndex(db, "t_user", "idx_user_score"); err != nil {
		t.Fatalf("drop legacy index: %v", err)
	}
	if err := dropColumn(db, "t_user", "score"); err != nil {
		t.Fatalf("drop indexed column: %v", err)
	}
	if db.Migrator().HasColumn("t_user", "score") {
		t.Fatal("expected score column to be removed")
	}
}
