package utils

import (
	"sort"
	"testing"
	"time"

	"fiatjaf.com/nostr"
)

func TestConversation_Struct(t *testing.T) {
	conv := Conversation{
		PubKey:   "abcd1234",
		LatestDM: DMMessage{
			Content:   "Hello!",
			FromMe:    false,
			Timestamp: nostr.Timestamp(time.Now().Unix()),
		},
		LatestAt: nostr.Timestamp(time.Now().Unix()),
	}

	if conv.PubKey != "abcd1234" {
		t.Errorf("conv.PubKey = %q, want %q", conv.PubKey, "abcd1234")
	}
	if conv.LatestDM.Content != "Hello!" {
		t.Errorf("conv.LatestDM.Content = %q, want %q", conv.LatestDM.Content, "Hello!")
	}
	if conv.LatestDM.FromMe != false {
		t.Errorf("conv.LatestDM.FromMe = %v, want %v", conv.LatestDM.FromMe, false)
	}
}

func TestDMMessage_Struct(t *testing.T) {
	msg := DMMessage{
		Content:   "Test message",
		FromMe:    true,
		Timestamp: nostr.Timestamp(1700000000),
	}

	if msg.Content != "Test message" {
		t.Errorf("msg.Content = %q, want %q", msg.Content, "Test message")
	}
	if msg.FromMe != true {
		t.Errorf("msg.FromMe = %v, want %v", msg.FromMe, true)
	}
	if msg.Timestamp != nostr.Timestamp(1700000000) {
		t.Errorf("msg.Timestamp = %v, want %v", msg.Timestamp, nostr.Timestamp(1700000000))
	}
}

func TestDMMessage_TimestampConversion(t *testing.T) {
	ts := nostr.Timestamp(1700000000)
	msg := DMMessage{
		Content:   "test",
		FromMe:    false,
		Timestamp: ts,
	}

	if !msg.Timestamp.Time().Equal(time.Unix(1700000000, 0)) {
		t.Errorf("msg.Timestamp.Time() = %v, want %v", msg.Timestamp.Time(), time.Unix(1700000000, 0))
	}
}

func TestGiftWrapTagParsing(t *testing.T) {
	ourPubKey := "aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111"
	theirPubKey := "bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222"

	tags := nostr.Tags{
		{"p", ourPubKey},
	}

	if len(tags) != 1 {
		t.Errorf("len(tags) = %d, want 1", len(tags))
	}
	if tags[0][0] != "p" || tags[0][1] != ourPubKey {
		t.Errorf("tags[0] = %v, want [p, %s]", tags[0], ourPubKey)
	}

	_ = theirPubKey
}

func TestDMFilterBuild(t *testing.T) {
	ourPubKey := nostr.PubKey{}
	theirPubKey := nostr.PubKey{}

	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindGiftWrap},
		Tags:  nostr.TagMap{"p": []string{ourPubKey.Hex(), theirPubKey.Hex()}},
		Limit: 50,
	}

	if len(filter.Kinds) != 1 {
		t.Errorf("len(filter.Kinds) = %d, want 1", len(filter.Kinds))
	}
	if filter.Kinds[0] != nostr.KindGiftWrap {
		t.Errorf("filter.Kinds[0] = %v, want %v", filter.Kinds[0], nostr.KindGiftWrap)
	}
	if len(filter.Tags["p"]) != 2 {
		t.Errorf("len(filter.Tags[p]) = %d, want 2", len(filter.Tags["p"]))
	}
	if filter.Limit != 50 {
		t.Errorf("filter.Limit = %d, want 50", filter.Limit)
	}
}

func TestDMKindConstants(t *testing.T) {
	if nostr.KindGiftWrap != 1059 {
		t.Errorf("nostr.KindGiftWrap = %d, want 1059", nostr.KindGiftWrap)
	}
}

func TestMessageFromMe(t *testing.T) {
	msgs := []DMMessage{
		{Content: "sent", FromMe: true, Timestamp: nostr.Timestamp(1000)},
		{Content: "received", FromMe: false, Timestamp: nostr.Timestamp(2000)},
	}

	if !msgs[0].FromMe {
		t.Errorf("msgs[0].FromMe = %v, want %v", msgs[0].FromMe, true)
	}
	if msgs[1].FromMe {
		t.Errorf("msgs[1].FromMe = %v, want %v", msgs[1].FromMe, false)
	}
}

func TestConversationSorting(t *testing.T) {
	convs := []Conversation{
		{LatestAt: nostr.Timestamp(1000)},
		{LatestAt: nostr.Timestamp(3000)},
		{LatestAt: nostr.Timestamp(2000)},
	}

	sorted := make([]Conversation, len(convs))
	copy(sorted, convs)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LatestAt > sorted[j].LatestAt
	})

	if sorted[0].LatestAt != nostr.Timestamp(3000) {
		t.Errorf("sorted[0].LatestAt = %v, want 3000", sorted[0].LatestAt)
	}
	if sorted[1].LatestAt != nostr.Timestamp(2000) {
		t.Errorf("sorted[1].LatestAt = %v, want 2000", sorted[1].LatestAt)
	}
	if sorted[2].LatestAt != nostr.Timestamp(1000) {
		t.Errorf("sorted[2].LatestAt = %v, want 1000", sorted[2].LatestAt)
	}
}

func TestMessageSorting(t *testing.T) {
	msgs := []DMMessage{
		{Timestamp: nostr.Timestamp(3000)},
		{Timestamp: nostr.Timestamp(1000)},
		{Timestamp: nostr.Timestamp(2000)},
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Timestamp < msgs[j].Timestamp
	})

	if msgs[0].Timestamp != nostr.Timestamp(1000) {
		t.Errorf("msgs[0].Timestamp = %v, want 1000", msgs[0].Timestamp)
	}
	if msgs[1].Timestamp != nostr.Timestamp(2000) {
		t.Errorf("msgs[1].Timestamp = %v, want 2000", msgs[1].Timestamp)
	}
	if msgs[2].Timestamp != nostr.Timestamp(3000) {
		t.Errorf("msgs[2].Timestamp = %v, want 3000", msgs[2].Timestamp)
	}
}