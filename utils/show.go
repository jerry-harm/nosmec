package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/fatih/color"
)

// PrintEvent 打印 Nostr 事件
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

	ctx := context.Background()
	if profile, err := GetProfile(ctx, ev.PubKey); err == nil {
		var profileData map[string]interface{}
		if err := json.Unmarshal([]byte(profile.Content), &profileData); err == nil {
			if name, ok := profileData["name"].(string); ok && name != "" {
				color.Set(color.FgHiRed)
				fmt.Print(name + " ")
				color.Set(color.Reset)
			}
		}
	}

	color.Set(color.FgRed)
	npub := nip19.EncodeNpub(ev.PubKey)
	fmt.Println(npub)
	color.Set(color.Reset)

	// 打印内容
	fmt.Println(ev.Content)
	fmt.Println()
}
