package services

import (
	"testing"
	"time"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestMessageSendTaskCreatesOneMessagePerRecipient(t *testing.T) {
	dsn := "file:message_broadcast_test_" + time.Now().Format("20060102150405.000000000") + "?mode=memory&cache=shared&_fk=1"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "t_", SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqls.SetDB(db)
	if err := db.AutoMigrate(&models.User{}, &models.Message{}, &models.MessageSendTask{}, &models.MessageDelivery{}); err != nil {
		t.Fatal(err)
	}
	user := &models.User{Nickname: "recipient", Status: constants.StatusOk, CreateTime: time.Now().UnixMilli(), UpdateTime: time.Now().UnixMilli()}
	if err := repositories.UserRepository.Create(db, user); err != nil {
		t.Fatal(err)
	}

	task, err := MessageSendTaskService.CreateDraft(1, 1, "Announcement", "Content", "all", 0)
	if err != nil {
		t.Fatal(err)
	}
	if task.TotalCount != 1 {
		t.Fatalf("total=%d want 1", task.TotalCount)
	}
	if _, err := MessageSendTaskService.Start(task.Id); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		current := MessageSendTaskService.Get(task.Id)
		if current != nil && current.Status != models.MessageTaskRunning {
			if current.Status != models.MessageTaskCompleted || current.SentCount != 1 || current.FailedCount != 0 {
				t.Fatalf("unexpected task: %#v", current)
			}
			var count int64
			if err := db.Model(&models.Message{}).Where("user_id = ?", user.Id).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("messages=%d want 1", count)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("message task did not finish")
}
