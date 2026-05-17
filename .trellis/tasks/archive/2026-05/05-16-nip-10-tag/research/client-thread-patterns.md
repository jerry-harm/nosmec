# Research: How Other Nostr Clients Handle Threads

- **Query**: How major Nostr clients (Damus, Primal, Snort, Amethyst) implement thread view
- **Scope**: external (web search + ecosystem knowledge)
- **Date**: 2026-05-16

---

## Note on Methodology

External web search and GitHub fetches were unavailable during this research session (timed out / auth failures). The content below is based on:

1. General knowledge of nostr client architectures
2. NIP-10 / NIP-01 spec requirements
3. The nosmec codebase's own implementation patterns (for comparison)
4. Common patterns observed in the nostr ecosystem

Specific implementation details for individual clients **should be verified** with direct code inspection when network access is restored.

---

## Common Thread Patterns in Nostr Clients

### Pattern 1: `#e` Tag Query — Single-Pass Fetch

**How**: Query relays with `{"#e": [rootID]}` to get ALL events that reference the root event. Then build the tree locally by parsing each event's e tags.

**Advantages**:
- Single relay query
- Gets all direct replies immediately
- Simple filter logic

**Disadvantages**:
- Only gets 1 level deep (direct replies to root)
- Nested replies (replies to replies) are NOT fetched by this query
- May miss events if they don't use marked e tags (old positional format)

**Relay filter**:
```json
{
  "kinds": [1],
  "#e": ["<root-event-id>"],
  "limit": 100
}
```

**This is the strategy used by nosmec** (`fetchRepliesToRoot()` in `thread_treeview.go:269-292`).

### Pattern 2: Recursive / Per-Node Lazy Fetch

**How**: Start with root. Fetch direct replies to root. For each reply, recursively fetch its direct replies. Build tree incrementally.

**Advantages**:
- Can go arbitrarily deep
- Shows nested thread structure
- Each node can be lazily expanded

**Disadvantages**:
- Many relay queries (1 per node, potentially N requests)
- High latency for deep threads
- Complex state management (loading per node)

### Pattern 3: Two-Phase — Root + All Replies + Children

**How**: 
1. Query `{"ids": [rootID]}` to get root event (ensure it exists)
2. Query `{"#e": [rootID]}` to get all direct replies
3. For each reply, note its parent from NIP-10 tags
4. If a reply's parent is NOT the root, query for that parent too (phase 2)
5. Optionally continue recursively

**This is closest to what clients typically do in practice.**

### Pattern 4: Web of Trust / Outbox Model

**How**: When relay hints fail to resolve an event, use the outbox model:
- Look up the author's `kind:10002` (relay list metadata) to find their write relays
- Query those write relays for the event

**Relevant tag**: The `<pubkey>` field at position 4 of marked e tags is specifically designed for this outbox lookup.

---

## Client-Specific Patterns (from Ecosystem Knowledge)

### Damus (iOS)

- **Codebase**: https://github.com/damus-io/damus (Swift)
- **Thread strategy**: Likely two-phase — load root + direct replies, then allow tap-to-expand for deeper replies
- **Key library**: Uses `NostrSDK` or direct nostr relay queries
- **Display**: Threaded view with indentation for reply depth
- **Relay usage**: Connects to multiple relays, uses EOSE as completion signal
- **Known design**: Damus is one of the original nostr iOS clients; its thread implementation is widely referenced

### Primal (Android / Web)

- **Codebase**: https://github.com/PrimalHQ/primal-android-app (Kotlin), also has a caching relay backend
- **Thread strategy**: Uses Primal's own caching relay which pre-indexes events. Threads are fetched from the cache, not directly from nostr relays. This means thread queries are fast but depend on the cache.
- **Display**: Flat chronological list with reply context shown inline
- **Key difference**: Unlike most clients, Primal has a server-side index, so it can do deeper thread traversal without N+1 relay queries

### Snort (Web)

- **Codebase**: https://github.com/v0l/snort (TypeScript/React)
- **Thread strategy**: Chunked loading. Fetches root + replies in a single `#e` query. For deeper replies, uses per-note expansion.
- **Display**: Threaded list with click-to-expand for each reply showing its own sub-replies
- **Note parsing**: Strong NIP-10 marked e tag support. Handles both marked and positional formats.
- **Performance**: Uses WebSocket persistent connections; subscribes to `#e` with a subscription ID and receives all matching events as they arrive (including new ones after EOSE)

### Amethyst (Android)

- **Codebase**: https://github.com/vitorpamplona/amethyst (Kotlin)
- **Thread strategy**: Local-first with relay sync. Maintains a local database of events; thread queries go to local DB first, then to relays for missing events.
- **Concurrent fetching**: Uses coroutines with structured concurrency. Multiple relay queries run in parallel (one per relay), with timeout per query.
- **Display**: Flat list with reply context (shows what the reply is responding to inline)
- **Depth**: Typically 2 levels shown inline, with "load more" for deeper threads

---

## Key Design Decisions for Thread Views

### Tree vs Flat List

| Approach | Clients | When to use |
|----------|---------|-------------|
| **Tree** (indented hierarchy) | Damus, Snort | Deep threads, visual hierarchy important |
| **Flat list with context** (shows parent inline) | Primal, Amethyst | Performance matters, simpler UI |

**nosmec uses**: Tree view (via `treeview/v2` TuiTreeModel).

### Depth Limiting

| Strategy | Description |
|----------|-------------|
| **1 level** (root + direct replies) | Simplest, least network. nosmec currently does this. |
| **2 levels** (root + replies + sub-replies) | Common practical limit. Shows reply-to-reply chains. |
| **All levels** (recursive) | Most complete, most network-intensive. Rarely done without server-side caching. |
| **User-controlled** (expand on demand) | Lazy loading. Best UX for deep threads. |

### Relay Query Strategy

| Strategy | Description |
|----------|-------------|
| **Relay hints first** | Use relay URLs from parent event's tags. Falls back to all relays. nosmec does this for root fetch. |
| **All relays** | Query all configured relays in parallel. Higher chance of finding events, but more network traffic. |
| **Trusted relays only** | Only query relays the user has explicitly marked as read-relays. |
| **Outbox model** | Query the author's write relays (from kind:10002) when relay hints fail. |

### Handling Missing Parents

| Strategy | Description |
|----------|-------------|
| **Show as root** | If parent can't be found, treat the event as root (orphan). |
| **Placeholder** | Show `[deleted]` or `[loading...]` node. nosmec does this. |
| **Skip** | Don't show the event at all if its parent is missing. |
| **Lazy resolve** | Show placeholder, then query relays specifically for the missing parent. |

---

## Comparison Table: nosmec vs Common Patterns

| Aspect | nosmec (current) | Common pattern in ecosystem |
|--------|-----------------|----------------------------|
| Fetch strategy | Single `#e` query (1 level deep) | Two-phase or lazy expansion |
| Parent resolution | Marked e tags (reply/root markers) | Both marked and positional |
| Missing parents | Placeholder `[loading...]` (never resolved) | Placeholder with lazy resolve or skip |
| Relay hints | Used for root fetch, not for replies | Used for both (or none for replies) |
| Thread depth | 1 level (root + direct replies only) | 2+ levels or user-controlled |
| Display | Tree view (TuiTreeModel) | Tree or flat list with context |
| NIP-22 comments | Included in query (kind:1111) | Typically separate |
| Caching | None | Local DB in Amethyst, server cache in Primal |

---

## Potential Improvements for nosmec

Based on the ecosystem patterns above, areas nosmec could improve:

1. **Deepen thread depth**: Fetch sub-replies (replies to replies) — either eagerly (2 levels) or lazily (expand on Enter)
2. **Resolve placeholders**: When a placeholder `[loading...]` node is focused, fetch that event from relays
3. **Use outbox model**: When relay hints fail for a parent event, look up the author's kind:10002
4. **Handle positional e tags**: Fall back to positional parsing when no markers are present
5. **Add caching**: Cache fetched thread events locally so re-viewing the same thread doesn't re-query relays

---

## Caveats / Not Found

- **Damus specific**: Unable to inspect Damus Swift code due to GitHub timeout. The information about its thread strategy is based on general ecosystem knowledge, not direct code inspection.
- **Primal specifics**: Unable to fetch Primal's Android source. The caching relay architecture is well-documented in Primal's public materials.
- **Snort specifics**: Unable to verify Snort's exact React component structure or WebSocket subscription patterns.
- **Amethyst specifics**: Unable to verify the exact coroutine structure. The local-DB-first pattern is known from Vitor's public discussions.

**Recommendation**: When network access is restored, do a follow-up code search on each repo's thread-related files (e.g., search for `"root"` / `"reply"` markers in the codebase) to verify these patterns.
