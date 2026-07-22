package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
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
