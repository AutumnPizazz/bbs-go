package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/pkg/msg"
	"bbs-go/internal/repositories"
	"log/slog"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
)

var MessageService = newMessageService()

func newMessageService() *messageService {
	return &messageService{BaseService: newBaseService[models.Message, crudRepository[models.Message]](repositories.MessageRepository)}
}

type messageService struct {
	BaseService[models.Message, crudRepository[models.Message]]
}

func (s *messageService) Delete(id int64) {
	repositories.MessageRepository.Delete(sqls.DB(), id)
}

// GetUnReadCount 获取未读消息数量
func (s *messageService) GetUnReadCount(userId int64) (count int64) {
	sqls.DB().Where("user_id = ? and status = ?", userId, msg.StatusUnread).Model(&models.Message{}).Count(&count)
	return
}

// MarkRead 将所有消息标记为已读
func (s *messageService) MarkRead(userId int64) {
	sqls.DB().Exec("update t_message set status = ? where user_id = ? and status = ?", msg.StatusHaveRead,
		userId, msg.StatusUnread)
}

// SendMsg 发送站内消息
func (s *messageService) SendMsg(from, to int64, msgType msg.Type,
	title, content, quoteContent string, extraData interface{}) {
	if !SysConfigService.IsSiteNoticeEnabled(msgType) {
		return
	}
	t := &models.Message{
		FromId:       from,
		UserId:       to,
		Title:        title,
		Content:      content,
		QuoteContent: quoteContent,
		Type:         int(msgType),
		ExtraData:    jsons.ToJsonStr(extraData),
		Status:       msg.StatusUnread,
		CreateTime:   dates.NowTimestamp(),
	}
	if err := s.Create(t); err != nil {
		slog.Error(err.Error(), slog.Any("err", err))
	}
}
