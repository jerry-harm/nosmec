package config

import (
	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/eventstore"
)

type StoreInterface = eventstore.Store

type RelayEvent = nostr.RelayEvent
type PublishResult = nostr.PublishResult
type ReplaceableKey = nostr.ReplaceableKey