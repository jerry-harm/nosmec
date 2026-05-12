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
| `known_relays` | `NOSMEC_KNOWN_RELAYS` |

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
│   ├── note.go            # Note commands
│   ├── relay.go           # Relay management
│   ├── subscribe.go       # Subscription management
│   ├── profile.go         # Profile commands
│   ├── community.go       # Community commands
│   ├── alias.go           # Alias commands
│   └── dm.go              # DM commands
│
├── config/                # Configuration management
│   ├── config.go         # Viper initialization
│   ├── types.go          # Type definitions
│   └── relay.go          # Relay configuration
│
├── utils/                 # Utility functions
│   ├── post.go           # Publishing (PostNote, Reply, Quote)
│   ├── get.go            # Querying (GetTimeline, GetEvent)
│   ├── profile.go         # Profile operations
│   ├── community.go       # Community operations
│   ├── subscription.go    # Subscription operations
│   ├── dm.go             # DM operations
│   └── alias.go          # Alias operations
│
├── logger/                # Logging utilities
├── tui/                   # TUI (work in progress)
│   ├── timeline/         # Timeline TUI (needs rework)
│   └── common/          # Common TUI components
│
└── docs/                  # Documentation
```

## Known Issues

- TUI timeline is incomplete and needs rework
- DM functionality needs testing
- NIP-46 Remote Signing not implemented

## License

MIT
