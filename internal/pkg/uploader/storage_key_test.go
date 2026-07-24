package uploader

import "testing"

func TestNormalizeStorageKey(t *testing.T) {
	valid, err := NormalizeStorageKey("attachments/2026/07/23/file.pdf")
	if err != nil || valid != "attachments/2026/07/23/file.pdf" {
		t.Fatalf("valid key=%q err=%v", valid, err)
	}
	for _, value := range []string{"", "/absolute", "../outside", "attachments/../file", "attachments\\file", "attachments/\x00file"} {
		if _, err := NormalizeStorageKey(value); err == nil {
			t.Fatalf("expected invalid storage key %q", value)
		}
	}
}
