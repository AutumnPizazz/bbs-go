package services

import (
	"testing"

	"bbs-go/internal/models/dto"
)

func TestAttachmentExtAllowedWildcard(t *testing.T) {
	service := &attachmentService{}

	if !service.extAllowed(".exe", []string{"*"}) {
		t.Fatal("expected wildcard to allow any extension")
	}
	if !service.extAllowed("", []string{"*/*"}) {
		t.Fatal("expected MIME wildcard to allow files without an extension")
	}
}

func TestAttachmentPolicyZeroLimitsAreUnlimited(t *testing.T) {
	service := &attachmentService{}
	cfg := dto.AttachmentConfig{
		Enabled:      true,
		AllowedTypes: []string{".pdf"},
		MaxSizeMB:    0,
		MaxCount:     0,
	}

	if err := service.validateAttachmentPolicy(cfg, "guide.pdf", 1024*1024*1024); err != nil {
		t.Fatalf("expected zero size limit to be unlimited, got %v", err)
	}
}

func TestParseAttachmentAllowedTypes(t *testing.T) {
	got := parseAttachmentAllowedTypes("pdf, .DOC .zip\nPDF")
	want := []string{".pdf", ".doc", ".zip"}

	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
