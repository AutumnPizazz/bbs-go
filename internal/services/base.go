package services

import (
	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

// crudRepository 标准 CRUD 仓库接口，由 repositories.BaseRepository[T] 泛型实现满足。
type crudRepository[T any] interface {
	Get(db *gorm.DB, id any) *T
	Take(db *gorm.DB, where ...any) *T
	Find(db *gorm.DB, cnd *sqls.Cnd) []T
	FindOne(db *gorm.DB, cnd *sqls.Cnd) *T
	FindPageByParams(db *gorm.DB, p *params.QueryParams) ([]T, *sqls.Paging)
	FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) ([]T, *sqls.Paging)
	Count(db *gorm.DB, cnd *sqls.Cnd) int64
	Create(db *gorm.DB, t *T) error
	Update(db *gorm.DB, t *T) error
	Updates(db *gorm.DB, id any, columns map[string]interface{}) error
	UpdateColumn(db *gorm.DB, id any, name string, value interface{}) error
}

// BaseService 提供标准 CRUD 透传样板，各业务服务通过泛型嵌入复用。
type BaseService[T any, R crudRepository[T]] struct {
	repo R
}

func newBaseService[T any, R crudRepository[T]](repo R) BaseService[T, R] {
	return BaseService[T, R]{repo: repo}
}

func (s BaseService[T, R]) Get(id any) *T {
	return s.repo.Get(sqls.DB(), id)
}

func (s BaseService[T, R]) Take(where ...any) *T {
	return s.repo.Take(sqls.DB(), where...)
}

func (s BaseService[T, R]) Find(cnd *sqls.Cnd) []T {
	return s.repo.Find(sqls.DB(), cnd)
}

func (s BaseService[T, R]) FindOne(cnd *sqls.Cnd) *T {
	return s.repo.FindOne(sqls.DB(), cnd)
}

func (s BaseService[T, R]) FindPageByParams(p *params.QueryParams) ([]T, *sqls.Paging) {
	return s.repo.FindPageByParams(sqls.DB(), p)
}

func (s BaseService[T, R]) FindPageByCnd(cnd *sqls.Cnd) ([]T, *sqls.Paging) {
	return s.repo.FindPageByCnd(sqls.DB(), cnd)
}

func (s BaseService[T, R]) Count(cnd *sqls.Cnd) int64 {
	return s.repo.Count(sqls.DB(), cnd)
}

func (s BaseService[T, R]) Create(t *T) error {
	return s.repo.Create(sqls.DB(), t)
}

func (s BaseService[T, R]) Update(t *T) error {
	return s.repo.Update(sqls.DB(), t)
}

func (s BaseService[T, R]) Updates(id any, columns map[string]interface{}) error {
	return s.repo.Updates(sqls.DB(), id, columns)
}

func (s BaseService[T, R]) UpdateColumn(id any, name string, value interface{}) error {
	return s.repo.UpdateColumn(sqls.DB(), id, name, value)
}
