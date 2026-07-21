package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"testing"
)

func TestNormalizeAttachmentConfig_PreservesUnlimitedValues(t *testing.T) {
	cfg := normalizeAttachmentConfig(dto.AttachmentConfig{MaxSizeMB: -1, MaxCount: -1})

	if cfg.MaxSizeMB != -1 || cfg.MaxCount != -1 {
		t.Fatalf("expected -1 attachment limits to remain unlimited, got size=%d count=%d", cfg.MaxSizeMB, cfg.MaxCount)
	}
}

func TestNormalizeAttachmentConfig_DefaultsZeroValues(t *testing.T) {
	cfg := normalizeAttachmentConfig(dto.AttachmentConfig{})

	if cfg.MaxSizeMB != 10 || cfg.MaxCount != 5 {
		t.Fatalf("expected zero attachment limits to use defaults, got size=%d count=%d", cfg.MaxSizeMB, cfg.MaxCount)
	}
}

func TestParseModulesConfig_BackfillsQaFromTopicForLegacyConfig(t *testing.T) {
	cfg := parseModulesConfig(`{"tweet":true,"topic":true,"article":false}`)

	if !cfg.QA {
		t.Fatalf("expected legacy config without qa to keep QA enabled when topic is enabled")
	}
}

func TestParseModulesConfig_RespectsExplicitQaSwitch(t *testing.T) {
	cfg := parseModulesConfig(`{"tweet":true,"topic":true,"qa":false,"article":true}`)

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
