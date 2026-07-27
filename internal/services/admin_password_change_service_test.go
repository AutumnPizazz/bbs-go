package services

import (
	"testing"
	"time"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/repositories"

	"github.com/mlogclub/simple/common/passwd"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func TestAdminPasswordChangeKeepsOldPasswordUntilNewLogin(t *testing.T) {
	db := setupAdminPasswordChangeTestDB(t)

	user := &models.User{
		Username:   sqls.SqlNullString("owner-test"),
		Nickname:   "Owner",
		Password:   passwd.EncodePassword("old-password"),
		Roles:      constants.RoleOwner,
		Status:     constants.StatusOk,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}
	if err := repositories.UserRepository.Create(db, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.UserToken{
		Token:      "old-session",
		UserId:     user.Id,
		ExpiredAt:  time.Now().Add(time.Hour).UnixMilli(),
		Status:     constants.StatusOk,
		CreateTime: time.Now().UnixMilli(),
	}).Error; err != nil {
		t.Fatalf("create token: %v", err)
	}

	if err := UserService.UpdatePasswordByAdmin(user, user.Id, "old-password", "new-password", "new-password", nil); err != nil {
		t.Fatalf("stage password change: %v", err)
	}
	stored := repositories.UserRepository.Get(db, user.Id)
	if !passwd.ValidatePassword(stored.Password, "old-password") {
		t.Fatal("expected old password to remain active while pending")
	}
	if repositories.AdminPasswordChangeRepository.GetByUserId(db, user.Id) == nil {
		t.Fatal("expected pending password change")
	}

	oldLogin, err := UserService.SignInWithPassword("owner-test", "old-password")
	if err != nil || oldLogin.PendingAdminPassword {
		t.Fatalf("expected old password login to remain normal, result=%#v err=%v", oldLogin, err)
	}
	newLogin, err := UserService.SignInWithPassword("owner-test", "new-password")
	if err != nil || !newLogin.PendingAdminPassword {
		t.Fatalf("expected new password login to require confirmation, result=%#v err=%v", newLogin, err)
	}

	if err := UserService.CommitPendingAdminPassword(user.Id, "new-password"); err != nil {
		t.Fatalf("commit pending password: %v", err)
	}
	stored = repositories.UserRepository.Get(db, user.Id)
	if !passwd.ValidatePassword(stored.Password, "new-password") {
		t.Fatal("expected new password to become active")
	}
	if repositories.AdminPasswordChangeRepository.GetByUserId(db, user.Id) != nil {
		t.Fatal("expected pending password change to be deleted")
	}
	var token models.UserToken
	if err := db.Where("token = ?", "old-session").First(&token).Error; err != nil {
		t.Fatalf("load old token: %v", err)
	}
	if token.Status != constants.StatusDeleted {
		t.Fatalf("expected old session to be revoked, got status %d", token.Status)
	}
}

func TestAdminPasswordChangeRejectsWrongCurrentPasswordWithoutChangingState(t *testing.T) {
	db := setupAdminPasswordChangeTestDB(t)

	user := &models.User{
		Username:   sqls.SqlNullString("owner-test-wrong-current"),
		Nickname:   "Owner",
		Password:   passwd.EncodePassword("old-password"),
		Roles:      constants.RoleOwner,
		Status:     constants.StatusOk,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}
	if err := repositories.UserRepository.Create(db, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := UserService.UpdatePasswordByAdmin(user, user.Id, "wrong-password", "new-password", "new-password", nil); err == nil {
		t.Fatal("expected wrong current password to be rejected")
	}
	if repositories.AdminPasswordChangeRepository.GetByUserId(db, user.Id) != nil {
		t.Fatal("did not expect a pending password change")
	}
}

func TestAdminPasswordChangeExpiresWithoutChangingFormalPassword(t *testing.T) {
	db := setupAdminPasswordChangeTestDB(t)

	user := &models.User{
		Username:   sqls.SqlNullString("owner-test-expired"),
		Nickname:   "Owner",
		Password:   passwd.EncodePassword("old-password"),
		Roles:      constants.RoleOwner,
		Status:     constants.StatusOk,
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}
	if err := repositories.UserRepository.Create(db, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := UserService.UpdatePasswordByAdmin(user, user.Id, "old-password", "new-password", "new-password", nil); err != nil {
		t.Fatalf("stage password change: %v", err)
	}
	pending := repositories.AdminPasswordChangeRepository.GetByUserId(db, user.Id)
	pending.ExpiresAt = time.Now().Add(-time.Minute).UnixMilli()
	if err := repositories.AdminPasswordChangeRepository.Update(db, pending); err != nil {
		t.Fatalf("expire pending password change: %v", err)
	}

	if _, err := UserService.SignInWithPassword("owner-test-expired", "new-password"); err == nil {
		t.Fatal("expected expired password to be rejected")
	}
	if _, err := UserService.SignInWithPassword("owner-test-expired", "old-password"); err != nil {
		t.Fatalf("expected old password to remain active: %v", err)
	}
	stored := repositories.UserRepository.Get(db, user.Id)
	if !passwd.ValidatePassword(stored.Password, "old-password") {
		t.Fatal("expected formal password to remain unchanged")
	}
}

func setupAdminPasswordChangeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	originalConfig := config.Instance
	config.Instance = &config.Config{Language: config.LanguageEnUS}
	t.Cleanup(func() { config.Instance = originalConfig })

	db := setupTestDB(t)
	if err := db.AutoMigrate(&models.UserToken{}, &models.AdminPasswordChange{}, &models.OperateLog{}); err != nil {
		t.Fatalf("auto migrate password change models: %v", err)
	}
	return db
}
