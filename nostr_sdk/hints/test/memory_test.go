package test

import (
	"testing"

	"github.com/jerry-harm/nosmec/nostr_sdk/hints/memoryh"
)

func TestMemoryHints(t *testing.T) {
	runTestWith(t, memoryh.NewHintDB())
}
