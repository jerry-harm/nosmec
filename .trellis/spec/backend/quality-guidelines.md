# Quality Guidelines

> Code quality standards for backend development.

---

## Overview

<!--
Document your project's quality standards here.

Questions to answer:
- What patterns are forbidden?
- What linting rules do you enforce?
- What are your testing requirements?
- What code review standards apply?
-->

(To be filled by the team)

---

## Forbidden Patterns

<!-- Patterns that should never be used and why -->

(To be filled by the team)

---

## Common Mistakes

### ID/PK Parsing with `copy()` instead of `IDFromHex`/`PubKeyFromHex`

**Symptom**: Event/PubKey queries return nil events even though valid hex IDs are provided.

**Bug Location**: `utils/get.go:136` and similar

**Wrong**:
```go
var id nostr.ID
copy(id[:], noteID)  // BUG: copies ASCII bytes into 32-byte array
```

**Correct**:
```go
id, err := nostr.IDFromHex(noteID)
if err != nil {
    return nil
}
```

**Why it's bad**: `nostr.ID` is `[32]byte` but hex strings are 64 characters. `copy(id[:], noteID)` copies 64 ASCII bytes (not decoded hex) into 32 bytes, resulting in garbage.

**Prevention**: Always use `nostr.IDFromHex()`, `nostr.PubKeyFromHex()`, `nostr.SecretKeyFromHex()` for hex-to-type conversions.

---

## Required Patterns

<!-- Patterns that must always be used -->

### Hex String to nostr Type Conversion

Always use SDK-provided conversion functions:

| From | To | Function |
|------|-----|----------|
| 64-char hex string | `nostr.ID` | `nostr.IDFromHex(s)` |
| 64-char hex string | `nostr.PubKey` | `nostr.PubKeyFromHex(s)` |
| 64-char hex string | `nostr.SecretKey` | `nostr.SecretKeyFromHex(s)` |
| `nostr.ID` | hex string | `id.Hex()` |
| `nostr.PubKey` | hex string | `pk.Hex()` |

---

## NIP-19 Format Convention

All user-facing outputs (CLI, TUI, logs) MUST use NIP-19 bech32 format:
- PubKeys: `npub1...` via `nip19.EncodeNpub(pk)`
- Event IDs: `nevent1...` via `nip19.EncodeNevent(id, relays, author)`
- Private Keys: `nsec1...` via `nip19.EncodeNsec(sk)` (config file only, never in output)

**Internal storage**: Hex format is OK for DB/internal use.
**CLI output**: Always NIP-19.
**Command input**: Accept both hex (64-char) and NIP-19 formats. Use `nip19.ToPointer()` for NIP-19 decoding.

---

## Testing Requirements

<!-- What level of testing is expected -->

(To be filled by the team)

---

## Code Review Checklist

<!-- What reviewers should check -->

(To be filled by the team)
