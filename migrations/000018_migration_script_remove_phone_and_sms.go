package migrations

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type sqliteTableColumn struct {
	Name         string         `gorm:"column:name"`
	Type         string         `gorm:"column:type"`
	NotNull      int            `gorm:"column:notnull"`
	DefaultValue sql.NullString `gorm:"column:dflt_value"`
	PrimaryKey   int            `gorm:"column:pk"`
}

type sqliteIndex struct {
	Name   string `gorm:"column:name"`
	Unique int    `gorm:"column:unique"`
}

type sqliteIndexColumn struct {
	Name string `gorm:"column:name"`
}

func migrate_remove_phone_and_sms() error {
	db := modelsDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := dropTable(tx, "t_sms_code"); err != nil {
			return err
		}
		if !tx.Migrator().HasColumn("t_user", "phone") {
			return nil
		}

		switch tx.Dialector.Name() {
		case "sqlite":
			return dropSQLiteUserPhone(tx)
		case "mysql":
			if tx.Migrator().HasIndex("t_user", "uni_t_user_phone") {
				if err := tx.Migrator().DropIndex("t_user", "uni_t_user_phone"); err != nil {
					return err
				}
			}
		case "postgres":
			if err := tx.Exec("ALTER TABLE " + quoteIdentifier(tx, "t_user") + " DROP CONSTRAINT IF EXISTS " + quoteIdentifier(tx, "uni_t_user_phone")).Error; err != nil {
				return err
			}
		}

		return dropColumn(tx, "t_user", "phone")
	})
}

func dropSQLiteUserPhone(db *gorm.DB) error {
	var columns []sqliteTableColumn
	if err := db.Raw(`PRAGMA table_info("t_user")`).Scan(&columns).Error; err != nil {
		return err
	}

	var indexes []sqliteIndex
	if err := db.Raw(`PRAGMA index_list("t_user")`).Scan(&indexes).Error; err != nil {
		return err
	}

	columnDefinitions := make([]string, 0, len(columns))
	columnNames := make([]string, 0, len(columns))
	for _, column := range columns {
		if strings.EqualFold(column.Name, "phone") {
			continue
		}

		definition := quoteIdentifier(db, column.Name)
		if column.Type != "" {
			definition += " " + column.Type
		}
		if column.PrimaryKey > 0 {
			definition += " PRIMARY KEY"
			if strings.EqualFold(column.Name, "id") && strings.EqualFold(column.Type, "INTEGER") {
				definition += " AUTOINCREMENT"
			}
		}
		if column.NotNull != 0 {
			definition += " NOT NULL"
		}
		if column.DefaultValue.Valid {
			definition += " DEFAULT " + column.DefaultValue.String
		}

		columnDefinitions = append(columnDefinitions, definition)
		columnNames = append(columnNames, quoteIdentifier(db, column.Name))
	}

	uniqueDefinitions := make([]string, 0)
	nonUniqueIndexes := make([]struct {
		name    string
		columns []string
	}, 0)
	for indexNumber, index := range indexes {
		var indexColumns []sqliteIndexColumn
		indexSQL := `PRAGMA index_info(` + quoteIdentifier(db, index.Name) + `)`
		if err := db.Raw(indexSQL).Scan(&indexColumns).Error; err != nil {
			return err
		}

		columnsForIndex := make([]string, 0, len(indexColumns))
		containsPhone := false
		for _, indexColumn := range indexColumns {
			if strings.EqualFold(indexColumn.Name, "phone") {
				containsPhone = true
			}
			columnsForIndex = append(columnsForIndex, quoteIdentifier(db, indexColumn.Name))
		}
		if containsPhone || len(columnsForIndex) == 0 {
			continue
		}

		if index.Unique != 0 {
			uniqueDefinitions = append(uniqueDefinitions, fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)",
				quoteIdentifier(db, fmt.Sprintf("uniq_t_user_%d", indexNumber)),
				strings.Join(columnsForIndex, ",")))
			continue
		}
		nonUniqueIndexes = append(nonUniqueIndexes, struct {
			name    string
			columns []string
		}{name: index.Name, columns: columnsForIndex})
	}

	definitions := append(columnDefinitions, uniqueDefinitions...)
	temporaryTable := quoteIdentifier(db, "t_user_without_phone")
	if err := db.Exec("DROP TABLE IF EXISTS " + temporaryTable).Error; err != nil {
		return err
	}
	createSQL := "CREATE TABLE " + temporaryTable + " (" + strings.Join(definitions, ",") + ")"
	if err := db.Exec(createSQL).Error; err != nil {
		return err
	}

	columnList := strings.Join(columnNames, ",")
	if err := db.Exec("INSERT INTO " + temporaryTable + " (" + columnList + ") SELECT " + columnList + " FROM " + quoteIdentifier(db, "t_user")).Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE " + quoteIdentifier(db, "t_user")).Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE " + temporaryTable + " RENAME TO " + quoteIdentifier(db, "t_user")).Error; err != nil {
		return err
	}

	for _, index := range nonUniqueIndexes {
		if err := db.Exec("CREATE INDEX " + quoteIdentifier(db, index.name) + " ON " + quoteIdentifier(db, "t_user") + " (" + strings.Join(index.columns, ",") + ")").Error; err != nil {
			return err
		}
	}
	return nil
}
