# Backend Development Guidelines

> Best practices for backend development in this project.

---

## Overview

This directory contains guidelines for backend development. Each file covers a specific aspect of the codebase.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | Module organization, file naming | ✅ Complete |
| [Database Guidelines](./database-guidelines.md) | LMDB+Bleve store, KVStore, Close() | ✅ Complete |
| [Error Handling](./error-handling.md) | Error types, handleError, TUI errors | ✅ Complete |
| [Logging Guidelines](./logging-guidelines.md) | Structured logging, log levels | ⚠️ Template |
| [NIP Conventions](./nip-conventions.md) | NIP-19/10/65/17 output/input rules | ✅ Complete |
| [Nostr SDK Usage](./nostr-sdk-usage.md) | PubKey/EventID conversion, API reference | ✅ Complete |
| [Forked SDK Architecture](./forked-sdk-architecture.md) | Why we fork, what's forked, dependency boundary | ✅ Complete |
| [Quality Guidelines](./quality-guidelines.md) | TDD, TUI patterns, anti-patterns, NIP rules | ✅ Complete |
| [Relay Guidelines](./relay-guidelines.md) | NIP-10 e tag, relay selection, event→relay tracking | ✅ Complete |
| [AppContext](./app-context.md) | DI container, all methods and fields | ✅ Complete |
| [Query Patterns](./query-patterns.md) | Sync/async/streaming, timeout rule, GetQueryRelays | ✅ Complete |

---

## Quick Reference

| Topic | Spec File |
|-------|-----------|
| NIP-19 output format | `nip-conventions.md` |
| NIP-10 e tag format | `nip-conventions.md` + `relay-guidelines.md` |
| NIP-65 relay list | `nip-conventions.md` |
| copy() anti-pattern | `quality-guidelines.md` |
| TUI key events | `quality-guidelines.md` |
| Timeout rule | `query-patterns.md` |
| LMDB store paths | `database-guidelines.md` |
