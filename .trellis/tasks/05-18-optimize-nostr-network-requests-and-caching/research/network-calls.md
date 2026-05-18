# Research: Nostr Network Request Patterns

- **Query**: Map all nostr relay network requests — find every call site, trace call chains, identify redundancies
- **Scope**: internal
- **Date**: 2026-05-18

## Findings

### Complete Network Call Inventory

#### A. Pool().QuerySingle — ~19 distinct call sites

Single-event queries (returns first result from relay list).

| File:Line | Function | Fetches | Triggers |
|---|---|---|---|
| `utils/get.go:60` | `GetEvent` (replaceable, local first) | Replaceable event from local relay | `GetNote`, `GetCommunity`, `GetFollowedCommunities`, `GetPost` |
| `utils/get.go:71` | `GetEvent` (replaceable, all relays) | Replaceable event, fallback | Same as above (after local miss) |
| `utils/get.go:82` | `GetEvent` (non-replaceable, local first) | Non-replaceable event from local | `GetNote`, `GetCommunity`, etc. |
| `utils/get.go:92` | `GetEvent` (non-replaceable, all relays) | Non-replaceable event, fallback | Same as above |
| `utils/get.go:189` | `GetProfile` | Kind 0 profile metadata | Everywhere a profile name is needed |
| `utils/get.go:347` | `GetParentEvent` | Parent event by ID | Thread navigation |
| `utils/get.go:401` | `GetProfileAsync` | Kind 0 profile (async) | TUI event view, DM TUI |
| `utils/get.go:490` | `GetEventAsync` | Any event by filter (async) | TUI event view from ID |
| `tui/thread/thread.go:418` | `fetchEventByID` | Single event by ID | Thread parent chain walking |
| `utils/subscription.go:108` | `syncCommunitiesFromNetwork` | Kind 10004 community list | CLI `sync` command |
| `utils/subscription.go:153` | `syncUsersFromNetwork` | Kind 3 follow list | CLI `sync` command |
| `utils/subscription.go:210` | `syncHashtagsFromNetwork` | Kind 10015 interest list | CLI `sync` command |
| `utils/profile.go:224` | `FetchRecipientDMRelays` | Kind 10050 DM relay list | `SendDM`, `GetFullProfile` |
| `utils/profile.go:253` | `FetchRecipientReadRelays` | Kind 10002 relay list metadata | `SendDM` (fallback) |
| `utils/profile.go:322` | `GetFullProfile` | Kind 10002 relay list | `cmd profile --full` |
| `utils/profile.go:359` | `GetFullProfile` | Kind 3 follow list | `cmd profile --full` |
| `utils/profile.go:387` | `GetFullProfile` | Kind 10004 community list | `cmd profile --full` |
| `utils/profile.go:405` | `GetFullProfile` | Kind 10015 interest list | `cmd profile --full` |
| `utils/relay_list.go:47` | `syncRelayListFromNetwork` | Kind 10002 relay list | CLI `sync relays` |
| `utils/relay_list.go:113` | `syncDMRelaysFromNetworkImpl` | Kind 10050 DM relay list | CLI `sync relays` |
| `cmd/config_commands.go:288` | config relay sync command | Kind 10002 relay list | CLI `config sync-relays` |

#### B. Pool().FetchMany — ~10 call sites

Multi-event queries (streams all results).

| File:Line | Function | Fetches | Triggers |
|---|---|---|---|
| `utils/get.go:302` | `QueryRepliesToRoot` | All replies (kind 1 + 1111) | Thread view |
| `utils/get.go:523` | `GetMyTimeline` | User's notes | Timeline TUI (mine) |
| `utils/get.go:552` | `GetGlobalTimeline` | Global notes | Timeline TUI (global) |
| `utils/get.go:608` | `GetFollowedTimeline` | Notes from followed + communities | Timeline TUI (followed) |
| `utils/community.go:41` | `FetchCommunityEvents` | Kind 34550 community definions | Community discover TUI |
| `utils/community.go:311` | `GetCommunityPosts` | Kind 1111 comments for community | Timeline TUI (community) |
| `utils/community.go:412` | `GetMyCreatedCommunities` | Kind 34550 (filtered by author) | CLI `community list` |
| `utils/community.go:437` | `GetPostedCommunities` | Kind 1111 comments (by author) | CLI `community list` |
| `utils/search.go:101` | `SearchEvents` | NIP-50 text search | CLI `search` |
| `tui/thread/thread.go:463` | `fetchThreadReplies` | Kind 1+1111 replies (batching by depth) | Thread view |

#### C. Pool().FetchManyReplaceable — ~4 call sites

Multi-event replaceable queries (returns latest per d-tag pubkey pair).

| File:Line | Function | Fetches | Triggers |
|---|---|---|---|
| `utils/get.go:71` | `GetEvent` (replaceable) | Latest replaceable event | `GetNote` for replaceable kinds |
| `utils/get.go:427` | `GetProfiles` | Batch profiles (kind 0) | NOT currently used in TUI |
| `utils/get.go:483` | `GetEventAsync` (replaceable) | Latest replaceable event (async) | TUI event view |
| `utils/user_relays.go:56` | `DiscoverUserRelays` | Kind 10002 relay list | Async goroutine from `GetProfile`/`GetProfileAsync` |

#### D. Pool().SubscribeMany — ~3 call sites

Ongoing subscriptions (real-time streaming).

| File:Line | Function | Fetches | Triggers |
|---|---|---|---|
| `utils/get.go:104` | `SubscribeWithCache` (wrapper) | Any filter + caches results | Timeline TUI subscription, DM TUI subscription |
| `utils/dm.go:131` | `ListDMConversations` | Kind 1059 gift wraps (to user) | DM conversation list |
| `utils/dm.go:224` | `QueryDMHistory` | Kind 1059 gift wraps (to/from) | DM history query |

#### E. Pool().PublishMany — ~12 call sites

Write operations (publish events to writable relays).

| File:Line | Function | Publishes | Triggers |
|---|---|---|---|
| `utils/post.go:33` | `PostNote` | Kind 1 note | CLI `note post` / TUI compose |
| `utils/post.go:86` | `ReplyNote` | Kind 1 reply | TUI compose reply |
| `utils/post.go:145` | `QuoteNote` | Kind 1 quote | TUI compose quote |
| `utils/post.go:179,188` | `DeleteNote` | Kind 5 deletion | TUI event view delete |
| `utils/subscription.go:292,329,362` | `publishFollowList/CommunityList/InterestList` | Kind 3/10004/10015 | CLI `publish` |
| `utils/relay_list.go:177,210` | `publishRelayListMetadata/DMRelayList` | Kind 10002/10050 | CLI `publish` |
| `utils/profile.go:169` | `SetProfile` | Kind 0 profile | CLI `profile set` |
| `utils/community.go:137,206` | `CreateCommunity`/`PostToCommunity` | Kind 34550/1111 | CLI `community create`/`post` |
| `tui/compose/model.go:509` | Compose send | Kind 1 note (via PostNote) | TUI compose |

#### F. NIP-17 helper calls

| File:Line | Function | Notes |
|---|---|---|
| `utils/dm.go:47` | `nip17.PublishMessage` | Gift-wrap + publish to our+their relays |
| `utils/dm.go:81` | `nip17.ListenForMessages` | Subscribe to gift-wraps on DM relays |

### Call Chain Trace: TUI Timeline Load

```
CLI: nosmec note timeline (or note timeline --global / --mine)
  -> cmd/note_commands.go:44  timeline.RunTimeline(app, filter, ...)
    -> tui/timeline/model.go:249  m.fetchTimeline()
      -> utils/get.go:552  GetGlobalTimeline() or GetMyTimeline() or GetFollowedTimeline()
        -> Pool().FetchMany(ctx, relays, filter, opts)  // ONE relay query

  -> tui/timeline/model.go:287  m.fetchProfileNames(pubkeys)
      -> [PER-PUBKEY] utils/get.go:298  GetProfileName() -> GetProfile() -> QuerySingle()
      -> N separate relay queries for N unique authors!

  -> tui/timeline/model.go:451  m.startSubscription()
      -> utils/get.go:104  SubscribeWithCache() -> SubscribeMany()
      -> ONE ongoing subscription query

  -> tui/timeline/model.go:310  m.fetchMoreOld() [on pagination]
      -> Same FetchMany as above but with `Until` filter
```

### Call Chain Trace: Event Detail View

```
CLI: nosmec event <id>
  -> cmd/event_commands.go:55  RunEventDetail()
    -> tui/event/event.go:103  NewFromID(eventID, ...)
      -> tui/event/event.go:174  fetchEventAsync()
        -> utils/get.go:215  GetNoteAsync() -> GetEventAsync() -> QuerySingle()
        // [#1] One relay query for the event

      -> tui/event/event.go:183  fetchProfileNameAsync() [sequentially after event loaded]
        -> utils/get.go:357  GetProfileNameAsync() -> GetProfileAsync() -> QuerySingle()
        // [#2] One relay query for the author profile

  -> User presses "t" (thread view):
    -> tui/thread/thread.go:304  fetchThread()
      -> m.fetchRootEvent() -> fetchEventByID() -> QuerySingle()  // [#3]
      -> m.fetchParentChain() -> fetchEventByID() -> QuerySingle() x N  // [#4..N+3]
      -> m.fetchThreadReplies() -> FetchMany() x M batches  // [#N+4]
      -> m.fetchProfileNames() -> GetProfileName() x K per pubkey  // [#N+5..]
```

### Call Chain Trace: DM Thread

```
CLI: nosmec dm <npub>
  -> tui/dm/main.go:13  RunDM()
    -> tui/dm/model.go:120  Init()
      -> fetchHistory() -> QueryDMHistory() -> SubscribeMany()  // [#1]
      -> startSubscription() -> SubscribeWithCache() -> SubscribeMany()  // [#2] DUPLICATE!
      -> fetchRecipientProfileNameAsync() -> GetProfileNameAsync() -> QuerySingle()  // [#3]
```

### Call Chain Trace: Community Discover

```
CLI: nosmec community discover
  -> tui/community/discover/model.go:126  loadCommunities()
    -> utils/community.go:25  FetchCommunityEvents()
      -> Pool().FetchMany(ctx, relays, filter)  // [#1] One call
```

### Call Chain Trace: Search Output (profile names per result)

```
CLI: nosmec search "query"
  -> utils/search.go:75  SearchEvents() -> FetchMany()  // [#1]
  -> cmd/search_commands.go:91  printSearchResult() per result
    -> utils/profile.go:91 (from note_commands.go pattern)  GetProfileName() -> GetProfile() -> QuerySingle()
    -> N individual profile queries for N search results
```

### Call Chain Trace: Gossip (Relay Discovery)

```
CLI: nosmec gossip
  -> cmd/gossip_commands.go:27  runGossip()
    -> [PER-FOLLOWED-USER] utils/user_relays.go:41  DiscoverUserRelays()
      -> Pool().FetchManyReplaceable(ctx, relays, filter)  // One per followed user
      -> N separate queries for N followed users
```

---

## TOP 5 Most Redundant/Expensive Patterns

### 1. Profile fetching is N+1 — NOT using batch `GetProfiles`

**Location**: `tui/timeline/model.go:287-308`, `tui/thread/thread.go:577-615`

**What happens**: After fetching a timeline (50+ events from multiple authors), `fetchProfileNames` spawns a goroutine **per pubkey** that individually calls `GetProfileName` → `GetProfile` → `QuerySingle`. For 50 unique authors = 50+ separate relay round-trips.

**What already exists but is unused**: `utils/get.go:409` `GetProfiles` accepts a slice of pubkeys and uses `FetchManyReplaceable` in one relay call — but NO TUI code calls it. `GetProfiles` has zero callers outside its own test.

### 2. DM Thread subscribes TWICE to the same relay filter

**Location**: `tui/dm/model.go:120-126` (`Init`)

`fetchHistory()` calls `QueryDMHistory` which does `SubscribeMany` with the filter. **Simultaneously**, `startSubscription()` calls `SubscribeWithCache` which does another `SubscribeMany` with the **exact same filter** (kind 1059 gift wraps, same p-tags). Two subscriptions to the same data stream on the same relays — the history one gets them first, the subscription one is redundant for the initial load.

### 3. Thread parent chain walks one-at-a-time (sequential, not batch)

**Location**: `tui/thread/thread.go:369-401` (`fetchParentChain`)

For a deep thread (e.g. 5 parents), `fetchParentChain` loops and calls `fetchEventByID` → `QuerySingle` for each parent **sequentially**. Each call waits for a full relay timeout. Could batch all IDs into one `FetchMany` with `IDs: [...]` filter.

### 4. GetFullProfile makes 5-6 sequential QuerySingle calls

**Location**: `utils/profile.go:279-415` (`GetFullProfile`)

Called from `cmd profile --full`. Makes these queries sequentially:
1. `GetProfile` (kind 0 profile metadata)
2. Kind 10002 relay list (via `QuerySingle`)
3. `FetchRecipientDMRelays` (kind 10050, another `QuerySingle`)
4. Kind 3 follow list (via `QuerySingle`)
5. Kind 10004 community list (via `QuerySingle`)
6. Kind 10015 interest list (via `QuerySingle`)

All query the same relay set. Could be parallelized.

### 5. Timeline repeats the same FetchMany on refresh and pagination

**Location**: `tui/timeline/model.go:249` (`fetchTimeline`) and `tui/timeline/model.go:310` (`fetchMoreOld`)

On refresh (key "r"), `fetchTimeline` re-queries everything from scratch — even though `startSubscription` is already capturing new events in real-time. The refresh re-fetches the same events again from all relays. Furthermore, the subscription filter (`startSubscription`) and the initial fetch (`fetchTimeline`) query the same relay set with overlapping filters — the initial FetchMany returns events that the subscription will also receive.

### 6. (Bonus) Search output fetches profile per result

**Location**: `cmd/search_commands.go:91` (inside `printSearchResult`)

For each search result, a `GetProfileName` → `GetProfile` → `QuerySingle` call is made. 50 search results = 50 individual profile queries.

### 7. (Bonus) DM conversation list fetches profile per conversation

**Location**: `tui/dm/list/model.go:125-156` (`loadConversations`)

Calls `ListDMConversations` (SubscribeMany), then for each conversation item, calls `GetProfileName` → `GetProfile` → `QuerySingle`. 20 conversations = 20 profile queries.

---

## Existing Batching/Deduplication

| Mechanism | Code Location | What it does |
|---|---|---|
| `GetProfiles` (batch profile fetch) | `utils/get.go:409` | Accepts `[]PubKey`, uses `FetchManyReplaceable`. **Zero callers in TUI.** |
| `SubscribeWithCache` | `utils/get.go:103` | Wraps `SubscribeMany` with `CacheEvent` before forwarding to channel |
| `CacheEvent` | `utils/get.go:135` | Pushes event to local relay for future reuse |
| Local relay first pattern | `utils/get.go:56-77` | `GetEvent` tries local relay (2s timeout) before all relays |
| `DiscoverUserRelays` async | `utils/get.go:181,394` | Runs in background goroutine during `GetProfile`/`GetProfileAsync` |
| `seenEventIDs` dedup | `tui/timeline/model.go:112` | Prevents duplicate event insertion in list |
| `FetchManyReplaceable` for replaceable | `utils/get.go:71,427,483` | Uses pool's native replaceable dedup |
| `RelayStream` (round-robin) | `access/relay_stream.go` | Thread-safe relay rotation. **Exists but NOT used by any query code** — all callers manually build relay lists with `AllReadableRelays()`, `KnownRelays`, etc. |
| `HintsDB` + `GetQueryRelays` | `utils/user_relays.go:130` | Priority-based relay selection from tag hints + outbox hints |
| `System.TrackEventRelay` (first-write-wins) | `access/system.go:64` | Event-to-relay mapping persistence |
| `AppContext.QueryTimeout()` | `config/context.go:155` | Centralized timeout configuration |
| Rate limiting on refresh | `tui/timeline/model.go:252` | 2-second cooldown between refreshes |

## Caveats / Not Found

- **RelayStream is unused**: The `access/RelayStream` round-robin abstraction exists but no query code uses it — all callers construct relay lists ad-hoc from `AllReadableRelays()` + `KnownRelays`. This is an abstraction that was built but never integrated.
- **No in-memory profile cache**: Every `GetProfile` call hits the relay. There's no TTL-based in-memory cache. The `shouldCache`/`CacheEvent` mechanism only writes to local relay — it doesn't prevent re-fetches within the same session.
- **No request coalescing**: If two TUI views need the same profile simultaneously, both hit the relay independently.
- **`SubscriptionOptions.Label` inconsistent**: Some `SubscribeMany` calls set `Label` ("timeline", "dmconversations", "dm-tui", "dmhistory"), others don't. This affects the nostr pool's ability to dedup subscriptions server-side.
- **FetchManyReplaceable channel consumption**: In `GetEvent` (line 71), the `FetchManyReplaceable` results channel is consumed inline with `results.Range` which blocks until the channel closes. This means the function waits for ALL relays to respond (or timeout) even though we only need the first result.

### Related Specs

- `.trellis/spec/backend/index.md` — Backend coding guidelines
- `.trellis/spec/tui/index.md` — TUI layer guidelines
