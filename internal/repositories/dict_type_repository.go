package repositories

import (
	"bbs-go/internal/models"
)

var DictTypeRepository = newDictTypeRepository()

func newDictTypeRepository() *dictTypeRepository {
	return &dictTypeRepository{}
}

type dictTypeRepository struct {
	BaseRepository[models.DictType]
}
