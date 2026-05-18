package access

import "sync"

// RelayStream is a thread-safe round-robin URL stream.
// It cycles through a list of URLs, returning the next on each call to Next().
type RelayStream struct {
	mu    sync.Mutex
	urls  []string
	index int
}

// NewRelayStream creates a RelayStream pre-populated with the given URLs.
func NewRelayStream(urls ...string) *RelayStream {
	return &RelayStream{urls: urls}
}

// Next returns the next URL in rotation.
// Returns "" if the stream is empty.
// Wraps around to the beginning after the last URL.
// Thread-safe.
func (rs *RelayStream) Next() string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if len(rs.urls) == 0 {
		return ""
	}
	url := rs.urls[rs.index]
	rs.index = (rs.index + 1) % len(rs.urls)
	return url
}
