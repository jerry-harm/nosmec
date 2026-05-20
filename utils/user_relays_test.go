package utils

import (
	"reflect"
	"testing"

	"github.com/jerry-harm/nosmec/config"
	"github.com/spf13/viper"
)

func TestGetAllCandidateRelays_UsesConfiguredReadableRelaysOnly(t *testing.T) {
	t.Parallel()

	app := config.NewAppContext(nil, config.Config{
		RelayList: []config.Relay{
			{URL: "wss://relay-b.example", Read: config.BoolPtr(true), Write: config.BoolPtr(false)},
			{URL: "wss://relay-a.example", Read: config.BoolPtr(true), Write: config.BoolPtr(true)},
			{URL: "wss://write-only.example", Read: config.BoolPtr(false), Write: config.BoolPtr(true)},
		},
	}, viper.New())

	got := getAllCandidateRelays(app)
	want := []string{"wss://relay-b.example", "wss://relay-a.example"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getAllCandidateRelays() = %#v, want %#v", got, want)
	}
}
