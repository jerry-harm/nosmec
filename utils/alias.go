package utils

import (
	"fmt"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

func ResolveAlias(app *config.AppContext, identifier string) (string, error) {
	cfg := app.Config()

	if cfg.Alias != nil {
		if value, ok := cfg.Alias[identifier]; ok {
			return value, nil
		}
	}

	return identifier, nil
}

func ResolveAliasToPubKey(app *config.AppContext, identifier string) (nostr.PubKey, error) {
	resolved, err := ResolveAlias(app, identifier)
	if err != nil {
		return nostr.PubKey{}, err
	}

	var pubKey nostr.PubKey
	if len(resolved) == 64 {
		for _, c := range resolved {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return nostr.PubKey{}, fmt.Errorf("invalid hex pubkey: contains non-hex characters")
			}
		}
		copy(pubKey[:], resolved)
		return pubKey, nil
	}

	prefix, decoded, err := nip19.Decode(resolved)
	if err == nil {
		if prefix == "npub" {
			if pk, ok := decoded.(nostr.PubKey); ok {
				return pk, nil
			}
		}
	}

	return nostr.PubKey{}, fmt.Errorf("invalid user identifier, expected alias, npub or 64-character hex pubkey")
}

func ListAliases(app *config.AppContext) map[string]string {
	cfg := app.Config()
	if cfg.Alias == nil {
		return make(map[string]string)
	}
	return cfg.Alias
}
