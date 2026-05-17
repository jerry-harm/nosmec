package hints

import (
	"math"
	"sync"

	"fiatjaf.com/nostr"
)

// HintKey represents the type of relay hint being recorded.
type HintKey int

const (
	HintFetchAttempt  HintKey = 0 // tried to query, might fail (-500 pts)
	HintEventFetched  HintKey = 1 // successfully got event (700 pts)
	HintInRelayList   HintKey = 2 // author listed in kind:10002 (350 pts)
	HintFromTag       HintKey = 3 // from tag/nprofile/nevent hint (20 pts)
)

// basePoints maps each HintKey to its base score.
var basePoints = map[HintKey]float64{
	HintFetchAttempt:  -500,
	HintEventFetched:  700,
	HintInRelayList:   350,
	HintFromTag:       20,
}

// hintEntry stores a single hint record with its timestamp.
type hintEntry struct {
	key       HintKey
	timestamp nostr.Timestamp
}

// HintsDB learns which relays carry events for which pubkeys.
// Scores decay over time (^1.3) so recent hints dominate.
// Safe for concurrent use.
type HintsDB struct {
	mu   sync.RWMutex
	data map[string]map[string][]hintEntry // pubkey hex → relay URL → hints
}

// NewHintsDB creates an empty HintsDB.
func NewHintsDB() *HintsDB {
	return &HintsDB{
		data: make(map[string]map[string][]hintEntry),
	}
}

// Record stores a hint that pubkey has activity on relay.
func (h *HintsDB) Record(pubkey, relay string, key HintKey) {
	if pubkey == "" || relay == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	pkMap, ok := h.data[pubkey]
	if !ok {
		pkMap = make(map[string][]hintEntry)
		h.data[pubkey] = pkMap
	}
	pkMap[relay] = append(pkMap[relay], hintEntry{key: key, timestamp: nostr.Now()})
}

// TopN returns the top N relay URLs for a pubkey, sorted by score descending.
func (h *HintsDB) TopN(pubkey string, n int) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	pkMap, ok := h.data[pubkey]
	if !ok {
		return nil
	}

	type scoredRelay struct {
		url   string
		score float64
	}

	var relays []scoredRelay
	now := nostr.Now()

	for relayURL, entries := range pkMap {
		var score float64
		for _, e := range entries {
			bp := basePoints[e.key]
			age := now - e.timestamp
			// age can't be negative; clamp to 1 second minimum
			if age < 1 {
				age = 1
			}
			// Same decay formula as SDK: basePoints * 10^10 / (age + 86400)^1.3
			decay := math.Pow(float64(age+86400), 1.3)
			score += bp * 1e10 / decay
		}
		relays = append(relays, scoredRelay{url: relayURL, score: score})
	}

	// Sort by score descending (simple sort for n results)
	for i := 0; i < len(relays); i++ {
		best := i
		for j := i + 1; j < len(relays); j++ {
			if relays[j].score > relays[best].score {
				best = j
			}
		}
		if best != i {
			relays[i], relays[best] = relays[best], relays[i]
		}
	}

	if n > len(relays) {
		n = len(relays)
	}

	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = relays[i].url
	}
	return result
}

// Size returns the number of tracked pubkeys.
func (h *HintsDB) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.data)
}

var globalHints *HintsDB

func GlobalHints() *HintsDB {
	if globalHints == nil {
		globalHints = NewHintsDB()
	}
	return globalHints
}
