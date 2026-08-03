package repositories

import (
	"bbs-go/internal/models"
)

var VoteRecordRepository = newVoteRecordRepository()

func newVoteRecordRepository() *voteRecordRepository {
	return &voteRecordRepository{}
}

type voteRecordRepository struct {
	BaseRepository[models.VoteRecord]
}
