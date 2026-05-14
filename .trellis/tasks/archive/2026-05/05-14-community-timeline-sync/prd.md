# community-timeline-sync

## Goal

Sync note timeline features to community timeline (polls, loading more, etc).

## What I Already Know

* Community timeline uses `GetCommunityPosts` which returns a channel via `FetchMany`
* Note timeline has subscription-based polling via `SubscribeMany`
* Community timeline currently does one-shot fetch, not subscription

## Current State

The community timeline at `tui/timeline/model.go:272` uses `GetCommunityPosts` which:
- Returns a channel from `app.Pool().FetchMany`
- Is a one-shot fetch, not subscription-based
- No polling for new events
- No load more

## Requirements

* Add subscription-based polling similar to note timeline (using `SubscribeMany` with `EOSE`)
* Add load more for older posts
* This requires changing `GetCommunityPosts` to return subscription or adding a new function

## Status

Deferred - requires significant architecture changes to community fetching.