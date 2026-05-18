package access

import (
	"testing"
)

func TestRelayStream_Next_Rotation(t *testing.T) {
	rs := NewRelayStream("wss://a.example.com", "wss://b.example.com", "wss://c.example.com")

	if got := rs.Next(); got != "wss://a.example.com" {
		t.Errorf("Next() = %q, want %q", got, "wss://a.example.com")
	}
	if got := rs.Next(); got != "wss://b.example.com" {
		t.Errorf("Next() = %q, want %q", got, "wss://b.example.com")
	}
	if got := rs.Next(); got != "wss://c.example.com" {
		t.Errorf("Next() = %q, want %q", got, "wss://c.example.com")
	}
	// wraps around
	if got := rs.Next(); got != "wss://a.example.com" {
		t.Errorf("Next() = %q, want %q (wrap)", got, "wss://a.example.com")
	}
}

func TestRelayStream_Next_Empty(t *testing.T) {
	rs := NewRelayStream()
	if got := rs.Next(); got != "" {
		t.Errorf("Next() = %q, want empty", got)
	}
	if got := rs.Next(); got != "" {
		t.Errorf("Next() = %q, want empty", got)
	}
}

func TestRelayStream_Next_Single(t *testing.T) {
	rs := NewRelayStream("wss://only.example.com")
	for i := range 5 {
		if got := rs.Next(); got != "wss://only.example.com" {
			t.Errorf("iteration %d: Next() = %q, want %q", i, got, "wss://only.example.com")
		}
	}
}

func TestRelayStream_Next_Concurrent(t *testing.T) {
	rs := NewRelayStream("wss://a.example.com", "wss://b.example.com")
	const n = 100
	results := make(chan string, n)

	for range n {
		go func() {
			results <- rs.Next()
		}()
	}

	counts := map[string]int{}
	for range n {
		counts[<-results]++
	}
	if counts["wss://a.example.com"] == 0 {
		t.Error("expected wss://a.example.com to be served")
	}
	if counts["wss://b.example.com"] == 0 {
		t.Error("expected wss://b.example.com to be served")
	}
}
