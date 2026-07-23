package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

const siteNavsConfigKey = constants.SysConfigSiteNavs

// migrate_remove_task_growth_and_points removes the retired feature data and
// makes existing installations match the reduced model set.
func migrate_remove_task_growth_and_points() error {
	db := modelsDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	for _, table := range []string{
		"t_task_config",
		"t_user_task_event",
		"t_user_task_log",
		"t_badge",
		"t_user_badge",
		"t_level_config",
		"t_user_exp_log",
		"t_check_in",
		"t_user_score_log",
		"t_attachment_download_log",
	} {
		if err := dropTable(db, table); err != nil {
			return err
		}
	}

	if err := dropLegacyIndex(db, "t_user", "idx_user_score"); err != nil {
		return err
	}

	for _, item := range []struct {
		table  string
		column string
	}{
		{table: "t_user", column: "score"},
		{table: "t_user", column: "exp"},
		{table: "t_user", column: "level"},
		{table: "t_topic", column: "bounty_score"},
		{table: "t_attachment", column: "download_score"},
	} {
		if err := dropColumn(db, item.table, item.column); err != nil {
			return err
		}
	}

	if err := removeRetiredConfig(db); err != nil {
		return err
	}
	if err := removeTaskNavigation(db); err != nil {
		return err
	}
	if err := removeRetiredPermissions(db); err != nil {
		return err
	}
	if err := removeRetiredMessages(db); err != nil {
		return err
	}
	return nil
}

func modelsDB() *gorm.DB {
	return sqls.DB()
}

func dropTable(db *gorm.DB, table string) error {
	return db.Exec("DROP TABLE IF EXISTS " + quoteIdentifier(db, table)).Error
}

func dropColumn(db *gorm.DB, table, column string) error {
	if !db.Migrator().HasColumn(table, column) {
		return nil
	}
	return db.Exec("ALTER TABLE " + quoteIdentifier(db, table) + " DROP COLUMN " + quoteIdentifier(db, column)).Error
}

func dropLegacyIndex(db *gorm.DB, table, index string) error {
	if !db.Migrator().HasIndex(table, index) {
		return nil
	}
	return db.Migrator().DropIndex(table, index)
}

func quoteIdentifier(db *gorm.DB, value string) string {
	if db.Dialector.Name() == "mysql" {
		return "`" + strings.ReplaceAll(value, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func removeRetiredConfig(db *gorm.DB) error {
	return db.Exec("DELETE FROM "+quoteIdentifier(db, "t_sys_config")+" WHERE "+quoteIdentifier(db, "key")+" IN (?, ?, ?, ?)",
		"enableQaBounty", "qaBountyMin", "qaBountyMax", "qaBountyRequired").Error
}

func removeTaskNavigation(db *gorm.DB) error {
	var config models.SysConfig
	if err := db.Where(quoteIdentifier(db, "key")+" = ?", siteNavsConfigKey).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	var navs []interface{}
	if err := json.Unmarshal([]byte(config.Value), &navs); err != nil {
		return nil
	}
	filtered := removeTaskNavigationItems(navs)
	data, err := json.Marshal(filtered)
	if err != nil {
		return err
	}
	return db.Model(&models.SysConfig{}).Where(quoteIdentifier(db, "key")+" = ?", siteNavsConfigKey).Updates(map[string]interface{}{
		"value": string(data),
	}).Error
}

func removeTaskNavigationItems(items []interface{}) []interface{} {
	filtered := make([]interface{}, 0, len(items))
	for _, item := range items {
		nav, ok := item.(map[string]interface{})
		if !ok {
			filtered = append(filtered, item)
			continue
		}
		if url, ok := nav["url"].(string); ok && url == "/tasks" {
			continue
		}
		if children, ok := nav["children"].([]interface{}); ok {
			children = removeTaskNavigationItems(children)
			if len(children) == 0 {
				delete(nav, "children")
			} else {
				nav["children"] = children
			}
		}
		filtered = append(filtered, nav)
	}
	return filtered
}

func removeRetiredPermissions(db *gorm.DB) error {
	const codes = "('dashboard.badge.view','dashboard.badge.create','dashboard.badge.update','dashboard.badge.delete','dashboard.level.view','dashboard.level.update','dashboard.task.view','dashboard.task.create','dashboard.task.update','dashboard.task.delete','dashboard.userBadge.view','dashboard.userExpLog.view','dashboard.userTaskLog.view')"
	if err := db.Exec("DELETE FROM t_role_permission WHERE permission_id IN (SELECT id FROM t_permission WHERE code IN " + codes + ")").Error; err != nil {
		return err
	}
	return db.Exec("DELETE FROM t_permission WHERE code IN " + codes).Error
}

func removeRetiredMessages(db *gorm.DB) error {
	return db.Exec("DELETE FROM t_message WHERE type IN (7, 8)").Error
}
