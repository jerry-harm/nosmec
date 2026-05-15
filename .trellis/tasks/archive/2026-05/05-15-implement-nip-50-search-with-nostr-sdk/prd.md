# brainstorm: NIP-50 search implementation

## Goal

Implement full NIP-50 search with complete extension support and local relay integration.

## What I already know

* `nostr.Filter` already has `Search` field (NIP-50 supported)
* Local relay already has NIP-50 via Bleve (`config/config.go:302-344`)
* `SearchEvents` exists in `utils/search.go` - parses `kinds:`, `authors:`, `#t:`, `search:`
* Bleve backend needs `RawEventStore` wrapping another store

## Open Questions

* ParseSearchFilter needs NIP-50 extensions: `domain:` `language:` `sentiment:` `nsfw:` `include:spam`

## Requirements (evolving)

* Full NIP-50 extension parsing in ParseSearchFilter
* Local relay NIP-50 already enabled via Bleve

## Technical Notes

* Local relay already implements NIP-50 via Bleve + dual-store pattern