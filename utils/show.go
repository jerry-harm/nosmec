package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/fatih/color"
)

func PrintEvent(ev *nostr.Event, j bool) {
	if j {
		json.NewEncoder(os.Stdout).Encode(ev)
		return
	}

	color.Set(color.FgHiBlue)
	nevent := nip19.EncodeNevent(ev.ID, nil, ev.PubKey)
	fmt.Println(nevent)
	color.Set(color.Reset)

	fmt.Print(ev.CreatedAt.Time().Format("2006-01-02T15:04:05") + "\n")

	color.Set(color.FgRed)
	npub := nip19.EncodeNpub(ev.PubKey)
	fmt.Println(npub)
	color.Set(color.Reset)

	fmt.Println(ev.Content)
	fmt.Println()
}
