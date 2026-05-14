package utils

import (
	"fmt"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

func ParsePubKey(s string) (nostr.PubKey, error) {
	prefix, decoded, err := nip19.Decode(s)
	if err == nil {
		switch prefix {
		case "npub":
			if pk, ok := decoded.(nostr.PubKey); ok {
				return pk, nil
			}
			return nostr.PubKey{}, fmt.Errorf("invalid npub format")
		}
	}

	if len(s) == 66 && (s[:2] == "02" || s[:2] == "03") {
		return nostr.PubKeyFromHex(s[2:])
	}

	if len(s) == 64 {
		return nostr.PubKeyFromHex(s)
	}

	return nostr.PubKey{}, fmt.Errorf("unknown pubkey format")
}
