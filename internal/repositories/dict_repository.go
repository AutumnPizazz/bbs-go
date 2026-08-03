package repositories

import (
	"bbs-go/internal/models"
)

var DictRepository = newDictRepository()

func newDictRepository() *dictRepository {
	return &dictRepository{}
}

type dictRepository struct {
	BaseRepository[models.Dict]
}
