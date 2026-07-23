package admin

import (
	"testing"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	modelReq "bbs-go/internal/models/req"
)

func TestAdminBatchConfirmationUsesActionAndEligibleCount(t *testing.T) {
	if got := adminBatchConfirmation("delete", 4); got != "BATCH DELETE 4" {
		t.Fatalf("expected stable confirmation text, got %q", got)
	}
	preview := newAdminBatchPreview("delete", []int64{1, 2, 3}, []int64{1, 3})
	if preview.IneligibleCount != 1 || preview.ConfirmText != "BATCH DELETE 2" {
		t.Fatalf("unexpected preview: %#v", preview)
	}
}

func TestAdminBatchRequestPreservesIdsAndLimit(t *testing.T) {
	duplicate := modelReq.AdminBatchReq{Ids: "1,2,1"}
	if ids := duplicate.ParsedIds(); len(ids) != 3 {
		t.Fatalf("expected parsed ids to preserve request order, got %#v", ids)
	}
	if err := validateBatchConfirmation(modelReq.AdminBatchReq{ConfirmText: "BATCH DELETE 2"}, "delete", 2); err != nil {
		t.Fatalf("expected matching confirmation text to pass: %v", err)
	}
	if err := validateBatchConfirmation(modelReq.AdminBatchReq{ConfirmText: "BATCH DELETE 1"}, "delete", 2); err == nil {
		t.Fatal("expected stale confirmation text to fail")
	}

	if adminBatchMaxSize != 100 {
		t.Fatalf("expected explicit batch size limit, got %d", adminBatchMaxSize)
	}
}

func TestTopicBatchEligibleHonorsCurrentState(t *testing.T) {
	normal := &models.Topic{Status: constants.StatusOk}
	deleted := &models.Topic{Status: constants.StatusDeleted}
	recommended := &models.Topic{Status: constants.StatusOk, Recommend: true}

	if !topicBatchEligible(normal, "recommend") {
		t.Fatal("expected normal topic to be eligible for recommendation")
	}
	if topicBatchEligible(recommended, "recommend") {
		t.Fatal("expected already recommended topic to be skipped")
	}
	if !topicBatchEligible(deleted, "restore") || topicBatchEligible(normal, "restore") {
		t.Fatal("expected restore to target deleted topics only")
	}
	if !topicBatchEligible(normal, "delete") || topicBatchEligible(deleted, "delete") {
		t.Fatal("expected delete to target normal topics only")
	}
}

func TestBatchActionsValidateSupportedValues(t *testing.T) {
	if _, err := topicBatchPermission("unknown"); err == nil {
		t.Fatal("expected unsupported topic action to fail")
	}
	if _, err := userBatchPermission("forbid", 0); err == nil {
		t.Fatal("expected zero-day user ban to fail")
	}
	if status, err := reportBatchStatus("ignore"); err != nil || status != 2 {
		t.Fatalf("expected ignore to map to status 2, got %d, %v", status, err)
	}
}
