package repositories

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/locales"
	"errors"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"gorm.io/gorm"

	"bbs-go/internal/models"
)

var TagRepository = newTagRepository()

func newTagRepository() *tagRepository {
	return &tagRepository{}
}

type tagRepository struct {
	BaseRepository[models.Tag]
}

func (r *tagRepository) GetTagInIds(db *gorm.DB, tagIds []int64) []models.Tag {
	if len(tagIds) == 0 {
		return nil
	}
	var tags []models.Tag
	db.Where("id in (?)", tagIds).Find(&tags)
	return tags
}

func (r *tagRepository) GetByName(db *gorm.DB, name string) *models.Tag {
	if len(name) == 0 {
		return nil
	}
	return r.Take(db, "name = ?", name)
}

func (r *tagRepository) GetOrCreate(db *gorm.DB, name string) (*models.Tag, error) {
	if len(name) == 0 {
		return nil, errors.New(locales.Get("admin.tag_empty"))
	}
	// IMPORTANT: use the provided transaction `db` to avoid opening a second
	// connection when the caller is already inside a transaction (SQLite write
	// locks are connection scoped and can cause a deadlock/hang otherwise).
	tag := r.GetByName(db, name)
	if tag != nil {
		return tag, nil
	}

	tag = &models.Tag{
		Name:       name,
		Status:     constants.StatusOk,
		CreateTime: dates.NowTimestamp(),
		UpdateTime: dates.NowTimestamp(),
	}
	if err := r.Create(db, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *tagRepository) GetOrCreates(db *gorm.DB, tags []string) (tagIds []int64, err error) {
	for _, tagName := range tags {
		var tag *models.Tag
		tag, err = r.GetOrCreate(db, strings.TrimSpace(tagName))
		if err != nil {
			return
		}
		tagIds = append(tagIds, tag.Id)
	}
	return
}
