package timeline

import (
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/utils"
)

func TestFetchMoreOld_EmptyListReturnsError(t *testing.T) {
	// Create a minimal list with items
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetItems([]list.Item{})

	m := &model{
		list:          l,
		isLoadingMore: false,
	}

	cmd := m.fetchMoreOld()
	result := cmd()

	errMsg, ok := result.(loadMoreErrorMsg)
	if !ok {
		t.Fatalf("expected loadMoreErrorMsg, got %T", result)
	}
	if errMsg.isNew {
		t.Errorf("expected isNew=false for empty list error")
	}
}

func TestFetchMoreOld_WithItemsDoesNotPanic(t *testing.T) {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetItems([]list.Item{
		item{event: utils.TimelineEvent{Event: nostr.Event{CreatedAt: 1234567890}}},
	})

	m := &model{
		list:          l,
		isLoadingMore: false,
	}

	cmd := m.fetchMoreOld()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	_ = cmd
}

func TestUpdate_ListUpdateWithEmptyItems(t *testing.T) {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetItems([]list.Item{})

	m := &model{}
	m.list = l

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	_, _ = m.Update(msg)
}

func TestListItem_EmptyItems(t *testing.T) {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetItems([]list.Item{})

	items := l.Items()
	if len(items) != 0 {
		t.Errorf("len(items) = %d, want 0", len(items))
	}
}