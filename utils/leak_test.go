package utils

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve_index_api.AnalysisWorker"),
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve/v2/index/scorch.(*Scorch).introducerLoop"),
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve/v2/index/scorch.(*Scorch).persisterLoop"),
		goleak.IgnoreTopFunction("github.com/blevesearch/bleve/v2/index/scorch.(*Scorch).mergerLoop"),
		goleak.IgnoreTopFunction("github.com/dgraph-io/ristretto/v2.(*defaultPolicy[...]).processItems"),
		goleak.IgnoreTopFunction("github.com/dgraph-io/ristretto/v2.(*Cache[...]).processItems"),
		goleak.IgnoreAnyFunction("fiatjaf.com/nostr.(*Pool).startPenaltyBox.func1"),
	)
}
