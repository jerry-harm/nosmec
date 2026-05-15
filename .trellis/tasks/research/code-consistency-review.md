# Research: Code Consistency Review

- **Query**: Review nosmec codebase for consistency between spec and implementation
- **Scope**: internal (code vs spec files)
- **Date**: 2025-05-15

## Findings

### 1. relay-guidelines.md NIP-65 Implementation vs relay_list.go

#### Tag Format Parsing — CONSISTENT

**relay-guidelines.md** states:
- `["r", <url>]` — both read AND write (no marker)
- `["r", <url>, "read"]` — read only
- `["r", <url>, "write"]` — write only

**relay_list.go:syncRelayListFromNetwork** (lines 52-71):
```go
relayMap := make(map[string]struct{ read, write bool })
for _, tag := range result.Event.Tags {
    if len(tag) >= 2 && tag[0] == "r" {
        url := tag[1]
        r := relayMap[url]
        if len(tag) == 2 {
            r.read = true
            r.write = true
        } else {
            for _, p := range tag[2:] {
                if p == "read" {
                    r.read = true
                } else if p == "write" {
                    r.write = true
                }
            }
        }
        relayMap[url] = r
    }
}
```
✅ Correctly implements NIP-65 tag parsing: len==2 means both read+write, otherwise scan for markers.

**profile.go:FetchRecipientReadRelays** (lines 289-307) uses same pattern correctly.

#### Publish Relay List — INCONSISTENT

**relay-guidelines.md** (lines 119-120):
> **Never** create separate tags for the same relay with read AND write markers

**relay_list.go:publishRelayListMetadata** (lines 147-187):
```go
for _, relay := range relayList {
    read := relay.Read != nil && *relay.Read
    write := relay.Write != nil && *relay.Write
    if read && write {
        continue  // ← SKIPS both-read-write relays entirely
    }
    if read {
        tags = append(tags, nostr.Tag{"r", relay.URL, "read"})
    } else if write {
        tags = append(tags, nostr.Tag{"r", relay.URL, "write"})
    }
}
```
⚠️ **BUG**: When a relay has both `read=true` and `write=true`, the code skips it entirely — no tag is published. Per NIP-65 spec, it should publish `["r", url]` (no marker = both).

### 2. quality-guidelines.md NIP URL Pattern

**quality-guidelines.md** (line 35):
```
https://github.com/nostr-protocol/nips/raw/refs/heads/master/{nip}.md
```

This URL pattern is **correct** — GitHub supports `raw/refs/heads/...` for canonical raw content.

### 3. NIP-19 Format Convention — CONSISTENT

**quality-guidelines.md** (lines 107-116):
- CLI output MUST use NIP-19 bech32 format
- Input: Accept both hex and NIP-19 via `nip19.ToPointer()`

**Implementation** in config/context.go (lines 133-150):
```go
_, s, err := nip19.Decode(privKey)
sk, ok := s.(nostr.SecretKey)
```
✅ Uses `nip19.Decode` correctly for input handling.

**profile.go** encodes with:
```go
nip19.EncodeNpub(pubKey)  // line 336
nip19.EncodeNpub(pk)      // line 400
```
✅ Consistent NIP-19 encoding for output.

### 4. sdk.ProfileMetadata Field Name Inconsistency

**nostr-sdk-usage.md** (lines 255-268):
```go
type ProfileMetadata struct {
    ...
    LUD16     string  // ← uppercase LUD16
}
```

**Implementation** in utils/profile.go (lines 93-103):
```go
func ProfileMetadataFromSDK(pm sdk.ProfileMetadata) ProfileMetadata {
    return ProfileMetadata{
        ...
        LUD16: pm.LUD16,  // ← uses LUD16 (uppercase)
    }
}
```

But **sdk.ProfileMetadata** per nostr-sdk-usage.md has `LUD16` (uppercase) — however, the JSON tag in local ProfileMetadata is `lud16` (lowercase, line 26):
```go
type ProfileMetadata struct {
    ...
    Lud16 string `json:"lud16,omitempty"`
}
```
✅ Field name mismatch between Go struct (Lud16) and JSON tag (lud16) — this is intentional per Go conventions (unexported JSON-lowercased fields map correctly).

### 5. ExtractRelayHints — SPEC VS CODE MISMATCH

**relay-guidelines.md** (lines 24-38) specifies:
```go
switch tag[0] {
case "e", "p", "a", "q":
    if len(tag) >= 3 && tag[2] != "" {  // relay hint at tag[2]
        relays = append(relays, tag[2])
    }
}
```

**utils/get.go:ExtractRelayHints** (lines 26-36):
```go
if len(tag) < 3 {  // ← requires len >= 3
    continue
}
switch tag[0] {
case "e", "p", "a", "q":
    if relay := tag[2]; relay != "" && !seen[relay] {  // ← tag[2] = relay hint
        relays = append(relays, relay)
        seen[relay] = true
    }
}
```
✅ The **spec is correct** — `len >= 3` is checked correctly.

### 6. NIP-65 Parse Tags in user_relays.go vs relay-guidelines.md

**relay-guidelines.md** (lines 92-95):
> Parses `["r", <url>]` or `["r", <url>, "read"|"write"]` tags

**utils/user_relays.go** (lines 65-68):
```go
readRelays, writeRelays := nip65.ParseRelayList(*event)

reachableRead, _ := VerifyRelaysConnectivity(ctx, app, readRelays)
reachableWrite, _ := VerifyRelaysConnectivity(ctx, app, writeRelays)
```
✅ Uses `nip65.ParseRelayList` from the nostr library — delegate pattern is appropriate.

### 7. FollowInfo NIP-19 Storage — CONSISTENT

**nostr-sdk-usage.md** (lines 270-275):
```go
type ProfileRef struct {
    Pubkey  nostr.PubKey
    Relay   string
    Petname string
}
```

**Implementation** in utils/profile.go (lines 35-39):
```go
type FollowInfo struct {
    NPub    string `json:"npub"`     // NIP-19 encoded
    Relay   string `json:"relay,omitempty"`
    Petname string `json:"petname,omitempty"`
}
```
✅ Stores `npub` as NIP-19 string, matching quality-guidelines output requirement.

### 8. Profile Fetch Strategy — MATCHES SPEC

**relay-guidelines.md** (lines 183-212) describes parallel profile fetch:
```
GetProfile
  ├─ Query ALL relays in parallel for kind 0 (profile metadata)
  │    └─ Returns as soon as ANY relay responds
  └─ Goroutine: DiscoverUserRelays (async, parallel)
```

**utils/get.go:GetProfile** (lines 150-198):
```go
// Launch DiscoverUserRelays async to update KnownRelays for future use
go func() {
    DiscoverUserRelays(context.Background(), opts.App, pubKey)
}()

// Query all relays in parallel, return first result
result := opts.App.Pool().QuerySingle(ctxQuery, combined, filter, ...)
```
✅ Correctly implements the parallel discovery profile fetch strategy.

### 9. Relay Selection Fallback Order — CONSISTENT

**relay-guidelines.md** (lines 158-175):
```go
relays := opts.Relays
if len(relays) == 0 {
    relays = opts.App.AllReadableRelays()
}
if len(relays) == 0 {
    relays = opts.App.Config().KnownRelays
}
```

**utils/get.go** uses this pattern in `GetEvent` (line 45-48), `GetGlobalTimeline` (lines 481-487), and `GetFollowedTimeline` (lines 524-530).

### 10. publishRelayListMetadata Bug — DETAILED

**relay_list.go:publishRelayListMetadata** (lines 147-187):
```go
for _, relay := range relayList {
    read := relay.Read != nil && *relay.Read
    write := relay.Write != nil && *relay.Write
    if read && write {
        continue  // BUG: skips publishing tag entirely
    }
    if read {
        tags = append(tags, nostr.Tag{"r", relay.URL, "read"})
    } else if write {
        tags = append(tags, nostr.Tag{"r", relay.URL, "write"})
    }
}
```

**Expected per NIP-65**: `["r", url]` with no marker for both-read-write relays.
**Actual**: Relay is omitted from the event entirely.

This means when a user has a relay configured as both read AND write, their published NIP-65 event will NOT include that relay — violating the NIP-65 spec which requires one tag per relay, with no marker meaning both.

### 11. quality-guidelines.md placeholder content

**quality-guidelines.md** contains extensive TUI-specific content (lines 134-420) that appears copied from a BubbleTea v2 guide. The "Forbidden Patterns" (line 27) and "Testing Requirements" (line 124) sections are still marked as `(To be filled by the team)`.

### 12. search.go NIP-50 ParseSearchFilter

**nostr-sdk-usage.md** references NIP-50 but doesn't detail the search filter format.

**utils/search.go:ParseSearchFilter** (lines 22-72) parses `kinds:`, `authors:npub1...`, and `#t:hashtag` syntax.

No spec document was found that defines the expected search filter syntax for this project.

## Caveats / Not Found

- No spec for NIP-50 search filter syntax was found — `ParseSearchFilter` implementation was not compared against any spec
- `DiscoverAndVerifyRelays` function in relay-guidelines.md (lines 87-95) is not found in the codebase — may be unimplemented
- NIP-17 tag format for DM relays (`["relay", url]`) is correctly implemented in `syncDMRelaysFromNetworkImpl` and `publishDMRelayList`