package utils

import (
	"errors"
	"regexp"

	"fiatjaf.com/nostr"
)

var ErrInvalidNoteID = errors.New("invalid note ID: must be 64 hex characters")

var noteIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

func BuildNoteFilter(noteID string) (nostr.Filter, error) {
	if noteID == "" || !noteIDRegex.MatchString(noteID) {
		return nostr.Filter{}, ErrInvalidNoteID
	}
	id, err := nostr.IDFromHex(noteID)
	if err != nil {
		return nostr.Filter{}, ErrInvalidNoteID
	}
	return nostr.Filter{
		IDs:   []nostr.ID{id},
		Limit: 1,
	}, nil
}

func BuildProfileFilter(pubKey nostr.PubKey) nostr.Filter {
	return nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindProfileMetadata},
		Authors: []nostr.PubKey{pubKey},
		Limit:   1,
	}
}

func BuildProfilesFilter(pubKeys []nostr.PubKey) nostr.Filter {
	return nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindProfileMetadata},
		Authors: pubKeys,
		Limit:   1,
	}
}

func BuildTimelineFilter(pubKey nostr.PubKey, limit int, until nostr.Timestamp) nostr.Filter {
	filter := nostr.Filter{
		Kinds:   []nostr.Kind{nostr.KindTextNote},
		Authors: []nostr.PubKey{pubKey},
		Limit:   limit,
	}
	if until > 0 {
		filter.Until = until
	}
	return filter
}

func BuildGlobalTimelineFilter(limit int, until nostr.Timestamp) nostr.Filter {
	filter := nostr.Filter{
		Kinds: []nostr.Kind{nostr.KindTextNote},
		Limit: limit,
	}
	if until > 0 {
		filter.Until = until
	}
	return filter
}

func BuildFollowedTimelineFilter(authors []nostr.PubKey, communityAddrs []string, hashtags []string, limit int, until nostr.Timestamp) nostr.Filter {
	kinds := []nostr.Kind{nostr.KindTextNote, nostr.KindComment}
	filter := nostr.Filter{
		Kinds: kinds,
		Limit: limit * 3,
	}
	if until > 0 {
		filter.Until = until
	}
	if len(authors) > 0 {
		filter.Authors = authors
	}
	if len(communityAddrs) > 0 {
		filter.Tags = nostr.TagMap{"a": communityAddrs}
	}
	return filter
}

func BuildParentEventFilter(noteID string) (nostr.Filter, error) {
	if noteID == "" {
		return nostr.Filter{}, ErrInvalidNoteID
	}
	id, err := nostr.IDFromHex(noteID)
	if err != nil {
		return nostr.Filter{}, ErrInvalidNoteID
	}
	return nostr.Filter{
		IDs:   []nostr.ID{id},
		Limit: 1,
	}, nil
}
