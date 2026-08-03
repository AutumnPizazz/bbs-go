package repositories

import (
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

// BaseRepository 提供标准 CRUD 样板实现，各业务仓库通过泛型嵌入复用。
type BaseRepository[T any] struct{}

func (r *BaseRepository[T]) Get(db *gorm.DB, id any) *T {
	ret := new(T)
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *BaseRepository[T]) Take(db *gorm.DB, where ...any) *T {
	ret := new(T)
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *BaseRepository[T]) Find(db *gorm.DB, cnd *sqls.Cnd) []T {
	var list []T
	cnd.Find(db, &list)
	return list
}

func (r *BaseRepository[T]) FindOne(db *gorm.DB, cnd *sqls.Cnd) *T {
	ret := new(T)
	if err := cnd.FindOne(db, ret); err != nil {
		return nil
	}
	return ret
}

func (r *BaseRepository[T]) FindPageByParams(db *gorm.DB, p *params.QueryParams) ([]T, *sqls.Paging) {
	return r.FindPageByCnd(db, &p.Cnd)
}

func (r *BaseRepository[T]) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) ([]T, *sqls.Paging) {
	var list []T
	cnd.Find(db, &list)
	count := cnd.Count(db, new(T))
	return list, &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
}

func (r *BaseRepository[T]) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, new(T))
}

func (r *BaseRepository[T]) Create(db *gorm.DB, t *T) error {
	return db.Create(t).Error
}

func (r *BaseRepository[T]) Update(db *gorm.DB, t *T) error {
	return db.Save(t).Error
}

func (r *BaseRepository[T]) Updates(db *gorm.DB, id any, columns map[string]interface{}) error {
	return db.Model(new(T)).Where("id = ?", id).Updates(columns).Error
}

func (r *BaseRepository[T]) UpdateColumn(db *gorm.DB, id any, name string, value interface{}) error {
	return db.Model(new(T)).Where("id = ?", id).UpdateColumn(name, value).Error
}

func (r *BaseRepository[T]) Delete(db *gorm.DB, id any) {
	db.Delete(new(T), "id = ?", id)
}

func (r *BaseRepository[T]) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) []T {
	var list []T
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return list
}

func (r *BaseRepository[T]) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) int64 {
	var count int64
	db.Raw(sqlStr, paramArr...).Count(&count)
	return count
}
