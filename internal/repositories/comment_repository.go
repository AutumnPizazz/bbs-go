package repositories

import (
	"bbs-go/internal/models"
)

var CommentRepository = newCommentRepository()

func newCommentRepository() *commentRepository {
	return &commentRepository{}
}

type commentRepository struct {
	BaseRepository[models.Comment]
}
