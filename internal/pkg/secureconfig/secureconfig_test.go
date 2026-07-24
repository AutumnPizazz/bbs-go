package secureconfig

import (
	"testing"

	"bbs-go/internal/pkg/config"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Setenv("BBSGO_CONFIG_ENCRYPTION_KEY", "test-encryption-key")
	sealed, err := Encrypt("provider-secret")
	if err != nil {
		t.Fatal(err)
	}
	if sealed == "provider-secret" || !IsEncrypted(sealed) {
		t.Fatalf("expected encrypted value, got %q", sealed)
	}
	plain, err := Decrypt(sealed)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "provider-secret" {
		t.Fatalf("plain=%q", plain)
	}
}

func TestPlaintextRemainsCompatible(t *testing.T) {
	plain, err := Decrypt("legacy-plaintext")
	if err != nil || plain != "legacy-plaintext" {
		t.Fatalf("plain=%q err=%v", plain, err)
	}
}

func TestEncryptRequiresExplicitKey(t *testing.T) {
	previous := config.Instance
	config.Instance = &config.Config{}
	t.Cleanup(func() { config.Instance = previous })
	t.Setenv("BBSGO_CONFIG_ENCRYPTION_KEY", "")
	if _, err := Encrypt("secret"); err == nil {
		t.Fatal("expected encryption to fail without an explicit key")
	}
}
