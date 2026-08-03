package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
)

var OperateLogService = newOperateLogService()

func newOperateLogService() *operateLogService {
	return &operateLogService{BaseService: newBaseService[models.OperateLog, crudRepository[models.OperateLog]](repositories.OperateLogRepository)}
}

type operateLogService struct {
	BaseService[models.OperateLog, crudRepository[models.OperateLog]]
}

func (s *operateLogService) Delete(id int64) {
	repositories.OperateLogRepository.Delete(sqls.DB(), id)
}

func (s *operateLogService) AddOperateLog(userId int64, opType, dataType string, dataId int64,
	description string, r *http.Request) {
	s.AddOperateLogResult(userId, opType, dataType, dataId, description, "success", r)
}

func (s *operateLogService) AddOperateLogResult(userId int64, opType, dataType string, dataId int64,
	description, result string, r *http.Request) {
	if strings.EqualFold(strings.TrimSpace(result), "success") {
		result = "success"
	} else {
		result = "failure"
	}

	operateLog := &models.OperateLog{
		UserId:      userId,
		OpType:      opType,
		DataType:    dataType,
		DataId:      dataId,
		Description: description,
		Result:      result,
		CreateTime:  dates.NowTimestamp(),
	}
	if r != nil {
		operateLog.Ip = web.GetRequestIP(r)
		operateLog.UserAgent = web.GetUserAgent(r)
		operateLog.Referer = r.Header.Get("Referer")
	}
	if err := repositories.OperateLogRepository.Create(sqls.DB(), operateLog); err != nil {
		slog.Error(err.Error(), slog.Any("err", err))
	}
}

func (s *operateLogService) AddOperateLogFailure(userId int64, opType, dataType string, dataId int64,
	description string, cause error, r *http.Request) {
	if cause != nil {
		message := strings.TrimSpace(cause.Error())
		if len(message) > 768 {
			message = message[:768]
		}
		if message != "" {
			description = strings.TrimSpace(description) + ": " + message
		}
	}
	s.AddOperateLogResult(userId, opType, dataType, dataId, description, "failure", r)
}
