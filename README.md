# nosmec

Nostr CLI client for power users.

## Quick Start

```bash
# Build
go build -o nosmec .

# Configure private key
# Edit ~/.config/nosmec/nosmec.yaml, set private_key (nsec format)

# Add relays
./nosmec relay add wss://relay.example.com

# Post a note
./nosmec note post "Hello, Nostr!"

# View timeline
./nosmec note timeline
./nosmec note timeline --global
./nosmec note timeline --mine
```

## Command Structure

```
nosmec
├── note        # Notes (NIP-10)
│   ├── post <content>       # Post a note
│   ├── reply <id> <content> # Reply to a note
│   ├── quote <id> <content> # Quote a note
│   └── timeline            # View timeline (global/mine/followed)
│
├── relay       # Relay management (NIP-65)
│   ├── list              # List all relays
│   ├── add <url>         # Add read/write relay
│   ├── remove <url>     # Remove relay
│   ├── set <url>         # Set relay properties
│   ├── publish           # Publish Kind 10002
│   ├── sync              # Sync relay list from network
│   ├── fetch <pubkey>   # Fetch someone's relay list
│   ├── dm               # DM relay management (NIP-17)
│   │   ├── add <url>
│   │   ├── remove <url>
│   │   ├── list
│   │   └── publish       # Publish Kind 10050
│   └── search           # Search relay management
│       ├── add <url>
│       ├── remove <url>
│       └── list
│
├── subscribe   # Subscription management (NIP-02, NIP-51)
│   ├── add <community|user|hashtag> <identifier>
│   ├── remove <community|user|hashtag> <identifier>
│   ├── list [community|user|hashtag]
│   ├── sync              # Sync from network
│   └── publish          # Publish to network
│
├── profile     # Profile management
│   ├── set <name> <about> <picture>
│   └── get [pubkey]
│
├── community   # Community (NIP-72)
│   ├── list            # List communities
│   ├── create <name> <desc>
│   ├── join <community-id>
│   └── post <content>
│
├── alias       # Alias management
│   ├── list
│   ├── add <name> <npub-or-hex>
│   └── remove <name>
│
└── dm         # Direct messages (NIP-17)
    ├── list              # List conversations
    ├── send <npub> <msg> # Send DM
    └── recv              # Receive DMs (polls)
```

## Configuration

Config file: `~/.config/nosmec/nosmec.yaml`

### Environment Variables

All config can be overridden with `NOSMEC_` prefix:

| Config Key | Env Variable |
|------------|--------------|
| `private_key` | `NOSMEC_PRIVATE_KEY` |
| `relay_list` | `NOSMEC_RELAY_LIST` |
| `dm_relays` | `NOSMEC_DM_RELAYS` |

### Proxy Support

`proxy.socks` and `proxy.i2p_socks` are available. Both are SOCKS5 proxies.
- `socks` handles all traffic if `i2p_socks` is not set; otherwise non-.i2p traffic.
- `i2p_socks` only routes `.i2p` domains.
- Onion (.onion) support: put your Tor proxy address in `proxy.socks` (e.g., `127.0.0.1:9050`).

## Supported NIPs

| NIP | Name | Status |
|-----|------|--------|
| NIP-01 | Basic Protocol | ✓ |
| NIP-02 | Follow List (Kind 3) | ✓ |
| NIP-05 | NIP-05 Verification | ✓ |
| NIP-06 | Key Formats (nsec/npub) | ✓ |
| NIP-10 | Reply Conventions | ✓ |
| NIP-17 | DM Relay List (Kind 10050) | ✓ |
| NIP-19 | Bech32 Encoded Entities | ✓ |
| NIP-21 | `nostr:` URL Scheme | ✓ |
| NIP-40 | Expiration Timestamp | ✓ |
| NIP-44 | NIP-44 Encryption | ✓ |
| NIP-51 | Lists (10003, 10004, 10015) | ✓ |
| NIP-65 | Relay List Metadata (Kind 10002) | ✓ |
| NIP-72 | Community Boards (Kind 34550, 1111) | ✓ |
| NIP-46 | Remote Signing | Planned |
| NIP-47 | Nostr Wallet Connect | Planned |

## Development

```bash
# Build
go build -o nosmec .

# Run
go run .

# Test
go test ./...

# Update dependencies
go mod tidy
```

## Project Structure

```
nosmec/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command
│   ├── note_commands.go   # Note commands (Kind 1)
│   ├── event_commands.go  # Generic event commands (all kinds)
│   ├── relay_commands.go  # Relay management (NIP-65, NIP-17)
│   ├── search_commands.go # Search commands (NIP-50)
│   ├── gossip_commands.go # Gossip relay list
│   ├── config_commands.go # Config management
│   ├── profile_commands.go # Profile commands (Kind 0)
│   ├── community_commands.go # Community commands (NIP-72)
│   ├── dm_commands.go     # DM commands (NIP-17)
│   ├── registry.go        # Command registration
│   ├── errors.go          # Error types
│   └── completion/        # Shell completion
│
├── config/                # Configuration management
│   ├── config.go         # Viper initialization
│   ├── types.go          # Type definitions
│   ├── relay.go          # Relay configuration
│   ├── context.go        # AppContext (DI container)
│   └── interfaces.go     # StoreInterface, etc.
│
├── utils/                 # Business logic
│   ├── get.go            # Querying (GetEvent, GetProfile, GetTimeline)
│   ├── post.go           # Publishing (PostNote, Reply, Quote)
│   ├── profile.go         # Profile operations
│   ├── community.go       # Community operations (NIP-72)
│   ├── subscription.go    # Subscription/follow (NIP-02, NIP-51)
│   ├── dm.go             # DM operations (NIP-17, nip59 GiftWrap)
│   ├── relay_list.go     # Relay list publish/parse
│   ├── user_relays.go    # NIP-65 discovery, GetQueryRelays
│   ├── search.go         # NIP-50 search
│   ├── filters.go        # Pure nostr.Filter builders (testable)
│   ├── alias.go          # Alias management
│   ├── show.go           # Display formatting (NIP-19 bech32)
│   ├── sync.go           # Sync from network
│   ├── proxy.go          # SOCKS/I2P proxy support
│   └── *_test.go         # Test files alongside implementation
│
├── tui/                   # Terminal UI (BubbleTea v2)
│   ├── timeline/         # Timeline view + list
│   ├── compose/          # Note/DM compose
│   ├── thread/           # Thread view with treeview
│   ├── event/            # Event detail view
│   ├── dm/               # DM list + chat
│   ├── community/        # Community view
│   ├── bubblon/          # Window management (bubblon.Controller)
│   └── cmd/              # TUI command registry
│
├── logger/                # Structured logging (slog)
│
└── docs/                  # Documentation
    ├── README.md         # This file
    ├── DEV.md            # Development guide
    ├── NIP.md           # NIP protocol reference
    ├── CONFIG.md         # Configuration details
    └── RELAY.md          # Relay management details
```

## Supported NIPs
