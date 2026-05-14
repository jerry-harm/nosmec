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

	prefix, decoded, err := nip19.Decode(resolved)
	if err == nil {
		if prefix == "npub" {
			if pk, ok := decoded.(nostr.PubKey); ok {
				return pk, nil
			}
		}
	}

	if len(resolved) == 66 && (resolved[:2] == "02" || resolved[:2] == "03") {
		return nostr.PubKeyFromHex(resolved[2:])
	}

	if len(resolved) == 64 {
		return nostr.PubKeyFromHex(resolved)
	}

	return nostr.PubKey{}, fmt.Errorf("invalid user identifier, expected alias, npub or 64-character hex pubkey")
}

func PubKeyToNpub(pk nostr.PubKey) string {
	return nip19.EncodeNpub(pk)
}

func ListAliases(app *config.AppContext) map[string]string {
	cfg := app.Config()
	if cfg.Alias == nil {
		return make(map[string]string)
	}
	return cfg.Alias
}
