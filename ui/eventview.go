package ui

import (
	"fmt"
	"time"

	"fiatjaf.com/nostr"
	"github.com/rivo/tview"
)

// EventView is a tview component that displays a Nostr event in a card-like layout.
// Layout:
// - Top-left: User name (pubkey shortened)
// - Top-right: Time (relative or formatted)
// - Middle: Event content
// - Bottom: Stats (likes, replies counts)
type EventView struct {
	*tview.Flex
	event *nostr.Event

	// Child components
	headerFlex *tview.Flex
	userName   *tview.TextView
	timestamp  *tview.TextView
	content    *tview.TextView
	statsFlex  *tview.Flex
	likes      *tview.TextView
	replies    *tview.TextView
	reposts    *tview.TextView
}

// NewEventView creates a new EventView component.
func NewEventView() *EventView {
	ev := &EventView{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	// Create child components
	ev.userName = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText("Unknown User")

	ev.timestamp = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetText("just now")

	ev.content = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true).
		SetText("No content")

	ev.likes = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("♥ 0")

	ev.replies = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("↩ 0")

	ev.reposts = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("↻ 0")

	// Build header (user name left, timestamp right)
	ev.headerFlex = tview.NewFlex().
		AddItem(ev.userName, 0, 1, false).
		AddItem(ev.timestamp, 0, 1, false)

	// Build stats row
	ev.statsFlex = tview.NewFlex().
		AddItem(ev.likes, 0, 1, false).
		AddItem(ev.replies, 0, 1, false).
		AddItem(ev.reposts, 0, 1, false)

	// Assemble main layout
	ev.Flex.
		AddItem(ev.headerFlex, 1, 0, false). // Header row
		AddItem(ev.content, 0, 1, false).    // Content (expands)
		AddItem(ev.statsFlex, 1, 0, false)   // Stats row

	// Set border for card appearance
	ev.Flex.SetBorder(true).SetTitle("Nostr Event")

	return ev
}

// SetEvent updates the view with a new Nostr event.
func (ev *EventView) SetEvent(event *nostr.Event) {
	ev.event = event
	ev.updateView()
}

// updateView refreshes all display elements based on the current event.
func (ev *EventView) updateView() {
	if ev.event == nil {
		return
	}

	// User name (shortened pubkey)
	pubkeyHex := ev.event.PubKey.Hex()
	if len(pubkeyHex) > 16 {
		pubkeyHex = pubkeyHex[:8] + "…" + pubkeyHex[len(pubkeyHex)-8:]
	}
	ev.userName.SetText("[blue]" + pubkeyHex)

	// Timestamp
	t := time.Unix(int64(ev.event.CreatedAt), 0)
	now := time.Now()
	diff := now.Sub(t)

	var timeText string
	switch {
	case diff < time.Minute:
		timeText = "just now"
	case diff < time.Hour:
		timeText = fmt.Sprintf("%dm", int(diff.Minutes()))
	case diff < 24*time.Hour:
		timeText = fmt.Sprintf("%dh", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		timeText = fmt.Sprintf("%dd", int(diff.Hours()/24))
	default:
		timeText = t.Format("2006-01-02")
	}
	ev.timestamp.SetText("[gray]" + timeText)

	// Content
	ev.content.SetText(ev.event.Content)

	// Stats (placeholder - in real app, these would come from event tags)
	// For now, we'll show some placeholder stats
	ev.likes.SetText("[red]♥ 0")
	ev.replies.SetText("[yellow]↩ 0")
	ev.reposts.SetText("[green]↻ 0")
}

// GetEvent returns the currently displayed event.
func (ev *EventView) GetEvent() *nostr.Event {
	return ev.event
}

// SetBorder enables or disables the border around the event card.
func (ev *EventView) SetBorder(show bool) *EventView {
	ev.Flex.SetBorder(show)
	return ev
}

// SetTitle sets the title of the event card border.
func (ev *EventView) SetTitle(title string) *EventView {
	ev.Flex.SetTitle(title)
	return ev
}
