package services

import "testing"

func TestAttachmentExtAllowedWildcard(t *testing.T) {
	service := &attachmentService{}

	if !service.extAllowed(".exe", []string{"*"}) {
		t.Fatal("expected wildcard to allow any extension")
	}
	if !service.extAllowed("", []string{"*/*"}) {
		t.Fatal("expected MIME wildcard to allow files without an extension")
	}
}
