package migrations

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/simple/sqls"
)

func migrate_unify_article_topics() error {
	articlePermissionCodes := []string{
		"dashboard.article.view",
		"dashboard.article.update",
		"dashboard.article.audit",
		"dashboard.article.delete",
		"dashboard.article.tags",
	}

	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Exec("UPDATE t_topic SET format = 'post' WHERE format IS NULL OR format = ''").Error; err != nil {
			return err
		}
		if err := ctx.Tx.Exec("UPDATE t_topic SET update_time = create_time WHERE update_time IS NULL OR update_time = 0").Error; err != nil {
			return err
		}

		var permissionIds []int64
		if err := ctx.Tx.Model(&models.Permission{}).
			Where("code in ?", articlePermissionCodes).
			Pluck("id", &permissionIds).Error; err != nil {
			return err
		}
		if len(permissionIds) > 0 {
			if err := ctx.Tx.Where("permission_id in ?", permissionIds).Delete(&models.RolePermission{}).Error; err != nil {
				return err
			}
		}
		if err := ctx.Tx.Where("code in ?", articlePermissionCodes).Delete(&models.Permission{}).Error; err != nil {
			return err
		}

		if err := ctx.Tx.Exec("DROP TABLE IF EXISTS t_article_tag").Error; err != nil {
			return err
		}
		return ctx.Tx.Exec("DROP TABLE IF EXISTS t_article").Error
	})
}
