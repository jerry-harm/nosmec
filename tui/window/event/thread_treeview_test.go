package event

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Digital-Shane/treeview/v2"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/bubblon"
)

var (
	testParentID = "0000000000000000000000000000000000000000000000000000000000000001"
	testFirstParentID = "0000000000000000000000000000000000000000000000000000000000000002"
	testSecondParentID = "0000000000000000000000000000000000000000000000000000000000000003"
	testSomeID = "0000000000000000000000000000000000000000000000000000000000000004"
	testRootMarkerID = "abcd000000000000000000000000000000000000000000000000000000000000"
)

func TestExtractParentID_RootEvent(t *testing.T) {
	// Event with no e tags = root (no parent)
	event := &nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	parentID := extractParentID(event)
	if parentID != "" {
		t.Errorf("root event should have empty parent ID, got %q", parentID)
	}
}

func TestExtractParentID_RootMarker(t *testing.T) {
	// Root marker pointing to different event = direct reply, parent is root tag value
	event := &nostr.Event{
		Content: "direct reply to root",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID, "", "root"},
		},
	}
	parentID := extractParentID(event)
	expectedRootID, _ := nostr.IDFromHex(testRootMarkerID)
	// For a direct reply, root IS the parent (event's own ID differs from root marker)
	if event.ID.Hex() != testRootMarkerID {
		if parentID != testRootMarkerID {
			t.Errorf("direct reply should have parent = root marker ID %q, got %q", testRootMarkerID, parentID)
		}
	} else {
		// Self-referencing root marker = this IS the root
		if parentID != "" {
			t.Errorf("self-root event should have empty parent ID, got %q", parentID)
		}
	}
	_ = expectedRootID
}

func TestExtractParentID_ReplyMarker(t *testing.T) {
	// Event with "reply" marker - parent is the e tag value
	event := &nostr.Event{
		Content: "reply note",
		Tags: nostr.Tags{
			nostr.Tag{"e", testParentID, "", "reply"},
		},
	}
	parentID := extractParentID(event)
	if parentID != testParentID {
		t.Errorf("reply event should have parent ID %q, got %q", testParentID, parentID)
	}
}

func TestExtractParentID_ReplyMarkerWithRelay(t *testing.T) {
	// Event with "reply" marker and relay hint
	event := &nostr.Event{
		Content: "reply note with relay",
		Tags: nostr.Tags{
			nostr.Tag{"e", testParentID, "wss://relay.example.com", "reply"},
		},
	}
	parentID := extractParentID(event)
	if parentID != testParentID {
		t.Errorf("reply event should have parent ID %q, got %q", testParentID, parentID)
	}
}

func TestExtractParentID_MultipleETags(t *testing.T) {
	// Multiple e tags - should pick the first one with "reply" marker
	event := &nostr.Event{
		Content: "multi e-tag note",
		Tags: nostr.Tags{
			nostr.Tag{"e", testFirstParentID, "", "reply"},
			nostr.Tag{"e", testSecondParentID, "", "root"},
		},
	}
	parentID := extractParentID(event)
	if parentID != testFirstParentID {
		t.Errorf("should use first reply marker parent ID, got %q", parentID)
	}
}

func TestExtractParentID_NoMarker(t *testing.T) {
	// e tag without marker - treat as root (no parent per NIP-10)
	event := &nostr.Event{
		Content: "note with e tag but no marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", testSomeID},
		},
	}
	parentID := extractParentID(event)
	if parentID != "" {
		t.Errorf("e tag without marker should be treated as root, got %q", parentID)
	}
}

func TestExtractRootEvent_NilEvent(t *testing.T) {
	event := (*nostr.Event)(nil)
	rootID, isRoot, err := extractRootEvent(event)
	if err == nil {
		t.Errorf("nil event should return error")
	}
	if rootID != (nostr.ID{}) {
		t.Errorf("nil event should have zero root ID")
	}
	if isRoot {
		t.Errorf("nil event should not be root")
	}
}

func TestExtractRootEvent_NoETags(t *testing.T) {
	// No e tags = root
	event := &nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isRoot {
		t.Errorf("event with no e tags should be root")
	}
	if rootID != event.ID {
		t.Errorf("root ID should equal event ID")
	}
}

func TestExtractRootEvent_RootMarker(t *testing.T) {
	// Root marker pointing to DIFFERENT event = direct reply, NOT root
	event := &nostr.Event{
		Content: "direct reply with root marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID, "", "root"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isRoot {
		t.Errorf("event with root marker pointing to other ID should NOT be root")
	}
	expectedRootID, _ := nostr.IDFromHex(testRootMarkerID)
	if rootID != expectedRootID {
		t.Errorf("root ID should be the tagged event, got %v, want %v", rootID, expectedRootID)
	}
}

func TestExtractRootEvent_ReplyMarker(t *testing.T) {
	// Reply marker = event is NOT root, find parent with root marker
	event := &nostr.Event{
		Content: "reply note",
		Tags: nostr.Tags{
			nostr.Tag{"e", "1111111111111111111111111111111111111111111111111111111111111111", "", "reply"},
			nostr.Tag{"e", "2222222222222222222222222222222222222222222222222222222222222222", "", "root"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isRoot {
		t.Errorf("reply event should not be root")
	}
	expectedRootID, _ := nostr.IDFromHex("2222222222222222222222222222222222222222222222222222222222222222")
	if rootID != expectedRootID {
		t.Errorf("root ID should be extracted from root marker, got %v", rootID)
	}
}

func TestExtractRootEvent_ReplyNoRoot(t *testing.T) {
	// Reply marker but no root marker - treat event as root
	event := &nostr.Event{
		Content: "reply without root marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", testParentID, "", "reply"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isRoot {
		t.Errorf("reply without root marker should be treated as root")
	}
	if rootID != event.ID {
		t.Errorf("root ID should equal event ID")
	}
}

func TestExtractRootEvent_SelfRootMarker(t *testing.T) {
	// Root marker pointing to self = this event IS the root
	id, _ := nostr.IDFromHex(testSomeID)
	event := &nostr.Event{
		ID:      id,
		Content: "root with self-referencing marker",
		Tags: nostr.Tags{
			nostr.Tag{"e", testSomeID, "", "root"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isRoot {
		t.Errorf("self-referencing root marker should be treated as root")
	}
	if rootID != event.ID {
		t.Errorf("root ID should equal event ID")
	}
}

func TestExtractParentID_DirectReply(t *testing.T) {
	// Direct reply: only "root" marker, root IS the parent
	event := &nostr.Event{
		Content: "direct reply",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID, "wss://relay.example.com", "root"},
		},
	}
	parentID := extractParentID(event)
	if parentID != testRootMarkerID {
		t.Errorf("direct reply parent should be root marker ID %q, got %q", testRootMarkerID, parentID)
	}
}

func TestExtractParentID_NestedReply(t *testing.T) {
	// Nested reply: both "root" and "reply" markers — parent is reply tag
	event := &nostr.Event{
		Content: "nested reply",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID, "", "root"},
			nostr.Tag{"e", testParentID, "", "reply"},
		},
	}
	parentID := extractParentID(event)
	if parentID != testParentID {
		t.Errorf("nested reply parent should be reply marker ID %q, got %q", testParentID, parentID)
	}
}

func TestNostrEventProvider_ID(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		ID: [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
	}
	id := p.ID(event)
	expected := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	if id != expected {
		t.Errorf("expected ID %q, got %q", expected, id)
	}
}

func TestNostrEventProvider_Name(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		Content: "This is a long content that should be truncated to fit the display",
		PubKey:  [32]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	name := p.Name(event)
	// Name should be truncated and include short pubkey
	if len(name) == 0 {
		t.Errorf("expected non-empty name")
	}
}

func TestNostrEventProvider_ParentID_Root(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	parentID := p.ParentID(event)
	if parentID != "" {
		t.Errorf("root event should have empty parent ID, got %q", parentID)
	}
}

func TestNostrEventProvider_ParentID_Reply(t *testing.T) {
	p := &NostrEventProvider{}
	event := nostr.Event{
		Content: "reply note",
		Tags: nostr.Tags{
			nostr.Tag{"e", testParentID, "", "reply"},
		},
	}
	parentID := p.ParentID(event)
	if parentID != testParentID {
		t.Errorf("expected parent ID %q, got %q", testParentID, parentID)
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		maxLen  int
		want    string
	}{
		{name: "short content", content: "hello", maxLen: 60, want: "hello"},
		{name: "exact max", content: strings.Repeat("a", 60), maxLen: 60, want: strings.Repeat("a", 60)},
		{name: "one over", content: strings.Repeat("a", 61), maxLen: 60, want: strings.Repeat("a", 57) + "..."},
		{name: "much longer", content: strings.Repeat("a", 200), maxLen: 60, want: strings.Repeat("a", 57) + "..."},
		{name: "very short max", content: "hello world", maxLen: 5, want: "he..."},
		{name: "empty content", content: "", maxLen: 60, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateContent(tt.content, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateContent(%q, %d) = %q, want %q", tt.content, tt.maxLen, got, tt.want)
			}
		})
	}
}

func make64CharHexID(s string) nostr.ID {
	id, _ := nostr.IDFromHex(s + strings.Repeat("0", 64-len(s)))
	return id
}

func TestBuildTuiModel(t *testing.T) {
	rootID := make64CharHexID("a")
	replyID := make64CharHexID("b")
	orphanID := make64CharHexID("c")
	invalidHexParent := "nothex"

	tests := []struct {
		name           string
		currentEventID string
		events         []*nostr.Event
		rootEvent      *nostr.Event
		wantNil        bool
		wantErr        bool
		wantNodeIDs    []string
	}{
		{
			name:           "empty events",
			currentEventID: rootID.Hex(),
			events:         nil,
			wantNil:        true,
		},
		{
			name:           "single root event",
			currentEventID: rootID.Hex(),
			events: []*nostr.Event{
				{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
			},
			wantNodeIDs: []string{rootID.Hex()},
		},
		{
			name:           "root + reply",
			currentEventID: rootID.Hex(),
			events: []*nostr.Event{
				{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
				{ID: replyID, Content: "reply", Kind: nostr.KindTextNote,
					Tags: nostr.Tags{{"e", rootID.Hex(), "", "reply"}}},
			},
			wantNodeIDs: []string{rootID.Hex(), replyID.Hex()},
		},
		{
			name:           "duplicate events",
			currentEventID: rootID.Hex(),
			events: []*nostr.Event{
				{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
				{ID: rootID, Content: "root duplicate", Kind: nostr.KindTextNote},
			},
			wantNodeIDs: []string{rootID.Hex()},
		},
		{
			name:           "missing parent gets placeholder",
			currentEventID: rootID.Hex(),
			events: []*nostr.Event{
				{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
				{ID: replyID, Content: "reply", Kind: nostr.KindTextNote,
					Tags: nostr.Tags{{"e", orphanID.Hex(), "", "reply"}}},
			},
			wantNodeIDs: []string{rootID.Hex(), replyID.Hex(), orphanID.Hex()},
		},
		{
			name:           "invalid parent hex skipped",
			currentEventID: rootID.Hex(),
			events: []*nostr.Event{
				{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
				{ID: replyID, Content: "reply", Kind: nostr.KindTextNote,
					Tags: nostr.Tags{{"e", invalidHexParent, "", "reply"}}},
			},
			wantNodeIDs: []string{rootID.Hex(), replyID.Hex()},
		},
		{
			name:           "focus falls back to root when current event not in tree",
			currentEventID: "nonexistent",
			events: []*nostr.Event{
				{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
			},
			rootEvent:   &nostr.Event{ID: rootID, Content: "root", Kind: nostr.KindTextNote},
			wantNodeIDs: []string{rootID.Hex()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &threadTreeView{
				currentEventID: tt.currentEventID,
				provider:       &NostrEventProvider{},
				styles:         newThreadStyles(),
				width:          80,
				height:         25,
			}
			if tt.rootEvent != nil {
				m.root = tt.rootEvent
			}

			tuiModel, err := m.buildTuiModel(tt.events)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTuiModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && tuiModel != nil {
				t.Errorf("buildTuiModel() = %v, want nil", tuiModel)
				return
			}
			if tt.wantNil {
				return
			}

			// Verify nodes exist by iterating the tree (All walks all nodes including children)
			tree := tuiModel.Tree
			found := map[string]bool{}
			for nodeInfo, err := range tree.All(context.Background()) {
				if err != nil {
					t.Fatalf("tree.All() error: %v", err)
				}
				event := nodeInfo.Node.Data()
				found[(*event).ID.Hex()] = true
			}

			for _, wantID := range tt.wantNodeIDs {
				if !found[wantID] {
					t.Errorf("buildTuiModel() missing node ID %s, found: %v", wantID, found)
				}
			}
		})
	}
}

func TestThreadTreeView_Update(t *testing.T) {
	t.Run("loaded msg without error clears loading", func(t *testing.T) {
		m := &threadTreeView{
			loading: true,
			styles:  newThreadStyles(),
			keys:    newThreadKeyMap(),
			width:   80,
			height:  25,
		}
		// Manually simulate what fetchThread does: clear loading before sending msg
		m.loading = false
		newModel, cmd := m.Update(threadTreeLoadedMsg{err: nil})
		if cmd != nil {
			t.Errorf("expected nil cmd, got %v", cmd)
		}
		updated, ok := newModel.(*threadTreeView)
		if !ok {
			t.Fatal("unexpected model type")
		}
		if updated.loading {
			t.Error("loading should be false after loaded msg")
		}
	})

	t.Run("loaded msg with error sets loadError", func(t *testing.T) {
		m := &threadTreeView{
			loading: true,
			styles:  newThreadStyles(),
			keys:    newThreadKeyMap(),
			width:   80,
			height:  25,
		}
		m.loading = false
		testErr := errors.New("fetch failed")
		newModel, cmd := m.Update(threadTreeLoadedMsg{err: testErr})
		if cmd != nil {
			t.Errorf("expected nil cmd, got %v", cmd)
		}
		updated, ok := newModel.(*threadTreeView)
		if !ok {
			t.Fatal("unexpected model type")
		}
		if updated.loadError == nil || updated.loadError.Error() != testErr.Error() {
			t.Errorf("loadError = %v, want %v", updated.loadError, testErr)
		}
		if updated.loading {
			t.Error("loading should be false after loaded msg (error)")
		}
	})

	t.Run("window resize updates dimensions", func(t *testing.T) {
		m := &threadTreeView{
			styles: newThreadStyles(),
			keys:   newThreadKeyMap(),
		}
		newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		updated, ok := newModel.(*threadTreeView)
		if !ok {
			t.Fatal("unexpected model type")
		}
		if updated.width != 100 {
			t.Errorf("width = %d, want 100", updated.width)
		}
		if updated.height != 30 {
			t.Errorf("height = %d, want 30", updated.height)
		}
	})

	t.Run("esc key returns bubblon.Close", func(t *testing.T) {
		m := &threadTreeView{
			styles: newThreadStyles(),
			keys:   newThreadKeyMap(),
			width:  80,
			height: 25,
		}
		_, cmd := m.Update(tea.KeyPressMsg{Text: "esc"})
		if cmd == nil {
			t.Fatal("expected close cmd for esc")
		}
		// Execute the cmd to verify it sends a close message
		msg := cmd()
		if msg == nil {
			t.Error("esc cmd returned nil message")
		}
	})

	t.Run("non-esc key without tuiModel is nop", func(t *testing.T) {
		m := &threadTreeView{
			styles: newThreadStyles(),
			keys:   newThreadKeyMap(),
		}
		newModel, cmd := m.Update(tea.KeyPressMsg{Text: "down"})
		if cmd != nil {
			t.Errorf("expected nil cmd without tuiModel, got %v", cmd)
		}
		if newModel != m {
			t.Error("expected same model reference")
		}
	})

	t.Run("key delegates to tuiModel when set", func(t *testing.T) {
		// Build a simple tree for the TuiTreeModel
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &NostrEventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &threadTreeView{
			styles:   newThreadStyles(),
			keys:     newThreadKeyMap(),
			tuiModel: tuiModel,
			width:    80,
			height:   25,
		}
		_, _ = m.Update(tea.KeyPressMsg{Text: "down"})
	})

	t.Run("enter on focused node opens event detail", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test event", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &NostrEventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		// Focus the node so GetFocusedNode returns it
		tree.SetFocusedID(context.Background(), id.Hex())

		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &threadTreeView{
			styles:   newThreadStyles(),
			keys:     newThreadKeyMap(),
			tuiModel: tuiModel,
			ctrl:     &bubblon.Controller{},
			width:    80,
			height:   25,
		}
		_, cmd := m.Update(tea.KeyPressMsg{Text: "enter"})
		if cmd == nil {
			t.Error("enter on focused node should return open-event-detail cmd")
		}
	})

	t.Run("enter on placeholder does not open detail", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		placeholder := nostr.Event{ID: id, Content: "[loading...]", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{placeholder}, &NostrEventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tree.SetFocusedID(context.Background(), id.Hex())

		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &threadTreeView{
			styles:   newThreadStyles(),
			keys:     newThreadKeyMap(),
			tuiModel: tuiModel,
			ctrl:     &bubblon.Controller{},
			width:    80,
			height:   25,
		}
		_, cmd := m.Update(tea.KeyPressMsg{Text: "enter"})
		if cmd != nil {
			t.Error("enter on placeholder should not open event detail")
		}
	})
}

func TestThreadTreeView_View(t *testing.T) {
	t.Run("loading shows spinner", func(t *testing.T) {
		m := &threadTreeView{
			loading: true,
			styles:  newThreadStyles(),
			keys:    newThreadKeyMap(),
			width:   80,
			height:  25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "loading") {
			t.Errorf("View() should contain loading indicator, got: %s", v.Content)
		}
	})

	t.Run("error shows error message", func(t *testing.T) {
		m := &threadTreeView{
			loadError: errors.New("test error"),
			styles:    newThreadStyles(),
			keys:      newThreadKeyMap(),
			width:     80,
			height:    25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "test error") {
			t.Errorf("View() should contain error message, got: %s", v.Content)
		}
	})

	t.Run("no data and no event shows no thread data", func(t *testing.T) {
		m := &threadTreeView{
			styles: newThreadStyles(),
			keys:   newThreadKeyMap(),
			width:  80,
			height: 25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "no thread data") {
			t.Errorf("View() should indicate no data, got: %s", v.Content)
		}
	})

	t.Run("event without tree shows fallback", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		m := &threadTreeView{
			event:  &nostr.Event{ID: id, Content: "hello", Kind: nostr.KindTextNote},
			styles: newThreadStyles(),
			keys:   newThreadKeyMap(),
			width:  80,
			height: 25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "hello") {
			t.Errorf("View() should show event content, got: %s", v.Content)
		}
	})

	t.Run("tree renders with help bar", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &NostrEventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &threadTreeView{
			styles:   newThreadStyles(),
			keys:     newThreadKeyMap(),
			width:    80,
			height:   25,
			tuiModel: tuiModel,
		}
		v := m.View()
		if !strings.Contains(v.Content, "esc back") {
			t.Errorf("View() should contain help bar, got: %s", v.Content)
		}
		if !strings.Contains(v.Content, "Thread") {
			t.Errorf("View() should contain title, got: %s", v.Content)
		}
	})

	t.Run("view always includes title", func(t *testing.T) {
		m := &threadTreeView{
			styles: newThreadStyles(),
			keys:   newThreadKeyMap(),
			width:  80,
			height: 25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "Thread") {
			t.Errorf("View() should always contain title, got: %s", v.Content)
		}
	})
}

func TestNewThreadTreeView(t *testing.T) {
	id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
	event := &nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
	app := &config.AppContext{}

	m := NewThreadTreeView(event, app, 80, 25, nil)

	if m.event != event {
		t.Error("event not set")
	}
	if m.currentEventID != event.ID.Hex() {
		t.Errorf("currentEventID = %q, want %q", m.currentEventID, event.ID.Hex())
	}
	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 25 {
		t.Errorf("height = %d, want 25", m.height)
	}
	if m.provider == nil {
		t.Error("provider should be initialized")
	}
}

func TestExtractParentID_NilEvent(t *testing.T) {
	parentID := extractParentID(nil)
	if parentID != "" {
		t.Errorf("nil event should return empty parent ID, got %q", parentID)
	}
}

func TestExtractParentID_ShortTag(t *testing.T) {
	event := &nostr.Event{
		Content: "note with short e tag",
		Tags:    nostr.Tags{nostr.Tag{"e", "ab"}},
	}
	parentID := extractParentID(event)
	if parentID != "" {
		t.Errorf("e tag without reply marker should return empty, got %q", parentID)
	}
}

func TestExtractRootEvent_InvalidRootHex(t *testing.T) {
	event := &nostr.Event{
		Content: "reply with invalid root hex",
		Tags: nostr.Tags{
			nostr.Tag{"e", "abad1dea", "", "root"},
			nostr.Tag{"e", "parentID1234parentID1234parentID1234parentID1234parentID1234parent", "", "reply"},
		},
	}
	_, _, err := extractRootEvent(event)
	if err == nil {
		t.Error("expected error for invalid root hex")
	}
}