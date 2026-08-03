package repositories

import (
	"bbs-go/internal/models"
)

var LinkRepository = newLinkRepository()

func newLinkRepository() *linkRepository {
	return &linkRepository{}
}

type linkRepository struct {
	BaseRepository[models.Link]
}
