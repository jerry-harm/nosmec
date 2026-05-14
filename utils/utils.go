package utils

import (
	"encoding/hex"
	"fmt"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/btcsuite/btcd/btcec/v2"
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
		compressedBytes, err := hex.DecodeString(s)
		if err != nil {
			return nostr.PubKey{}, fmt.Errorf("invalid compressed pubkey hex: %w", err)
		}
		pk, err := btcec.ParsePubKey(compressedBytes)
		if err != nil {
			return nostr.PubKey{}, fmt.Errorf("invalid compressed pubkey: %w", err)
		}
		var nostrPK nostr.PubKey
		copy(nostrPK[:], pk.X().Bytes())
		return nostrPK, nil
	}

	if len(s) == 64 {
		var pk nostr.PubKey
		copy(pk[:], s)
		return pk, nil
	}

	return nostr.PubKey{}, fmt.Errorf("unknown pubkey format")
}
