package repositories

import (
	"bbs-go/internal/models"
)

var OperateLogRepository = newOperateLogRepository()

func newOperateLogRepository() *operateLogRepository {
	return &operateLogRepository{}
}

type operateLogRepository struct {
	BaseRepository[models.OperateLog]
}
