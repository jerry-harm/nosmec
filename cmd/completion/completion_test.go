package completion

import "testing"

func TestConfigKeyCompletionFunc_DoesNotExposeKnownRelays(t *testing.T) {
	t.Parallel()

	completions, _ := ConfigKeyCompletionFunc(nil, nil, "known")
	for _, completion := range completions {
		if completion == "known_relays" {
			t.Fatalf("ConfigKeyCompletionFunc() unexpectedly exposed removed key %q", completion)
		}
	}
}
