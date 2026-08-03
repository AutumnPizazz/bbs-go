package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var VoteOptionService = newVoteOptionService()

func newVoteOptionService() *voteOptionService {
	return &voteOptionService{BaseService: newBaseService[models.VoteOption, crudRepository[models.VoteOption]](repositories.VoteOptionRepository)}
}

type voteOptionService struct {
	BaseService[models.VoteOption, crudRepository[models.VoteOption]]
}

func (s *voteOptionService) Delete(id int64) {
	repositories.VoteOptionRepository.Delete(sqls.DB(), id)
}

func (s *voteOptionService) FindByVoteId(voteId int64) []models.VoteOption {
	return repositories.VoteOptionRepository.Find(sqls.DB(), sqls.NewCnd().Eq("vote_id", voteId).Asc("sort_no").Asc("id"))
}
