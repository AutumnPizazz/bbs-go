package eventhandler

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/event"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/pkg/msg"
	"bbs-go/internal/services"
	"reflect"
)

func init() {
	event.RegHandler(reflect.TypeOf(event.QaAnswerAcceptedEvent{}), handleQaAnswerAcceptedEvent)
}

func handleQaAnswerAcceptedEvent(i interface{}) {
	e := i.(event.QaAnswerAcceptedEvent)
	sendQaAnswerAcceptedMsg(e.TopicId, e.UserId)
}

// sendQaAnswerAcceptedMsg 回答被采纳站内通知（通知被采纳者）
func sendQaAnswerAcceptedMsg(topicId, acceptedUserId int64) {
	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return
	}
	if topic.UserId == acceptedUserId {
		return
	}
	var (
		from         = topic.UserId
		to           = acceptedUserId
		quoteContent = "《" + topic.GetTitle() + "》"
	)
	services.MessageService.SendMsg(from, to, msg.TypeQaAnswerAccepted,
		locales.Get("message.qa_answer_accepted_msg_title"), "", quoteContent,
		&msg.QaAnswerAcceptedExtraData{
			TopicId: topicId,
		})
}
