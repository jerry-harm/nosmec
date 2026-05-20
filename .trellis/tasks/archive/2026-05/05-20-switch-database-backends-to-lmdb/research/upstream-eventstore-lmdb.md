# Research: upstream-eventstore-lmdb

- **Query**: Research whether the current upstream dependency `fiatjaf.com/nostr` provides an LMDB-backed event store backend that this repo can adopt directly without forking the entire module.
- **Scope**: mixed
- **Date**: 2026-05-20

## Findings

### Files Found

| File Path | Description |
|---|---|
| `go.mod` | Repo currently pins `fiatjaf.com/nostr v0.0.0-20260310013726-4e490879b558`. |
| `config/config.go` | Current runtime wiring uses `eventstore/boltdb` with optional `eventstore/bleve` wrapper. |
| `/home/jerry/go/pkg/mod/fiatjaf.com/nostr@v0.0.0-20260310013726-4e490879b558/eventstore/lmdb/lib.go` | Upstream LMDB backend implementation present in the currently downloaded module version. |
| `/home/jerry/go/pkg/mod/fiatjaf.com/nostr@v0.0.0-20260310013726-4e490879b558/eventstore/README.md` | Upstream README lists `lmdb` as an available implementation. |
| `/home/jerry/go/pkg/mod/fiatjaf.com/nostr@v0.0.0-20260310013726-4e490879b558/eventstore/cmd/eventstore/main.go` | Upstream CLI supports `--type lmdb`, confirming LMDB is a first-class backend. |
| `/home/jerry/go/pkg/mod/fiatjaf.com/nostr@v0.0.0-20260310013726-4e490879b558/eventstore/test/db_test.go` | Shared upstream eventstore test suite runs against LMDB and BoltDB. |
| `/home/jerry/go/pkg/mod/fiatjaf.com/nostr@v0.0.0-20260310013726-4e490879b558/eventstore/bleve/bleve_test.go` | Upstream Bleve wrapper test uses LMDB as `RawEventStore`, showing wrapper compatibility. |

### Code Patterns

An upstream `eventstore/lmdb` package does exist in the current module source tree at import path `fiatjaf.com/nostr/eventstore/lmdb`. The package defines `type LMDBBackend struct` and asserts interface compatibility with `eventstore.Store` in `eventstore/lmdb/lib.go:14-18`. Its basic initialization pattern is the same shape as the current BoltDB backend: instantiate the backend with a filesystem path and call `Init()`.

Example upstream initialization from the backend implementation and CLI:

```go
db := &lmdb.LMDBBackend{Path: path}
if err := db.Init(); err != nil {
    return err
}
```

This exact pattern is used by the upstream CLI in `eventstore/cmd/eventstore/main.go:74-80`, and the repo's current BoltDB wiring has the analogous pattern in `config/config.go:220-223`.

The current repo dependency version appears compatible because the downloaded module for `fiatjaf.com/nostr v0.0.0-20260310013726-4e490879b558` already contains `eventstore/lmdb/*` files, and the repo already depends on `github.com/PowerDNS/lmdb-go v1.9.3` in `go.mod:12`. That means adopting upstream LMDB does not require forking the module just to obtain the backend package.

Upstream also appears to test LMDB as a peer backend rather than as an experimental add-on. `eventstore/test/db_test.go:44-49` runs the common eventstore test matrix against `&lmdb.LMDBBackend{Path: dbpath + "lmdb"}`, while `eventstore/test/db_test.go:51-55` runs the same suite against BoltDB. This is evidence that LMDB is intended to satisfy the same store interface as BoltDB.

Relative to the current repo layout, the event-cache migration implication is straightforward at the raw event store layer: `config/config.go:220-231` currently builds a BoltDB raw store at `nosmec_events.db`, then optionally wraps it in a Bleve index at `search_index`. Upstream Bleve does not require BoltDB specifically; it requires an `eventstore.Store` as `RawEventStore` (`eventstore/bleve/lib.go:21-24`). Upstream test coverage explicitly wires Bleve on top of LMDB in `eventstore/bleve/bleve_test.go:16-24`, so the existing Bleve + raw-store architecture can remain two-part while swapping the raw backend from BoltDB to LMDB.

Layout differences still matter. BoltDB is opened as a single file path in `eventstore/boltdb/lib.go:34-39`, and the repo currently uses a file-like path `nosmec_events.db` in `config/config.go:220`. LMDB's `Init()` creates a directory with `os.MkdirAll(b.Path, 0755)` before opening the environment in that directory in `eventstore/lmdb/lib.go:41-47` and `eventstore/lmdb/lib.go:101-117`. So the on-disk event store path semantics change from “single file” to “directory containing LMDB files,” even though the constructor still only takes a `Path` string.

The upstream LMDB backend also exposes LMDB-specific knobs not present in the current wiring, notably `MapSize int64` and `Compact(tmppath string)` in `eventstore/lmdb/lib.go:16-19` and `eventstore/lmdb/lib.go:76-99`. By default it sets a large map size when unset via `env.SetMapSize(1 << 38)` in `eventstore/lmdb/lib.go:109-113`.

### External References

- [Go package docs: fiatjaf.com/nostr/eventstore](https://pkg.go.dev/fiatjaf.com/nostr/eventstore) — documents `lmdb` as an available implementation and exposes the shared `eventstore.Store` contract.
- [Go package docs: fiatjaf.com/nostr/eventstore/lmdb](https://pkg.go.dev/fiatjaf.com/nostr/eventstore/lmdb) — documents `LMDBBackend`, its methods, and confirms the import path.

### Related Specs

- `.trellis/spec/backend/index.md` — backend spec index likely relevant for later implementation work.

## Caveats / Not Found

I did not find evidence of an in-place data migration tool from an existing BoltDB event cache into LMDB inside the current repo. The upstream CLI can open LMDB and BoltDB stores, but this pass only verified backend availability and wiring compatibility, not a ready-made conversion path.

The package docs fetched from `pkg.go.dev` resolved to a slightly newer published pseudo-version page than the exact version pinned locally, but local module inspection confirms the pinned version already contains `eventstore/lmdb`. For version compatibility judgment, local module source is the stronger evidence here.

The dedicated web search API was unavailable in this environment due to authentication failure, so external confirmation relies on directly fetched package documentation plus the locally downloaded module source.
