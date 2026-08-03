package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var VoteRecordService = newVoteRecordService()

func newVoteRecordService() *voteRecordService {
	return &voteRecordService{BaseService: newBaseService[models.VoteRecord, crudRepository[models.VoteRecord]](repositories.VoteRecordRepository)}
}

type voteRecordService struct {
	BaseService[models.VoteRecord, crudRepository[models.VoteRecord]]
}

func (s *voteRecordService) Delete(id int64) {
	repositories.VoteRecordRepository.Delete(sqls.DB(), id)
}

func (s *voteRecordService) GetBy(userId, voteId int64) *models.VoteRecord {
	return s.FindOne(sqls.NewCnd().Where("user_id = ? and vote_id = ?", userId, voteId))
}
