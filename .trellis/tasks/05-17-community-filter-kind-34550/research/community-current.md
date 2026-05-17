# Research: Current Community Implementation

- **Query**: Understand what we're replacing — how community discovery currently works, how kind 34550 events are structured, and how the DM list model works as a reuse pattern.
- **Scope**: internal + external (NIP-72 spec)
- **Date**: 2026-05-17

## Findings

### 1. How Community Discovery Currently Works

There is **no community discovery feature** today. The current flow is entirely "you must already know the community address":

| Step | How it works | Key Code |
|---|---|---|
| **Create** | `cmd/community_commands.go` — `community create <name> <id> [desc]` publishes a kind 34550 event | `utils/community.go:CreateCommunity()` (line 23) |
| **Follow** | `utils/subscription.go:FollowCommunity()` (line 15) — adds a `Subscription{Type:"community", ID: addr}` to local config | Also syncs from network via kind 10004 event (`syncCommunitiesFromNetwork`, line 93) |
| **List my communities** | `cmd/community_commands.go` — `community list` shows: (1) Followed (kind 10004) (2) Created (kind 34550 by my pubkey) (3) Posted (kind 1111 by my pubkey) | Lines 111-182 |
| **View timeline** | `cmd/community_commands.go` — `community timeline <addr>` — runs `timeline.RunTimeline(app, "community", nil, limit, addr)` | Line 242 |
| **Post** | `cmd/community_commands.go` — `community post <addr> <content>` — publishes kind 1111 with A/a/P/p/K/k tags | `utils/community.go:PostToCommunity()` (line 100) |

**Key gap**: There is no way to browse/discover communities you don't already know about. To follow a community, you must already have its full address (`34550:<author_pubkey>:<d-identifier>`).

### 2. How Community Data Is Currently Displayed

#### CLI Display (`community list`)
```
=== My Communities ===

[Following] (Kind 10004)
  - 34550:pubkey:community-name

[Created] (Kind 34550)
  - <name from name tag or d tag>

[Posted] (Kind 1111)
  - 34550:pubkey:community-name
```

#### TUI Timeline (`community timeline`)
- Uses `tui/timeline/model.go` with `filter: "community"` and `communityAddr` set
- Fetches posts via `utils.GetCommunityPosts()` (kind 1111 filtered by `a` tag)
- Displays each post as an item with author name, prefix label `[Community]`, and content preview
- Has infinite scroll, real-time subscription, profile name resolution

**The `communityAddr` format used throughout**: `34550:<author-pubkey-hex>:<d-identifier>` (e.g., `34550:6cf684fc1d77b652a90cb62767c5342f8efd1eb4bf27edc5c15ae1df3b71204d:my-community`)

Parsing is in `utils/community.go:ParseCommunityAddr()` (line 82):
```go
func ParseCommunityAddr(addr string) (nostr.PubKey, string, error) {
    parts := strings.Split(addr, ":")
    // parts[0] must be "34550"
    // parts[1] is hex pubkey
    // parts[2] is community identifier (d-tag value)
}
```

### 3. Kind 34550 Event Structure (CommunityDefinition)

From `fiatjaf.com/nostr` library (`kinds.go` line 411):
```go
KindCommunityDefinition Kind = 34550
```

**NIP-72 spec structure** (from `docs/NIP.md` lines 433-444 and the actual NIP-72 spec):

```json
{
  "kind": 34550,
  "created_at": <unix timestamp>,
  "tags": [
    ["d", "<community-d-identifier>"],
    ["name", "<Community name>"],
    ["description", "<Community description>"],
    ["image", "<Community image URL>", "<Width>x<Height>"],
    ["p", "<moderator-pubkey>", "<optional relay>", "moderator"],
    ["relay", "<relay URL>", "<purpose-marker>"]
  ],
  "content": ""
}
```

**Tag details from actual usage** in `utils/community.go:CreateCommunity()` (lines 33-53):
```go
tags := nostr.Tags{
    {"d", def.ID},              // d-tag: community identifier
    {"name", def.Name},         // name tag: display name
    {"description", def.Description},  // description tag
}
// Optional: {"image", def.ImageURL, "256x256"}
// Optional: {"p", moderator.Hex(), "", "moderator"} — for each moderator
// Optional: {"relay", relayURL, purpose}
```

**Kind classification**: 34550 is **addressable** (`IsAddressable()` returns true, since 30000 <= kind < 40000). Replaceable events use `d` tag as the replaceable key. Queries should use the relay's addressable event support.

### 4. How the DM List Model Works (Reuse Pattern)

File: `tui/dm/list/model.go` (232 lines)

This is the pattern we want to mimic for community discovery:

```
┌─────────────────────────────────┐
│ DM Conversations (title)         │
├─────────────────────────────────┤
│  @alice                         │ ← list.Item (Title: name, Desc: latest msg)
│  → hello                        │
├─────────────────────────────────┤
│  @bob                           │
│  ← how are you?                 │
├─────────────────────────────────┤
│  [enter to open conversation]   │ ← on enter, returns openDM msg
└─────────────────────────────────┘
```

**Architecture**:

| Component | What it does |
|---|---|
| `conversationItem` struct (line 18) | Implements `list.Item` interface — `Title()`, `Description()`, `FilterValue()` |
| `model` struct (line 48) | Holds `styles`, `list.Model`, `keys`, `app`, `items`, `errMsg`, `loaded` flag |
| `NewModel(app)` (line 108) | Constructor — creates list with default delegate, sets title |
| `Init()` (line 121) | Returns `m.loadConversations()` — async data loading |
| `loadConversations()` (line 125) | Async Cmd that calls `utils.ListDMConversations()`, returns `loadedMsg{items}` |
| `loadedMsg` handler (line 168) | Converts items to `list.Item` slice, sets on list, updates status |
| `errMsg` handler (line 181) | Shows error |
| `KeyMsg` handler (line 185) | `esc` → quit, `ctrl+c` → kill, `enter` → select (returns `openDM` msg) |
| `View()` (line 213) | Renders list with AltScreen |
| `RunDMList(app)` (line 219) | Entry point — creates `tea.NewProgram(m).Run()` — **no bubblon wrapper** |

**Key differences from timeline model**:
- DM list is a **standalone tea program** (no bubblon controller)
- DM list uses `list.NewDefaultDelegate()` not a custom delegate
- DM list doesn't use infinite scroll or real-time subscriptions
- DM list has no dark/light mode switching
- DM list is **much simpler** — ~230 lines vs. timeline's ~837 lines

**The pattern to reuse for community discovery**:
1. Create an `item` struct implementing `list.Item`
2. Create a `model` struct with `list.Model`, `app`, `keys`
3. `Init()` returns a data-loading Cmd
4. Handle `loadedMsg` to populate the list
5. Handle `enter` to select/join a community
6. Handle `esc` to go back

### 5. NIP-72 Community Discovery

**From the NIP-72 spec** (github.com/nostr-protocol/nips/72.md):

- **Community Definition**: `kind:34550` — a replaceable event defining the community, its moderators, and preferred relays
- **Community Posts**: `kind:1111` (NIP-22) with `A`/`a` tags pointing to the community
- **Moderation**: `kind:4550` — approval events by moderators
- **Community Lists**: `kind:10004` — per-user list of followed communities (NIP-51 lists)

**How discovery works in NIP-72**:
- There is **NO standard "global community list"** in NIP-72
- Community discovery is done by querying relays for **all kind 34550 events** (without author filter) — i.e., `{"kinds": [34550], "limit": N}`
- Each kind 34550 event has a `d` tag (addressable identifier), plus `name`, `description`, `image` tags
- The `name` tag is preferred over the `d` tag for display
- Relays store kind 34550 as addressable events (30000-40000 range)
- The event `content` field is typically empty (all metadata is in tags)

**Global community query approach**:
```go
filter := nostr.Filter{
    Kinds: []nostr.Kind{nostr.KindCommunityDefinition},  // 34550
    Limit: 50,
}
// No Authors filter = global search
events := pool.FetchMany(ctx, relays, filter, opts)
```

Each returned event contains:
- `Tags.Find("d")` → community identifier
- `Tags.Find("name")` → display name
- `Tags.Find("description")` → description
- `Tags.Find("image")` → image URL
- `event.PubKey` → community author/owner
- `event.CreatedAt` → creation time
- `Tags.FindAll("p")` with marker `"moderator"` → moderators

### 6. Related Code References

| File | Relevance |
|---|---|
| `cmd/community_commands.go:111-182` | Current `community list` CLI — shows 3 categories of user-affiliated communities |
| `utils/community.go:14-21` | `CommunityDefinition` struct — the Go model for community metadata |
| `utils/community.go:23-80` | `CreateCommunity()` — publishes kind 34550 |
| `utils/community.go:82-98` | `ParseCommunityAddr()` — parses `34550:pubkey:id` format |
| `utils/community.go:186-205` | `GetCommunity()` — fetches single community by author+d-tag |
| `utils/community.go:253-279` | `GetFollowedCommunities()` — reads kind 10004 |
| `utils/community.go:330-352` | `GetMyCreatedCommunities()` — kind 34550 by my pubkey |
| `utils/subscription.go:15-22` | `FollowCommunity()` — adds to local subscriptions |
| `utils/subscription.go:93-136` | `syncCommunitiesFromNetwork()` — reads kind 10004 for followed communities |
| `utils/subscription.go:303-338` | `publishCommunitiesList()` — publishes kind 10004 event |
| `config/types.go:128-133` | `Subscription` struct — `{Type, ID, Relay, Petname}` |
| `tui/timeline/model.go:196-230` | Timeline model — `NewModel` with filter param, shows how `community` filter works |
| `tui/timeline/model.go:262-283` | `fetchTimeline()` — community case uses `GetCommunityPosts()` |
| `tui/timeline/model.go:406-417` | `startSubscription()` — community case subscribes to `a` tag |
| `tui/dm/list/model.go` | **Target pattern** — simple list model to mimic for community discovery |
| `docs/NIP.md:21, 431-457` | NIP-72 documentation in project |
| `docs/NIP.md:5-21` | All supported NIPs — 34550, 1111, 4550 |

### 7. Caveats / Not Found

- **No existing "community discovery" code** in the project — this is a net-new feature
- **NIP-72 does not define a standard discovery mechanism** beyond "query relays for kind 34550 events" — there's no curated directory or recommendation system
- **Kind 34550 is addressable** (30000-40000 range), not replaceable (10000-20000) — `FetchManyReplaceable()` would NOT work; must use `FetchMany()` or a relay-specific addressable query
- **The `d` tag is the community identifier** — it may or may not be human-readable; the `name` tag should be preferred for display
- **Community addresses in-code use format `34550:pubkey:id`** (with actual hex pubkey), while NIP-72 examples show `34550:<community-author-pubkey>:<community-d-identifier>` — they match
- **The DM list model does NOT use bubblon** (unlike the timeline) — it runs as a standalone `tea.NewProgram` — this is simpler and appropriate for a discovery list that doesn't need window management
- **No relay hint extraction** is done for kind 34550 events currently beyond what `ExtractRelayHints` provides (reads relay from e/p/a/q tags, not from `relay` tags in community definitions)
