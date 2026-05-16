package utils

import (
	"testing"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
	"github.com/jerry-harm/nosmec/config"
)

func TestResolveAliasToPubKey_Hex(t *testing.T) {
	hexPk := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	tests := []struct {
		name       string
		identifier string
		wantPk     string
		wantErr    bool
	}{
		{
			name:       "64 char hex pubkey",
			identifier: hexPk,
			wantPk:     hexPk,
			wantErr:    false,
		},
		{
			name:       "02 prefix hex (33 bytes)",
			identifier: "02" + hexPk,
			wantPk:     hexPk,
			wantErr:    false,
		},
		{
			name:       "03 prefix hex (33 bytes)",
			identifier: "03" + hexPk,
			wantPk:     hexPk,
			wantErr:    false,
		},
		{
			name:       "invalid identifier too long",
			identifier: "notavalidentifiertonotavalidentifiertonotavalidentifiertonotavalidentifier",
			wantPk:     "",
			wantErr:    true,
		},
		{
			name:       "too short hex",
			identifier: "abc123",
			wantPk:     "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{}
			app := config.NewAppContext(nil, nil, cfg, nil)
			gotPk, err := ResolveAliasToPubKey(app, tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAliasToPubKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotPk.Hex() != tt.wantPk {
				t.Errorf("ResolveAliasToPubKey() = %v, want %v", gotPk.Hex(), tt.wantPk)
			}
		})
	}
}

func TestResolveAliasToPubKey_InvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantErr    bool
	}{
		{
			name:       "empty string",
			identifier: "",
			wantErr:    true,
		},
		{
			name:       "random string",
			identifier: "randomstring",
			wantErr:    true,
		},
		{
			name:       "npub invalid",
			identifier: "npub1invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{}
			app := config.NewAppContext(nil, nil, cfg, nil)
			_, err := ResolveAliasToPubKey(app, tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAliasToPubKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveAliasToPubKey_ViaAlias(t *testing.T) {
	// Create a valid npub from a known hex pubkey
	hexPk := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	// First create a valid npub to use as alias
	pk, _ := nostr.PubKeyFromHex(hexPk)
	npub := nip19.EncodeNpub(pk)

	tests := []struct {
		name       string
		identifier string
		aliasMap   map[string]string
		wantPk     string
		wantErr    bool
	}{
		{
			name:       "npub via alias",
			identifier: "mykey",
			aliasMap:   map[string]string{"mykey": npub},
			wantPk:     hexPk,
			wantErr:    false,
		},
		{
			name:       "hex via alias",
			identifier: "otherkey",
			aliasMap:   map[string]string{"otherkey": hexPk},
			wantPk:     hexPk,
			wantErr:    false,
		},
		{
			name:       "unknown alias returns identity",
			identifier: "unknownkey",
			aliasMap:   map[string]string{"somekey": npub},
			wantPk:     "unknownkey", // Will fail since it's not valid hex
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{Alias: tt.aliasMap}
			app := config.NewAppContext(nil, nil, cfg, nil)
			gotPk, err := ResolveAliasToPubKey(app, tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAliasToPubKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotPk.Hex() != tt.wantPk {
				t.Errorf("ResolveAliasToPubKey() = %v, want %v", gotPk.Hex(), tt.wantPk)
			}
		})
	}
}

func TestPubKeyToNpub(t *testing.T) {
	tests := []struct {
		name string
		pk   nostr.PubKey
	}{
		{
			name: "zero pubkey",
			pk:   nostr.PubKey{},
		},
		{
			name: "non-zero pubkey",
			pk: func() nostr.PubKey {
				pk, _ := nostr.PubKeyFromHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
				return pk
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PubKeyToNpub(tt.pk)
			if got == "" {
				t.Error("PubKeyToNpub() returned empty string")
			}
			// Verify it can be decoded back
			prefix, decoded, err := nip19.Decode(got)
			if err != nil {
				t.Errorf("PubKeyToNpub() output %q is not decodable: %v", got, err)
			}
			if prefix != "npub" {
				t.Errorf("PubKeyToNpub() prefix = %q, want npub", prefix)
			}
			decodedPk, ok := decoded.(nostr.PubKey)
			if !ok {
				t.Errorf("PubKeyToNpub() decoded type is not PubKey")
			}
			if decodedPk != tt.pk {
				t.Errorf("PubKeyToNpub() round-trip failed: got %v, want %v", decodedPk, tt.pk)
			}
		})
	}
}

func TestListAliases(t *testing.T) {
	tests := []struct {
		name     string
		aliasMap map[string]string
		wantLen  int
	}{
		{
			name:     "nil alias",
			aliasMap: nil,
			wantLen:  0,
		},
		{
			name:     "empty alias",
			aliasMap: map[string]string{},
			wantLen:  0,
		},
		{
			name:     "one alias",
			aliasMap: map[string]string{"alice": "npub1..."},
			wantLen:  1,
		},
		{
			name:     "multiple aliases",
			aliasMap: map[string]string{"alice": "npub1...", "bob": "npub2..."},
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{Alias: tt.aliasMap}
			app := config.NewAppContext(nil, nil, cfg, nil)
			got := ListAliases(app)
			if len(got) != tt.wantLen {
				t.Errorf("ListAliases() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}