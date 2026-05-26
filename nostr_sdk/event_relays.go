package nostr_sdk

import (
	"slices"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/nostr_sdk/kvstore"
)

const eventRelayPrefix = byte('r')

// makeEventRelayKey creates a key for storing event relay information.
// It uses the first 8 bytes of the event ID to create a compact key.
func makeEventRelayKey(id nostr.ID) []byte {
	// format: 'r' + first 8 bytes of event ID
	key := make([]byte, 9)
	key[0] = eventRelayPrefix
	copy(key[1:], id[:8])
	return key
}

// encodeRelayList serializes a list of relay URLs into a compact binary format.
// Each relay URL is prefixed with its length as a single byte.
func encodeRelayList(relays []string) []byte {
	totalSize := 0
	for _, relay := range relays {
		if len(relay) > 256 {
			continue
		}
		totalSize += 1 + len(relay) // 1 byte for length prefix
	}

	buf := make([]byte, totalSize)
	offset := 0

	for _, relay := range relays {
		if len(relay) > 256 {
			continue
		}
		buf[offset] = uint8(len(relay))
		offset += 1
		copy(buf[offset:], relay)
		offset += len(relay)
	}

	return buf
}

// decodeRelayList deserializes a binary-encoded list of relay URLs.
// It expects each relay URL to be prefixed with its length as a single byte.
func decodeRelayList(data []byte) []string {
	relays := make([]string, 0, 6)
	offset := 0

	for offset < len(data) {
		if offset+1 > len(data) {
			return nil // malformed
		}

		length := int(data[offset])
		offset += 1

		if offset+length > len(data) {
			return nil // malformed
		}

		relay := string(data[offset : offset+length])
		relays = append(relays, relay)
		offset += length
	}

	return relays
}

// trackEventRelay records that an event was seen on a particular relay.
// If onlyIfItExists is true, it will only update existing records and not create new ones.
func (sys *System) trackEventRelay(id nostr.ID, relay string, onlyIfItExists bool) {
	// get the key for this event
	key := makeEventRelayKey(id)

	// update the relay list atomically
	sys.KVStore.Update(key, func(data []byte) ([]byte, error) {
		var relays []string
		if data != nil {
			relays = decodeRelayList(data)

			// check if relay is already in list
			if slices.Contains(relays, relay) {
				return nil, kvstore.NoOp // no change needed
			}

			// append new relay
			relays = append(relays, relay)
			return encodeRelayList(relays), nil
		} else if onlyIfItExists {
			// when this flag exists and nothing was found we won't create anything
			return nil, kvstore.NoOp
		} else {
			// nothing exists, so create it
			return encodeRelayList([]string{relay}), nil
		}
	})
}

// GetEventRelays returns all known relay URLs an event is known to be available on.
// It is based on information kept on KVStore.
func (sys *System) GetEventRelays(id nostr.ID) []string {
	// get the key for this event
	key := makeEventRelayKey(id)

	// get stored relay list
	data, _ := sys.KVStore.Get(key)
	if data == nil {
		return nil
	}

	return decodeRelayList(data)
}

func (sys *System) ListKnownEventRelays() ([]string, error) {
	if sys == nil || sys.KVStore == nil {
		return []string{}, nil
	}

	seen := make(map[string]struct{}, 8)
	if sys.Hints != nil {
		hintRelays, err := sys.Hints.GetAllKnownRelays()
		if err != nil {
			return nil, err
		}
		for _, relay := range hintRelays {
			if relay == "" {
				continue
			}
			seen[relay] = struct{}{}
		}
	}

	err := sys.KVStore.Iterate(func(key, value []byte) error {
		if len(key) != 9 || key[0] != eventRelayPrefix {
			return nil
		}
		for _, relay := range decodeRelayList(value) {
			if relay == "" {
				continue
			}
			seen[relay] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	known := make([]string, 0, len(seen))
	for relay := range seen {
		known = append(known, relay)
	}
	slices.Sort(known)

	return known, nil
}

func (sys *System) EraseEventRelays(id nostr.ID) error {
	key := makeEventRelayKey(id)
	return sys.KVStore.Delete(key)
}
