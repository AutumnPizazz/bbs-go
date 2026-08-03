package repositories

import (
	"bbs-go/internal/models"
)

var VoteRepository = newVoteRepository()

func newVoteRepository() *voteRepository {
	return &voteRepository{}
}

type voteRepository struct {
	BaseRepository[models.Vote]
}
