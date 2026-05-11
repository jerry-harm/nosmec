# Research: NIP-19 encoding in fiatjaf.com/nostr SDK

- **Query**: NIP-19 encoding functions in fiatjaf.com/nostr/nip19
- **Scope**: external (SDK docs + codebase usage)
- **Date**: 2026-05-11

## Findings

### nip19 Package Functions

The `fiatjaf.com/nostr/nip19` package provides the following encoding functions:

| Function | Signature | Description |
|---|---|---|
| `EncodeNpub` | `func EncodeNpub(pk nostr.PubKey) string` | Encode a PubKey to `npub1...` format |
| `EncodeNsec` | `func EncodeNsec(sk [32]byte) string` | Encode a SecretKey (raw `[32]byte`) to `nsec1...` format |
| `EncodeNevent` | `func EncodeNevent(id nostr.ID, relays []string, author nostr.PubKey) string` | Encode Event ID + relay hints + author to `nevent1...` format |
| `EncodeNprofile` | `func EncodeNprofile(pk nostr.PubKey, relays []string) string` | Encode PubKey + relay hints to `nprofile1...` format |
| `EncodeNaddr` | `func EncodeNaddr(pk nostr.PubKey, kind nostr.Kind, identifier string, relays []string) string` | Encode addressable entity to `naddr1...` format |
| `EncodePointer` | `func EncodePointer(pointer nostr.Pointer) string` | Encode a `nostr.Pointer` to bech32 string |
| `Decode` | `func Decode(bech32string string) (prefix string, value any, err error)` | Decode any bech32 string — returns prefix + typed value |
| `ToPointer` | `func ToPointer(code string) (nostr.Pointer, error)` | Convert bech32 code to a `nostr.Pointer` struct |
| `NeventFromRelayEvent` | `func NeventFromRelayEvent(ie nostr.RelayEvent) string` | Encode a `nostr.RelayEvent` to `nevent1...` format |

### Key Type Conversion Functions (main `nostr` package)

For hex ↔ typed conversions:

| Function | Signature | Description |
|---|---|---|
| `nostr.IDFromHex` | `func IDFromHex(idh string) (ID, error)` | Convert 64-char hex string → `nostr.ID` |
| `nostr.MustIDFromHex` | `func MustIDFromHex(idh string) ID` | Panic variant |
| `nostr.PubKeyFromHex` | `func PubKeyFromHex(pkh string) (PubKey, error)` | Convert 64-char hex → `nostr.PubKey` |
| `nostr.MustPubKeyFromHex` | `func MustPubKeyFromHex(pkh string) PubKey` | Panic variant |
| `nostr.SecretKeyFromHex` | `func SecretKeyFromHex(skh string) (SecretKey, error)` | Convert 64-char hex → `nostr.SecretKey` |
| `ID.Hex()` | `func (id ID) Hex() string` | `nostr.ID` → 64-char hex string |
| `PubKey.Hex()` | `func (pk PubKey) Hex() string` | `nostr.PubKey` → 64-char hex string |
| `SecretKey.Hex()` | `func (sk SecretKey) Hex() string` | `nostr.SecretKey` → 64-char hex string |

### Codebase Usage Examples

**Decoding (already used throughout codebase)**:

`config/context.go:143` — Decoding private key:
```go
_, s, err := nip19.Decode(privKey)
sk, ok := s.(nostr.SecretKey)
```

`utils/utils.go:11` — Decoding pubkey input:
```go
prefix, decoded, err := nip19.Decode(s)
switch prefix {
case "npub":
    if pk, ok := decoded.(nostr.PubKey); ok {
        return pk, nil
    }
}
```

`utils/alias.go:40` — Resolving npub alias:
```go
prefix, decoded, err := nip19.Decode(resolved)
if prefix == "npub" {
    if pk, ok := decoded.(nostr.PubKey); ok {
        return pk, nil
    }
}
```

**Encoding (already used throughout codebase)**:

`utils/show.go:22` — Encoding event ID to nevent:
```go
nevent := nip19.EncodeNevent(ev.ID, nil, ev.PubKey)
```

`utils/show.go:42` — Encoding pubkey to npub:
```go
npub := nip19.EncodeNpub(ev.PubKey)
```

`utils/profile.go:305,358` — Encoding pubkey:
```go
nip19.EncodeNpub(pubKey)
nip19.EncodeNpub(pk)
```

`config/config.go:140` — Encoding generated secret key:
```go
config.PrivateKey = nip19.EncodeNsec(sk)
```

`cmd/config_commands.go:36` — CLI display:
```go
fmt.Printf("  NPub: %s\n", nip19.EncodeNpub(pubKey))
```

`utils/subscription.go:180` — Subscription ID:
```go
npub := nip19.EncodeNpub(pk)
```

### Critical Note for `EncodeNsec`

`EncodeNsec` takes `[32]byte`, NOT `nostr.SecretKey`:
```go
func EncodeNsec(sk [32]byte) string
```

In `config/config.go:140`, a `nostr.SecretKey` is passed — this works because `nostr.SecretKey` is `[32]byte` underneath.

### For NIP-19 nevent with relay/author hints

`EncodeNevent(id nostr.ID, relays []string, author nostr.PubKey)` — all three args can be provided for full pointer encoding. Pass `nil` for empty relay list.

## Related Specs

- `.trellis/tasks/05-11-fix-getnote-id-parsing/prd.md` — Task PRD with bug description
- `.trellis/spec/backend/index.md` — Backend spec index

## Caveats / Not Found

- `EncodeNsec` takes raw `[32]byte`, not `nostr.SecretKey` struct — but since `nostr.SecretKey` IS `[32]byte`, passing `nostr.SecretKey` works directly
- The SDK version in use is `fiatjaf.com/nostr v0.0.0-20260310013726-4e490879b558` (from go.mod). The current latest docs show `v0.0.0-20260508234157-a4c590d923ee` but function signatures are consistent
