package thread

import (
	"context"
	"strings"
	"testing"

	"github.com/Digital-Shane/treeview/v2"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/tui/component/bubblon"
	"github.com/jerry-harm/nosmec/tui/theme"
)

var (
	testParentID      = "0000000000000000000000000000000000000000000000000000000000000001"
	testFirstParentID = "0000000000000000000000000000000000000000000000000000000000000002"
	testSecondParentID = "0000000000000000000000000000000000000000000000000000000000000003"
	testSomeID        = "0000000000000000000000000000000000000000000000000000000000000004"
	testRootMarkerID  = "abcd000000000000000000000000000000000000000000000000000000000000"
)

func TestExtractParentID_RootEvent(t *testing.T) {
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
	event := &nostr.Event{
		Content: "direct reply to root",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID, "", "root"},
		},
	}
	parentID := extractParentID(event)
	if event.ID.Hex() != testRootMarkerID {
		if parentID != testRootMarkerID {
			t.Errorf("direct reply should have parent = root marker ID %q, got %q", testRootMarkerID, parentID)
		}
	} else {
		if parentID != "" {
			t.Errorf("self-root event should have empty parent ID, got %q", parentID)
		}
	}
}

func TestExtractParentID_ReplyMarker(t *testing.T) {
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

func TestExtractParentID_PositionalSingleTag(t *testing.T) {
	event := &nostr.Event{
		Content: "note with single positional e tag",
		Tags: nostr.Tags{
			nostr.Tag{"e", testSomeID},
		},
	}
	parentID := extractParentID(event)
	if parentID != testSomeID {
		t.Errorf("single positional e tag should be parent, got %q", parentID)
	}
}

func TestExtractParentID_PositionalTwoTags(t *testing.T) {
	event := &nostr.Event{
		Content: "note with two positional e tags",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID},
			nostr.Tag{"e", testParentID},
		},
	}
	parentID := extractParentID(event)
	if parentID != testParentID {
		t.Errorf("last positional e tag should be parent, got %q (want %q)", parentID, testParentID)
	}
}

func TestExtractParentID_NoMarker(t *testing.T) {
	event := &nostr.Event{
		Content: "note with e tag but no valid hex",
		Tags: nostr.Tags{
			nostr.Tag{"e", "invalid"},
		},
	}
	parentID := extractParentID(event)
	want := "0000000000000000000000000000000000000000000000000000000000000000"
	if parentID != want {
		t.Errorf("nip10 normalizes invalid hex to zero ID, got %q, want %q", parentID, want)
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
	if isRoot {
		t.Errorf("nip10: reply without root marker means first e tag is thread root, not self")
	}
	expectedRootID, _ := nostr.IDFromHex(testParentID)
	if rootID != expectedRootID {
		t.Errorf("root ID should be first e tag per NIP-10, got %v", rootID)
	}
}

func TestExtractRootEvent_SelfRootMarker(t *testing.T) {
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

func TestExtractRootEvent_PositionalTwoTags(t *testing.T) {
	event := &nostr.Event{
		Content: "note with positional e tags",
		Tags: nostr.Tags{
			nostr.Tag{"e", testRootMarkerID},
			nostr.Tag{"e", testParentID},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isRoot {
		t.Errorf("event with two positional e tags should not be root")
	}
	expectedRootID, _ := nostr.IDFromHex(testRootMarkerID)
	if rootID != expectedRootID {
		t.Errorf("first positional e tag should be root, got %v", rootID)
	}
}

func TestExtractParentID_DirectReply(t *testing.T) {
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

func TestEventProvider_ID(t *testing.T) {
	p := &eventProvider{}
	event := nostr.Event{
		ID: [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
	}
	id := p.ID(event)
	expected := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	if id != expected {
		t.Errorf("expected ID %q, got %q", expected, id)
	}
}

func TestEventProvider_Name(t *testing.T) {
	p := &eventProvider{}
	event := nostr.Event{
		Content: "This is a long content that should be truncated to fit the display",
		PubKey:  [32]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	name := p.Name(event)
	if len(name) == 0 {
		t.Errorf("expected non-empty name")
	}
}

func TestEventProvider_ParentID_Root(t *testing.T) {
	p := &eventProvider{}
	event := nostr.Event{
		Content: "root note",
		Tags:    nostr.Tags{},
	}
	parentID := p.ParentID(event)
	if parentID != "" {
		t.Errorf("root event should have empty parent ID, got %q", parentID)
	}
}

func TestEventProvider_ParentID_Reply(t *testing.T) {
	p := &eventProvider{}
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
			m := &Model{
				currentEventID: tt.currentEventID,
				provider:       &eventProvider{},
				styles:         newStyles(theme.DefaultTheme(false)),
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

func TestBuildInitialTree(t *testing.T) {
	t.Run("root event builds self as tree", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := &nostr.Event{
			ID:      id,
			Content: "hello world",
			Kind:    nostr.KindTextNote,
			Tags:    nostr.Tags{},
		}
		m := &Model{
			event:          event,
			currentEventID: id.Hex(),
			provider:       &eventProvider{},
			styles:         newStyles(theme.DefaultTheme(false)),
			width:          80,
			height:         25,
		}
		m.buildInitialTree()

		if m.tuiModel == nil {
			t.Fatal("expected non-nil tuiModel")
		}

		v := m.View()
		if !strings.Contains(v.Content, "hello world") {
			t.Errorf("View should contain event content, got: %s", v.Content)
		}
	})

	t.Run("direct reply builds root placeholder + event", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("b", 64))
		rootID, _ := nostr.IDFromHex(testRootMarkerID)
		event := &nostr.Event{
			ID:      id,
			Content: "direct reply",
			Kind:    nostr.KindTextNote,
			Tags: nostr.Tags{
				{"e", testRootMarkerID, "", "root"},
			},
		}
		m := &Model{
			event:          event,
			currentEventID: id.Hex(),
			provider:       &eventProvider{},
			styles:         newStyles(theme.DefaultTheme(false)),
			width:          80,
			height:         25,
		}
		m.buildInitialTree()

		if m.tuiModel == nil {
			t.Fatal("expected non-nil tuiModel")
		}

		tree := m.tuiModel.Tree
		found := map[string]bool{}
		for nodeInfo, err := range tree.All(context.Background()) {
			if err != nil {
				t.Fatal(err)
			}
			ev := nodeInfo.Node.Data()
			found[(*ev).ID.Hex()] = true
		}
		if !found[id.Hex()] {
			t.Error("missing current event")
		}
		if !found[rootID.Hex()] {
			t.Error("missing root placeholder")
		}
	})

	t.Run("nested reply builds root + parent placeholders + event", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("c", 64))
		event := &nostr.Event{
			ID:      id,
			Content: "nested reply",
			Kind:    nostr.KindTextNote,
			Tags: nostr.Tags{
				{"e", testRootMarkerID, "", "root"},
				{"e", testParentID, "", "reply"},
			},
		}
		m := &Model{
			event:          event,
			currentEventID: id.Hex(),
			provider:       &eventProvider{},
			styles:         newStyles(theme.DefaultTheme(false)),
			width:          80,
			height:         25,
		}
		m.buildInitialTree()

		if m.tuiModel == nil {
			t.Fatal("expected non-nil tuiModel")
		}

		tree := m.tuiModel.Tree
		found := map[string]bool{}
		for nodeInfo, err := range tree.All(context.Background()) {
			if err != nil {
				t.Fatal(err)
			}
			ev := nodeInfo.Node.Data()
			found[(*ev).ID.Hex()] = true
		}
		if !found[id.Hex()] {
			t.Error("missing current event")
		}
		if !found[testRootMarkerID] {
			t.Error("missing root placeholder")
		}
		if !found[testParentID] {
			t.Error("missing parent placeholder")
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("loaded msg triggers rerender", func(t *testing.T) {
		m := &Model{
			styles: newStyles(theme.DefaultTheme(false)),
			keys:   newKeyMap(),
			width:  80,
			height: 25,
		}
		newModel, cmd := m.Update(loadedMsg{err: nil})
		if cmd != nil {
			t.Errorf("expected nil cmd, got %v", cmd)
		}
		if _, ok := newModel.(*Model); !ok {
			t.Fatal("unexpected model type")
		}
	})

	t.Run("window resize updates dimensions", func(t *testing.T) {
		m := &Model{
			styles: newStyles(theme.DefaultTheme(false)),
			keys:   newKeyMap(),
		}
		newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		updated, ok := newModel.(*Model)
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
		m := &Model{
			styles: newStyles(theme.DefaultTheme(false)),
			keys:   newKeyMap(),
			width:  80,
			height: 25,
		}
		_, cmd := m.Update(tea.KeyPressMsg{Text: "esc"})
		if cmd == nil {
			t.Fatal("expected close cmd for esc")
		}
		msg := cmd()
		if msg == nil {
			t.Error("esc cmd returned nil message")
		}
	})

	t.Run("non-esc key without tuiModel is nop", func(t *testing.T) {
		m := &Model{
			styles: newStyles(theme.DefaultTheme(false)),
			keys:   newKeyMap(),
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
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &eventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &Model{
			styles:   newStyles(theme.DefaultTheme(false)),
			keys:     newKeyMap(),
			tuiModel: tuiModel,
			width:    80,
			height:   25,
		}
		_, _ = m.Update(tea.KeyPressMsg{Text: "down"})
	})

	t.Run("enter on focused node opens event detail", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test event", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &eventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tree.SetFocusedID(context.Background(), id.Hex())

		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		called := false
		m := &Model{
			styles:   newStyles(theme.DefaultTheme(false)),
			keys:     newKeyMap(),
			tuiModel: tuiModel,
			ctrl:     &bubblon.Controller{},
			width:    80,
			height:   25,
			newEventView: func(ev *nostr.Event) tea.Model {
				called = true
				return nil
			},
		}
		_, cmd := m.Update(tea.KeyPressMsg{Text: "enter"})
		if cmd == nil || !called {
			t.Error("enter on focused node should call newEventView")
		}
	})

	t.Run("enter on placeholder does not open detail", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		placeholder := nostr.Event{ID: id, Content: "[...]", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{placeholder}, &eventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tree.SetFocusedID(context.Background(), id.Hex())

		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &Model{
			styles:   newStyles(theme.DefaultTheme(false)),
			keys:     newKeyMap(),
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

func TestView(t *testing.T) {
	t.Run("no data shows no thread data", func(t *testing.T) {
		m := &Model{
			styles: newStyles(theme.DefaultTheme(false)),
			keys:   newKeyMap(),
			width:  80,
			height: 25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "no thread data") {
			t.Errorf("View() should indicate no data, got: %s", v.Content)
		}
	})

	t.Run("tree renders with help bar", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &eventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &Model{
			styles:   newStyles(theme.DefaultTheme(false)),
			keys:     newKeyMap(),
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

	t.Run("unfetched shows fetching indicator in title", func(t *testing.T) {
		id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
		event := nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
		tree, err := treeview.NewTreeFromFlatData(context.Background(), []nostr.Event{event}, &eventProvider{})
		if err != nil {
			t.Fatal(err)
		}
		tuiModel := treeview.NewTuiTreeModel(tree,
			treeview.WithTuiWidth[nostr.Event](80),
			treeview.WithTuiHeight[nostr.Event](20),
		)

		m := &Model{
			styles:   newStyles(theme.DefaultTheme(false)),
			keys:     newKeyMap(),
			width:    80,
			height:   25,
			tuiModel: tuiModel,
			fetched:  false,
		}
		v := m.View()
		if !strings.Contains(v.Content, "fetching") {
			t.Errorf("View() should contain fetching indicator, got: %s", v.Content)
		}
	})

	t.Run("view always includes title", func(t *testing.T) {
		m := &Model{
			styles: newStyles(theme.DefaultTheme(false)),
			keys:   newKeyMap(),
			width:  80,
			height: 25,
		}
		v := m.View()
		if !strings.Contains(v.Content, "Thread") {
			t.Errorf("View() should always contain title, got: %s", v.Content)
		}
	})
}

func TestNew(t *testing.T) {
	id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
	event := &nostr.Event{ID: id, Content: "test", Kind: nostr.KindTextNote}
	app := &config.AppContext{}

	m := New(event, app, 80, 25, nil, nil)

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

func TestExtractParentID_PositionalFirstTagIsSelf(t *testing.T) {
	id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
	event := &nostr.Event{
		ID:      id,
		Content: "note",
		Tags: nostr.Tags{
			{"e", id.Hex()},
			{"e", testSomeID},
		},
	}
	parentID := extractParentID(event)
	if parentID != testSomeID {
		t.Errorf("last positional e tag should be parent when first is self, got %q", parentID)
	}
}

func TestExtractRootEvent_PositionalSingleTag(t *testing.T) {
	id, _ := nostr.IDFromHex(strings.Repeat("a", 64))
	someID, _ := nostr.IDFromHex(testSomeID)
	event := &nostr.Event{
		ID:      id,
		Content: "note",
		Tags: nostr.Tags{
			{"e", testSomeID},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isRoot {
		t.Errorf("single positional e tag should NOT be root (it's a reply to that event)")
	}
	if rootID != someID {
		t.Errorf("provisional root should be the parent event, got %v, want %v", rootID, someID)
	}
}

func TestExtractRootEvent_InvalidRootHex(t *testing.T) {
	event := &nostr.Event{
		Content: "reply with invalid root hex",
		Tags: nostr.Tags{
			nostr.Tag{"e", "abad1dea", "", "root"},
			nostr.Tag{"e", testParentID, "", "reply"},
		},
	}
	rootID, isRoot, err := extractRootEvent(event)
	if err != nil {
		t.Errorf("nip10 does not error on invalid hex, it normalizes: %v", err)
	}
	if rootID != (nostr.ID{}) {
		t.Errorf("nip10: invalid hex normalizes to zero ID, got %v", rootID)
	}
	if !isRoot {
		t.Errorf("nip10: event with zero root ID (invalid hex) and empty event.ID is treated as root")
	}
}
