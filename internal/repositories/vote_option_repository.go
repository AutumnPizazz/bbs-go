package repositories

import (
	"bbs-go/internal/models"
)

var VoteOptionRepository = newVoteOptionRepository()

func newVoteOptionRepository() *voteOptionRepository {
	return &voteOptionRepository{}
}

type voteOptionRepository struct {
	BaseRepository[models.VoteOption]
}
