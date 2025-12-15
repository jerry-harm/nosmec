package ui_test

import (
	"testing"
	"time"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/ui"
	"github.com/rivo/tview"
)

func TestView(t *testing.T) {
	app := tview.NewApplication()

	// Create a sample event
	// Use MustIDFromHex to create a valid ID (using a dummy hex string)
	id := nostr.MustIDFromHex("0000000000000000000000000000000000000000000000000000000000000001")
	pubkey := nostr.MustPubKeyFromHex("0000000000000000000000000000000000000000000000000000000000000002")

	event := &nostr.Event{
		ID:        id,
		PubKey:    pubkey,
		CreatedAt: nostr.Timestamp(time.Now().Unix() - 3600), // 1 hour ago
		Kind:      1,
		Content:   "Hello Nostr! This is a sample event showing how the EventView component works. It should display nicely in the terminal UI with proper formatting and layout.",
		Tags:      nostr.Tags{},
		Sig:       [64]byte{},
	}

	// Create the event view
	eventView := ui.NewEventView()
	eventView.SetEvent(event)
	eventView.SetTitle("Sample Nostr Event")

	// Center the event view on screen
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(eventView, 0, 1, false).
			AddItem(nil, 0, 1, false), 0, 1, false).
		AddItem(nil, 0, 1, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
