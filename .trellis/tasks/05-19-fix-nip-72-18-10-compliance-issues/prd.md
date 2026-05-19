# Fix NIP Compliance Issues

## Goal

Fix 3 NIP protocol compliance issues:
1. Community address uses moderator[0] pubkey instead of author pubkey (NIP-72)
2. Quote tags missing relay hint (NIP-10)
3. Deletion events missing p tag (NIP-18)

Note: After verifying against NIP-72 spec, the original concern about kind:1111 vs kind:1 was incorrect — NIP-72 explicitly says kind:1111 with A/a tags is correct. The K/k tags are also standard per NIP-72.

## What I already know

### NIP-72 (Communities) — CORRECTED
- Community definition: kind:34550
- Community address format: `34550:<author-pubkey-hex>:<d-value>` — **author pubkey is from the 34550 event's `pubkey` field, NOT from `p` tags (moderators)**
- Community posts: kind:1111 (NOT kind:1) with `A`/`a` tags — **this is correct per spec**
- The `K`/`k` tags with value "34550" are **standard per NIP-72**
- Bug: address uses `Moderators[0]` pubkey instead of event author pubkey

### NIP-10 (Thread Replies)
- `e` tags: 5-field format `["e", id, relay, marker, pubkey]`
- `q` tags for quotes: should also support relay hint field

### NIP-18 (Deletion)
- Kind:7 deletion events
- Tags: `e` (event to delete) + `p` (pubkey of event author)

## Requirements

1. [ ] Fix community address in `tui/community/discover/model.go` — use `def.Event.PubKey.Hex()` instead of `Moderators[0].Hex()`
2. [ ] Add relay hint to `q` tags in quote/repost (`utils/post.go:QuoteNote`)
3. [ ] Add `p` tag with author pubkey to deletion events (`utils/post.go:DeleteNote`)

## Acceptance Criteria

- [ ] Community address format: `34550:<event-author-pubkey>:<d-value>`
- [ ] Quote/repost events have relay hint in `q` tag
- [ ] Deletion events have `p` tag with original event author's pubkey
- [ ] `go build ./...` succeeds
- [ ] Existing tests pass

## Definition of Done

- All 3 changes implemented
- Build succeeds
- Spec updated if needed

## Out of Scope

- Changing kind:1111 to kind:1 — NIP-72 says use kind:1111 (current code is correct)
- Removing K/k tags — these are standard per NIP-72

## Technical Notes

- `def.Event.PubKey.Hex()` gives the 34550 event author's pubkey (correct for address)
- `config.GetEventRelay(eventID.Hex())` for relay hint lookups
- Files: `tui/community/discover/model.go`, `utils/post.go`