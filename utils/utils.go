package utils

import (
	"encoding/hex"
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
		xHex := s[2:]
		xBytes, err := hex.DecodeString(xHex)
		if err != nil {
			return nostr.PubKey{}, fmt.Errorf("invalid compressed pubkey hex: %w", err)
		}
		if len(xBytes) != 32 {
			return nostr.PubKey{}, fmt.Errorf("invalid x coordinate length: %d", len(xBytes))
		}
		var pk nostr.PubKey
		copy(pk[:], xBytes)
		return pk, nil
	}

	if len(s) == 64 {
		xBytes, err := hex.DecodeString(s)
		if err != nil {
			return nostr.PubKey{}, fmt.Errorf("invalid hex: %w", err)
		}
		var pk nostr.PubKey
		copy(pk[:], xBytes)
		return pk, nil
	}

	return nostr.PubKey{}, fmt.Errorf("unknown pubkey format")
}
