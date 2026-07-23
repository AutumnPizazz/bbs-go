package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/services"
)

func TestAttachmentRemoveSoftDeletesRecord(t *testing.T) {
	db := setupAdminAttachmentTestDB(t)
	attachment := &models.Attachment{
		Id:         "attachment-soft-delete",
		TopicId:    0,
		UserId:     1,
		FileName:   "draft.pdf",
		Status:     constants.StatusOk,
		CreateTime: time.Now().UnixMilli(),
	}
	if err := db.Create(attachment).Error; err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/admin/attachment/delete", strings.NewReader("ids=attachment-soft-delete"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	common.SetCurrentUser(ctx, &models.User{Model: models.Model{Id: 123456}, Roles: constants.RoleOwner})
	AttachmentRemove(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", recorder.Code, http.StatusOK)
	}
	var saved models.Attachment
	if err := db.First(&saved, "id = ?", attachment.Id).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Status != constants.StatusDeleted {
		t.Fatalf("status=%d want deleted", saved.Status)
	}
}

func TestAttachmentCleanupOnlyMarksOldOrphans(t *testing.T) {
	db := setupAdminAttachmentTestDB(t)
	now := time.Now().UnixMilli()
	attachments := []*models.Attachment{
		{Id: "old-orphan", TopicId: 0, Status: constants.StatusOk, CreateTime: now - 10_000},
		{Id: "new-orphan", TopicId: 0, Status: constants.StatusOk, CreateTime: now},
		{Id: "bound", TopicId: 42, Status: constants.StatusOk, CreateTime: now - 10_000},
	}
	for _, attachment := range attachments {
		if err := db.Create(attachment).Error; err != nil {
			t.Fatal(err)
		}
	}

	count, err := services.AttachmentService.CleanupOrphans(now - 1_000)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("cleaned=%d want 1", count)
	}

	var old, current, bound models.Attachment
	if err := db.First(&old, "id = ?", "old-orphan").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&current, "id = ?", "new-orphan").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&bound, "id = ?", "bound").Error; err != nil {
		t.Fatal(err)
	}
	if old.Status != constants.StatusDeleted || current.Status != constants.StatusOk || bound.Status != constants.StatusOk {
		t.Fatalf("unexpected statuses old=%d current=%d bound=%d", old.Status, current.Status, bound.Status)
	}
}

func setupAdminAttachmentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:admin_attachment_test_" + time.Now().Format("20060102150405.000000000") + "?mode=memory&cache=shared&_fk=1"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "t_", SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqls.SetDB(db)
	if err := db.AutoMigrate(&models.Attachment{}, &models.Category{}, &models.OperateLog{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Category{Model: models.Model{Id: 1}, Name: "test", Status: constants.StatusOk}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}
