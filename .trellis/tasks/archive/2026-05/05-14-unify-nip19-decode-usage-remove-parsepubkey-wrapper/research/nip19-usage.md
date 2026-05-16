# Research: nip19.Decode usage and ParsePubKey wrapper

- **Query**: Find all ParsePubKey call sites, check nip19.Decode return types per prefix, verify if 66-char compressed key hack is needed, check for standard library support
- **Scope**: internal + external (nostr library)
- **Date**: 2026-05-14

## Findings

### Files Found

| File Path | Description |
|---|---|
| `utils/utils.go` | Contains `ParsePubKey` wrapper function (lines 10-30) |
| `utils/alias.go` | Contains `ResolveAliasToPubKey` with duplicate nip19 logic (lines 23-47) |
| `utils/search.go` | Calls `ParsePubKey` at line 44 for author filter parsing |
| `tui/dm/main.go` | Calls `utils.ParsePubKey` at line 13 |
| `cmd/dm_commands.go` | Calls `utils.ParsePubKey` at lines 40, 85, 103 |
| `config/context.go` | Direct `nip19.Decode` at line 140 for nsec privkey decode |
| `config/config.go` | Direct `nip19.Decode` at line 157 for nsec privkey decode |

### ParsePubKey Call Sites (3 files)

| File | Line | Usage |
|---|---|---|
| `utils/search.go` | 44 | `ParsePubKey(authorStr)` — parses author filter in search query |
| `tui/dm/main.go` | 13 | `utils.ParsePubKey(npubOrHex)` — validates recipient pubkey |
| `cmd/dm_commands.go` | 40, 85, 103 | Multiple DM recipient/otherPK parsing |

### nip19.Decode Direct Usage (3 files, 4 call sites)

| File | Line | Purpose |
|---|---|---|
| `utils/utils.go` | 15 | Inside `ParsePubKey` — handles npub bech32 decode |
| `utils/alias.go` | 29 | Inside `ResolveAliasToPubKey` — identical logic to ParsePubKey |
| `config/context.go` | 140 | Decodes nsec to get `nostr.PrivateKey` for signing |
| `config/config.go` | 157 | Decodes nsec from config to get `nostr.PrivateKey` |

### Code Patterns

**Current ParsePubKey logic** (`utils/utils.go:10-30`):
```go
func ParsePubKey(s string) (nostr.PubKey, error) {
    if pk, err := nostr.PubKeyFromHex(s); err == nil {  // 64-char hex
        return pk, nil
    }
    prefix, decoded, err := nip19.Decode(s)
    if err == nil {
        if prefix == "npub" {
            if pk, ok := decoded.(nostr.PubKey); ok {
                return pk, nil
            }
            return nostr.PubKey{}, fmt.Errorf("invalid npub format")
        }
    }
    if len(s) == 66 && (s[:2] == "02" || s[:2] == "03") {  // 66-char compressed
        return nostr.PubKeyFromHex(s[2:])
    }
    return nostr.PubKey{}, fmt.Errorf("unknown pubkey format")
}
```

**Duplicate 66-char compressed key hack** — exists in BOTH:
- `utils/utils.go:25-27`
- `utils/alias.go:38-40`

**Duplicate npub+hex handling** — `ResolveAliasToPubKey` (`alias.go:29-46`) reimplements the same 3-step logic as `ParsePubKey`:
1. Try `nostr.PubKeyFromHex` (64-char hex)
2. Try `nip19.Decode` with `npub` prefix check
3. Try 66-char compressed key hack

### nip19.Decode Return Types by Prefix

Based on `fiatjaf.com/nostr` library usage in codebase:

| Prefix | Return Type | Notes |
|---|---|---|
| `npub` | `nostr.PubKey` | Used for pubkey decoding |
| `nsec` | `nostr.PrivateKey` | Used in config/context for private key |
| `note` | `string` | Raw event ID |
| `nevent` | `struct{ID string; Relays []string; PubKey nostr.PubKey}` | Event with relay hints |
| `naddr` | `struct{Kind int; Identifier string; Author nostr.PubKey; Relays []string}` | Addressable event |
| `nprofile` | `struct{PubKey nostr.PubKey; Relays []string}` | Profile with relay hints |

### 66-Char Compressed Key Hack Analysis

**Where it's used**: Both `utils.go:25-27` and `alias.go:38-40` check:
```go
if len(s) == 66 && (s[:2] == "02" || s[:2] == "03") {
    return nostr.PubKeyFromHex(s[2:])
}
```

**Why it exists**: NIP-19 `encodePack` can produce a "compressed pubkey" format (66 chars: 02/03 prefix + 64-char hex). This is NOT the same as raw hex — it has a parity/format prefix byte.

**Whether any caller passes this format**: Based on grep results, no caller explicitly constructs or passes 66-char format. All call sites pass either:
- `npub1...` bech32 strings
- Raw 64-char hex strings from `nostr.PubKey` objects
- User input that could theoretically be in compressed format

**Verdict**: The 66-char hack is a defensive fallback for a format that may not actually appear in practice. It should be verified whether this is dead code, but given the recent `05-14-npub-parsing-unification` task (archived), it was intentional to support this format.

### Standard Library Function for Compressed Pubkeys

The `fiatjaf.com/nostr` library does NOT have a native function that handles all three formats (npub bech32, 64-char hex, 66-char compressed) in one call. The `nip19.Decode` function returns the raw decoded data with type assertion required per prefix.

### Related Specs

- `.trellis/spec/backend/index.md` — backend package specs

### Caveats / Not Found

- The `fiatjaf.com/nostr` library API details for `nip19.Decode` return types are inferred from usage patterns in the codebase, not from library documentation. A web search for the exact API would confirm return types for `nevent`, `naddr`, `nprofile` prefixes.
- No research directory existed prior to this task — created at task path.