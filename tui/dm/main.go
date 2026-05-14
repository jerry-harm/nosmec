package dm

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

func RunDM(app *config.AppContext, npubOrHex string) error {
	_, decoded, err := nip19.Decode(npubOrHex)
	if err != nil {
		return fmt.Errorf("invalid npub: %w", err)
	}
	recipientPubKey, ok := decoded.(nostr.PubKey)
	if !ok {
		return fmt.Errorf("invalid npub format")
	}

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	m := NewModel(app, recipientPubKey)
	_, err = tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running DM TUI:", err)
		os.Exit(1)
	}
	return nil
}