package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeAttachmentConfig_PreservesUnlimitedValues(t *testing.T) {
	cfg := normalizeAttachmentConfig(dto.AttachmentConfig{MaxSizeMB: 0, MaxCount: 0})

	if cfg.MaxSizeMB != 0 || cfg.MaxCount != 0 {
		t.Fatalf("expected 0 attachment limits to remain unlimited, got size=%d count=%d", cfg.MaxSizeMB, cfg.MaxCount)
	}
}

func TestNormalizeAttachmentConfig_DefaultsAllowedTypes(t *testing.T) {
	cfg := normalizeAttachmentConfig(dto.AttachmentConfig{})

	if len(cfg.AllowedTypes) == 0 {
		t.Fatal("expected empty allowed types to use defaults")
	}
	if cfg.MaxSizeMB != 0 || cfg.MaxCount != 0 {
		t.Fatalf("expected zero attachment limits to remain unlimited, got size=%d count=%d", cfg.MaxSizeMB, cfg.MaxCount)
	}
}

func TestParseModulesConfig_BackfillsQaFromTopicForLegacyConfig(t *testing.T) {
	cfg := parseModulesConfig(`{"topic":true,"qa":true}`)

	if !cfg.QA {
		t.Fatalf("expected legacy config without qa to keep QA enabled when topic is enabled")
	}
}

func TestParseModulesConfig_RespectsExplicitQaSwitch(t *testing.T) {
	cfg := parseModulesConfig(`{"topic":true,"qa":false}`)

	if cfg.QA {
		t.Fatalf("expected explicit qa=false to disable QA independently from topic")
	}
}

func TestNormalizeTopicListStyle_DefaultsToStandardStyle(t *testing.T) {
	if got := normalizeTopicListStyle(""); got != constants.TopicListStyleDefault {
		t.Fatalf("expected empty topic list style to default to %q, got %q", constants.TopicListStyleDefault, got)
	}
}

func TestNormalizeTopicListStyle_AcceptsCompactStyle(t *testing.T) {
	if got := normalizeTopicListStyle(constants.TopicListStyleCompact); got != constants.TopicListStyleCompact {
		t.Fatalf("expected compact topic list style, got %q", got)
	}
}

func TestNormalizeTopicListStyle_RejectsUnknownStyle(t *testing.T) {
	if got := normalizeTopicListStyle("dense"); got != constants.TopicListStyleDefault {
		t.Fatalf("expected unknown topic list style to default to %q, got %q", constants.TopicListStyleDefault, got)
	}
}

func TestDecodeAdminConfigRejectsUnknownKeysAndSensitiveValues(t *testing.T) {
	if _, err := decodeAdminConfig(`{"unexpected":"value"}`, false); err == nil {
		t.Fatal("expected arbitrary configuration keys to be rejected")
	}
	if _, err := decodeAdminConfig(`{"loginConfig":{}}`, false); err == nil {
		t.Fatal("expected sensitive configuration to require the sensitive endpoint")
	}
}

func TestMergeSensitiveConfigPreservesBlankCredentialsAndSupportsExplicitClear(t *testing.T) {
	merged, err := mergeSensitiveConfig(
		`{"googleLogin":{"enabled":true,"clientSecret":"secret-value"}}`,
		`{"googleLogin":{"enabled":false,"clientSecret":""}}`,
		constants.SysConfigLoginConfig,
		nil,
	)
	if err != nil {
		t.Fatalf("merge sensitive config: %v", err)
	}
	if !strings.Contains(merged, "secret-value") || !strings.Contains(merged, `"enabled":false`) {
		t.Fatalf("expected blank secret to be preserved while other fields update, got %s", merged)
	}

	cleared, err := mergeSensitiveConfig(
		`{"googleLogin":{"clientSecret":"secret-value"}}`,
		`{"googleLogin":{}}`,
		constants.SysConfigLoginConfig,
		[]string{"loginConfig.googleLogin.clientSecret"},
	)
	if err != nil {
		t.Fatalf("clear sensitive config: %v", err)
	}
	if strings.Contains(cleared, "secret-value") {
		t.Fatalf("expected explicit clear to remove the secret, got %s", cleared)
	}
}

func TestDashboardConfigDTODoesNotSerializeCredentialValues(t *testing.T) {
	value, err := json.Marshal(dto.LoginConfigAdmin{
		GoogleLogin: dto.OAuthLoginAdmin{ClientSecretConfigured: true},
	})
	if err != nil {
		t.Fatalf("marshal dashboard config: %v", err)
	}
	if strings.Contains(string(value), "secret-value") || strings.Contains(string(value), "clientSecret\":\"secret") {
		t.Fatalf("dashboard config must not contain credential values: %s", value)
	}
}
