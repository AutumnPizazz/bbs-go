package services

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/msg"
	"bbs-go/internal/repositories"
)

var MessageSendTaskService = newMessageSendTaskService()

const MaxMessageBroadcastRecipients = 10000

type messageSendTaskService struct {
	startMu sync.Mutex
}

func newMessageSendTaskService() *messageSendTaskService { return &messageSendTaskService{} }

func (s *messageSendTaskService) Get(id int64) *models.MessageSendTask {
	var task models.MessageSendTask
	if err := sqls.DB().First(&task, id).Error; err != nil {
		return nil
	}
	return &task
}

func (s *messageSendTaskService) Find(cnd *sqls.Cnd) []models.MessageSendTask {
	var tasks []models.MessageSendTask
	query := sqls.DB().Model(&models.MessageSendTask{})
	if cnd != nil {
		query = cnd.Build(query)
	}
	query.Find(&tasks)
	return tasks
}

func (s *messageSendTaskService) FindPage(cnd *sqls.Cnd) ([]models.MessageSendTask, *sqls.Paging) {
	if cnd == nil {
		cnd = sqls.NewCnd().Desc("id")
	}
	var tasks []models.MessageSendTask
	cnd.Find(sqls.DB(), &tasks)
	return tasks, cnd.Paging
}

func (s *messageSendTaskService) RecipientIds(targetType string, targetId int64) ([]int64, error) {
	targetType = strings.TrimSpace(targetType)
	if targetType != "all" && targetType != "role" && targetType != "category" {
		return nil, errors.New("targetType must be all, role, or category")
	}
	if targetType != "all" && targetId <= 0 {
		return nil, errors.New("targetId is required for role or category targets")
	}
	query := sqls.DB().Model(&models.User{}).Where("status = ?", constants.StatusOk).Select("id")
	switch targetType {
	case "role":
		query = query.Where("id IN (?)", sqls.DB().Model(&models.UserRole{}).Select("user_id").Where("role_id = ?", targetId))
	case "category":
		query = query.Where("id IN (?)", sqls.DB().Model(&models.Topic{}).Select("user_id").Where("category_id = ? AND status <> ?", targetId, constants.StatusDeleted))
	}
	var ids []int64
	if err := query.Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	if len(ids) > MaxMessageBroadcastRecipients {
		return nil, fmt.Errorf("recipient count exceeds the limit of %d", MaxMessageBroadcastRecipients)
	}
	return ids, nil
}

func (s *messageSendTaskService) CreateDraft(creatorId, fromId int64, title, content, targetType string, targetId int64) (*models.MessageSendTask, error) {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return nil, errors.New("title and content are required")
	}
	ids, err := s.RecipientIds(targetType, targetId)
	if err != nil {
		return nil, err
	}
	now := dates.NowTimestamp()
	task := &models.MessageSendTask{
		CreatorId: creatorId, FromId: fromId, Title: strings.TrimSpace(title), Content: strings.TrimSpace(content),
		TargetType: strings.TrimSpace(targetType), TargetId: targetId, Status: models.MessageTaskDraft,
		TotalCount: int64(len(ids)), CreateTime: now, UpdateTime: now,
	}
	if err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for _, userId := range ids {
			if err := tx.Create(&models.MessageDelivery{TaskId: task.Id, UserId: userId, Status: models.MessageDeliveryPending, CreateTime: now, UpdateTime: now}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *messageSendTaskService) Start(id int64) (*models.MessageSendTask, error) {
	return s.start(id, false)
}

func (s *messageSendTaskService) Retry(id int64) (*models.MessageSendTask, error) {
	return s.start(id, true)
}

func (s *messageSendTaskService) start(id int64, retry bool) (*models.MessageSendTask, error) {
	s.startMu.Lock()
	defer s.startMu.Unlock()

	task := s.Get(id)
	if task == nil {
		return nil, errors.New("message task not found")
	}
	if task.Status == models.MessageTaskRunning {
		return task, errors.New("message task is already running")
	}
	if retry && task.Status != models.MessageTaskFailed {
		return task, errors.New("only failed message tasks can be retried")
	}
	if !retry && task.Status != models.MessageTaskDraft {
		return task, errors.New("only draft message tasks can be sent")
	}
	var running int64
	sqls.DB().Model(&models.MessageSendTask{}).Where("status = ? AND id <> ?", models.MessageTaskRunning, id).Count(&running)
	if running > 0 {
		return task, errors.New("another message task is already running")
	}
	now := dates.NowTimestamp()
	if retry {
		if err := sqls.DB().Model(&models.MessageDelivery{}).Where("task_id = ? AND status = ?", id, models.MessageDeliveryFailed).
			Updates(map[string]interface{}{"status": models.MessageDeliveryPending, "error": "", "update_time": now}).Error; err != nil {
			return task, err
		}
	}
	if err := sqls.DB().Model(&models.MessageSendTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": models.MessageTaskRunning, "error": "", "started_at": now, "finished_at": 0, "update_time": now,
	}).Error; err != nil {
		return task, err
	}
	go s.run(id)
	task.Status = models.MessageTaskRunning
	task.StartedAt = now
	task.Error = ""
	return task, nil
}

func (s *messageSendTaskService) run(id int64) {
	defer func() {
		if recovered := recover(); recovered != nil {
			s.failTask(id, fmt.Errorf("message task panicked: %v", recovered))
		}
	}()

	task := s.Get(id)
	if task == nil {
		return
	}
	var deliveries []models.MessageDelivery
	if err := sqls.DB().Where("task_id = ? AND status = ?", id, models.MessageDeliveryPending).Find(&deliveries).Error; err != nil {
		s.failTask(id, err)
		return
	}
	for i := range deliveries {
		delivery := &deliveries[i]
		if err := s.deliver(task, delivery); err != nil {
			if updateErr := sqls.DB().Model(&models.MessageDelivery{}).
				Where("id = ? AND status = ?", delivery.Id, models.MessageDeliveryPending).
				Updates(map[string]interface{}{
					"status":      models.MessageDeliveryFailed,
					"attempts":    delivery.Attempts + 1,
					"error":       err.Error(),
					"update_time": dates.NowTimestamp(),
				}).Error; updateErr != nil {
				s.failTask(id, fmt.Errorf("record delivery failure: %w", updateErr))
				return
			}
		}
	}
	var sent, failed int64
	if err := sqls.DB().Model(&models.MessageDelivery{}).Where("task_id = ? AND status = ?", id, models.MessageDeliverySent).Count(&sent).Error; err != nil {
		s.failTask(id, err)
		return
	}
	if err := sqls.DB().Model(&models.MessageDelivery{}).Where("task_id = ? AND status = ?", id, models.MessageDeliveryFailed).Count(&failed).Error; err != nil {
		s.failTask(id, err)
		return
	}
	status := models.MessageTaskCompleted
	errorMessage := ""
	if failed > 0 {
		status = models.MessageTaskFailed
		errorMessage = fmt.Sprintf("%d message deliveries failed", failed)
	}
	if err := sqls.DB().Model(&models.MessageSendTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status, "sent_count": sent, "failed_count": failed, "error": errorMessage,
		"finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
	}).Error; err != nil {
		s.failTask(id, err)
	}
}

func (s *messageSendTaskService) deliver(task *models.MessageSendTask, delivery *models.MessageDelivery) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		var current models.MessageDelivery
		if err := tx.First(&current, "id = ?", delivery.Id).Error; err != nil {
			return err
		}
		if current.Status != models.MessageDeliveryPending {
			return nil
		}
		attempts := current.Attempts + 1
		if err := tx.Model(&models.MessageDelivery{}).Where("id = ?", current.Id).
			Update("attempts", attempts).Error; err != nil {
			return err
		}
		message := &models.Message{FromId: task.FromId, UserId: current.UserId, Title: task.Title, Content: task.Content, Type: int(msg.TypeAdminAnnouncement), Status: msg.StatusUnread, CreateTime: dates.NowTimestamp()}
		if err := repositories.MessageRepository.Create(tx, message); err != nil {
			return err
		}
		return tx.Model(&models.MessageDelivery{}).Where("id = ?", current.Id).Updates(map[string]interface{}{
			"status": models.MessageDeliverySent, "message_id": message.Id, "error": "", "update_time": dates.NowTimestamp(),
		}).Error
	})
}

func (s *messageSendTaskService) failTask(id int64, err error) {
	if err == nil {
		return
	}
	_ = sqls.DB().Model(&models.MessageSendTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": models.MessageTaskFailed, "error": err.Error(), "finished_at": dates.NowTimestamp(), "update_time": dates.NowTimestamp(),
	}).Error
}

func (s *messageSendTaskService) Deliveries(taskId int64, status *int) []models.MessageDelivery {
	var list []models.MessageDelivery
	query := sqls.DB().Where("task_id = ?", taskId)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Order("id asc").Find(&list)
	return list
}
