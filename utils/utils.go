package utils

import (
	"fmt"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

func ParsePubKey(s string) (nostr.PubKey, error) {
	prefix, decoded, err := nip19.Decode(s)
	if err != nil {
		return nostr.PubKey{}, fmt.Errorf("invalid pubkey: %w", err)
	}

	switch prefix {
	case "npub":
		if pk, ok := decoded.(nostr.PubKey); ok {
			return pk, nil
		}
		return nostr.PubKey{}, fmt.Errorf("invalid npub format")
	default:
		if len(s) == 64 {
			var pk nostr.PubKey
			copy(pk[:], s)
			return pk, nil
		}
		return nostr.PubKey{}, fmt.Errorf("unknown prefix: %s", prefix)
	}
}
