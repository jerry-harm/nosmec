# Relay Guidelines

> NIP compliance and relay configuration patterns.

---

## NIP Relay List Events

### NIP-65 — Relay List Metadata (kind:10002)

Published to advertise user's preferred read/write relays.

```go
// kind:10002 structure
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay1.com"],              // both read+write
    ["r", "wss://relay2.com", "write"],    // write only
    ["r", "wss://relay3.com", "read"]      // read only
  ],
  "content": ""
}
```

Published via `utils.PublishRelayList(ctx, app)`.

### NIP-17 — DM Relay List (kind:10050)

Published to advertise user's DM inbox relays.

```go
// kind:10050 structure
{
  "kind": 10050,
  "tags": [
    ["relay", "wss://inbox.nostr.wine"],
    ["relay", "wss://myrelay.nostr1.com"]
  ],
  "content": ""
}
```

Published via `utils.PublishRelayList(ctx, app)` (same function handles both).

---

## Convention: Auto-publish on Config Mutation

**When relay configuration is mutated via CLI, always publish the updated relay list.**

```bash
# After these commands, PublishRelayList MUST be called:
nosmec config relay add <url>        # then PublishRelayList
nosmec config relay remove <url>     # then PublishRelayList
nosmec config relay sync             # then PublishRelayList
nosmec config dm-relay add <url>     # then PublishRelayList
nosmec config dm-relay remove <url>  # then PublishRelayList
nosmec config dm-relay sync         # then PublishRelayList
```

**Why**: `PublishRelayList` existed in `utils/relay_list.go` but was not wired to CLI mutations — relay lists were configured locally but never broadcast to the network.

---

## Relay Configuration Semantics

| Field | Purpose |
|-------|---------|
| `RelayList` | User's relays with read/write flags (NIP-65) |
| `PrivateRelays` | Private data relay (DMs, follows — sensitive data) |
| `DMRelays` | DM inbox relays (NIP-17 kind:10050) |
| `SearchRelays` | Search-only relay (BLEVE index queries) |
| `KnownRelays` | Fallback relay discovery list |

---

## Files

- `utils/relay_list.go` — `PublishRelayList` implementation
- `config/types.go` — relay config structs
- `config/context.go` — relay helper methods
