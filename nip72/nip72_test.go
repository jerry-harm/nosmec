package nip72

import (
	"testing"

	"fiatjaf.com/nostr"
)

func TestClassifyRole_Unknown(t *testing.T) {
	event := nostr.Event{
		Kind: nostr.KindTextNote,
		Tags: nostr.Tags{},
	}
	role, ok := ClassifyRole(&event)
	if ok {
		t.Error("expected ok=false for text note, got true")
	}
	if role != Unknown {
		t.Errorf("expected role Unknown, got %v", role)
	}
}

func TestIsCommunityDefinition(t *testing.T) {
	event := &nostr.Event{Kind: nostr.KindCommunityDefinition}
	if !IsCommunityDefinition(event) {
		t.Error("expected true for KindCommunityDefinition")
	}

	event = &nostr.Event{Kind: nostr.KindTextNote}
	if IsCommunityDefinition(event) {
		t.Error("expected false for KindTextNote")
	}

	if IsCommunityDefinition(nil) {
		t.Error("expected false for nil event")
	}
}

func TestGetDefinitionIdentifier(t *testing.T) {
	event := &nostr.Event{
		Kind: nostr.KindCommunityDefinition,
		Tags: nostr.Tags{{"d", "my-community"}},
	}
	if got := GetDefinitionIdentifier(event); got != "my-community" {
		t.Errorf("expected 'my-community', got %q", got)
	}

	event = &nostr.Event{Kind: nostr.KindTextNote}
	if got := GetDefinitionIdentifier(event); got != "" {
		t.Errorf("expected empty string for non-definition, got %q", got)
	}
}