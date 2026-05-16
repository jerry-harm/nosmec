package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"fiatjaf.com/nostr"
)

func TestPrintEvent_JsonOutput(t *testing.T) {
	pk, err := nostr.PubKeyFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("failed to create pubkey: %v", err)
	}
	ev := &nostr.Event{
		ID:         [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		PubKey:     pk,
		CreatedAt: nostr.Timestamp(1700000000),
		Kind:       nostr.KindTextNote,
		Content:    "Hello, World!",
		Tags:       nostr.Tags{},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call PrintEvent with j=true (JSON output)
	PrintEvent(ev, true)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	// Verify output is valid JSON
	var decoded nostr.Event
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Errorf("expected valid JSON output, got error: %v", err)
	}

	if decoded.Content != ev.Content {
		t.Errorf("decoded.Content = %q, want %q", decoded.Content, ev.Content)
	}
	if decoded.Kind != ev.Kind {
		t.Errorf("decoded.Kind = %v, want %v", decoded.Kind, ev.Kind)
	}
}

func TestPrintEvent_EmptyContent(t *testing.T) {
	// Use same valid pubkey as TestPrintEvent_JsonOutput - the pubkey value doesn't matter for this test
	pk, err := nostr.PubKeyFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("failed to create pubkey: %v", err)
	}
	ev := &nostr.Event{
		ID:         [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		PubKey:     pk,
		CreatedAt:  nostr.Timestamp(1700000000),
		Kind:       nostr.KindTextNote,
		Content:    "",
		Tags:       nostr.Tags{},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintEvent(ev, true)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	var decoded nostr.Event
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Errorf("expected valid JSON output, got error: %v", err)
	}

	if decoded.Content != "" {
		t.Errorf("decoded.Content = %q, want empty string", decoded.Content)
	}
}

func TestPrintEvent_ZeroValueEvent(t *testing.T) {
	ev := &nostr.Event{
		ID:      [32]byte{},
		PubKey:  nostr.PubKey{},
		CreatedAt: nostr.Timestamp(0),
		Kind:   0,
		Content: "",
		Tags:   nostr.Tags{},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintEvent(ev, true)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	var decoded nostr.Event
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Errorf("expected valid JSON output, got error: %v", err)
	}

	// Verify zero values are preserved
	if decoded.Kind != 0 {
		t.Errorf("decoded.Kind = %v, want 0", decoded.Kind)
	}
}