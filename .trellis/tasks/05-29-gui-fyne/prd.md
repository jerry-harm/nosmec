# brainstorm: Fyne GUI implementation

## Goal

Build a community-first desktop GUI for nosmec (Fyne), inspired by Baidu Tieba / Reddit / 4chan model. The top bar exposes `Community`, `DM`, and `Note` buttons, but MVP only implements the `Community` main experience; `DM` and `Note` are button-level placeholders for future work. `Profile` and `Settings` stay visible in the top bar but have no interaction yet, and will become dedicated pages later. Replace Bubble Tea TUI with a proper desktop shell.

## Layout

```
┌─────────────────────────────────────────────────────────────┐
│  [Community]  [DM]  [Note]    🔍 Search...     [👤] [⚙️]   │  ← Top bar (mode switcher)
├──────────────┬──────────────────────────────────────────────┤
│  ▼ My Feed  │              Main Content                    │
│  ▼ Global   │          (post cards + reply cards)         │
│  ▼Communities│                                             │
│    community1│                                             │
│    community2│                                             │
└──────────────┴──────────────────────────────────────────────┘
```

### Top Bar
- **Left**: Mode buttons — `Community`, `DM`, `Note`
- **Center**: Search input (placeholder, non-functional)
- **Right**: Profile button + Settings button (visible only, no interaction)

### Community Mode Sidebar (3 sections, collapsible)

**My Feed** — all posts from all followed communities, mixed, sorted by time

**Global** — all community events across all known communities

**Communities** — followed community list
- Each entry = community name
- Click → main view filters to that community's posts
- If no followed communities, show empty state

## Post Card System

Post cards are the primary UI element. There are two display modes that show the **same content with different spatial arrangement**:

### Card Content (per post)
- Author npub (truncated, e.g. `npub1abc...def`)
- Community name (e.g. `34550:abc...def/communities/community-name`)
- Post body preview (first ~100 chars)
- Reply count, timestamp

### Nested Replies (inside a card)
- Show up to 3 replies inline (either latest or most-liked — static for MVP)
- Each reply is one line: author snippet + body preview
- If there are more than 3 replies, show "N more replies" text
- Clicking the card enters the **Horizontal thread view** for that post

### Horizontal Mode (default for MVP)
- Layout: left-to-right or top-to-bottom stacking of cards (implementation decides)
- Each top-level post card shows up to 3 inline replies nested inside it
- Visual: card-within-card (reply cards indented inside the parent card)
- Click card → enters the full thread view (same layout, but showing the full reply chain)

### Vertical Mode (future, not MVP)
- Two-person discussion threads
- Small cards show reply-to-post content
- May reuse DM interface later
- **Not implemented in MVP**

## Visual Style

- Card-based layout, no multi-level indentation in list view
- Horizontal mode: reply cards visually embedded inside parent card
- Cards have clear borders, slight shadow or background color differentiation
- Mock data only — no real Nostr events

## Design Principles

* **Community-first**: Not a personal timeline client; communities are the primary navigation unit
* **Reference**: Tieba/Reddit/4chan mental model — forums, threads, nested replies
* **Desktop-native**: Proper window chrome, minimum size constraints, desktop patterns
* **MVP only**: Static mock data, no real network, just navigation and display wiring
* **Recursive card pattern**: Horizontal thread view is the same component reused at full depth

## Modes

### Community Mode
* Top bar + sidebar as above
* Main = card list with inline replies (horizontal mode)
* Click card → horizontal thread view (same component, deeper)

### DM Mode (future)
* Top bar button exists only
* No mode implementation in MVP

### Note Mode (future)
* Top bar button exists only
* No mode implementation in MVP

## Requirements

* Community mode fully functional: top bar, sidebar navigation, card list display
* Default Community view = `My Feed` (mixed posts from followed communities)
* Cards show: author, community name, body preview, reply count, timestamp
* Cards show up to 3 inline replies inside each card
* Horizontal thread view reachable by clicking a card
* DM/Note/Profile/Settings top bar buttons exist as visual placeholders only
* Search bar is visible in top bar as a non-functional placeholder
* Sidebar sections collapsible
* Mock post data with realistic content

## Acceptance Criteria

* [ ] Top bar shows Community/DM/Note buttons
* [ ] Search bar is visible in top bar
* [ ] Profile and Settings buttons are visible in top bar
* [ ] Community sidebar: My Feed + Global + community list (3 sections)
* [ ] Initial app state opens Community mode with `My Feed` selected
* [ ] Click My Feed → main shows card list (mixed posts)
* [ ] Click community name → main shows that community's cards
* [ ] Click Global → main shows all community cards
* [ ] Each card shows author, community, preview, reply count, timestamp
* [ ] Each card shows up to 3 inline replies inside it
* [ ] Clicking a card switches to horizontal thread view
* [ ] Window minimum size ≥ 800×600
* [ ] No locale warning on startup
* [ ] Build passes

## Definition of Done

* `nosmec gui` launches without errors or locale warnings
* Community mode fully navigable with mock data
* Card list with inline replies visible
* Horizontal thread view reachable from card click
* DM and Note buttons visible in top bar for future expansion
* Profile and Settings buttons visible in top bar for future expansion
* Search bar visible for future expansion
* Build `go build ./...` passes

## Out of Scope (MVP)

* Real Nostr relay connections or network requests
* User authentication / key management
* Database persistence
* Search functionality
* Person view (see all posts by a user)
* Unfollowed community browsing
* Vertical mode / two-person discussion threads
* DM mode implementation
* Note mode implementation
* Profile page
* Settings page

## Technical Notes

* Fyne v2.7.4, `fyne.io/fyne/v2`
* `container.NewHSplit` for sidebar + content layout
* Mode stored as `currentMode` variable (enum: community/dm/note)
* Sidebar selected item stored as `selectedItem` (enum: myfeed/global/<community-name>)
* Mock post data as package-level vars in `gui/` package
* Collapsible sections using manual collapse state (no Tree widget needed for MVP)
* Community list from `config.Subscriptions` where `Type == "community"` (for future real impl)
* Post card widget: custom `fyne.Widget` using `fyne.Container` with `Border` layout
* Reply preview: up to 3 items, simple `widget.Label` rows inside card
* Thread view: same card component reused, entry via click navigation (state machine: list → thread)