package nip72

import (
	"strconv"
	"strings"

	"fiatjaf.com/nostr"
)

type Role int

const (
	Unknown Role = iota
	TopLevelPost
	Reply
)

func GetCommunityPointer(event *nostr.Event) nostr.Pointer {
	if event == nil || event.Kind != nostr.KindComment {
		return nil
	}
	ptr, ok := parseCommunityPointer(event.Tags.Find("A"))
	if !ok || !hasRootScope(event.Tags, ptr) {
		return nil
	}
	return ptr
}

func GetRootPointer(event *nostr.Event) nostr.Pointer {
	if event == nil || event.Kind != nostr.KindComment {
		return nil
	}
	ptr, ok := parseCommunityPointer(event.Tags.Find("A"))
	if !ok || !hasTopLevelMarkers(event.Tags, ptr) {
		return nil
	}
	return ptr
}

func GetParentPointer(event *nostr.Event) nostr.Pointer {
	if event == nil || event.Kind != nostr.KindComment {
		return nil
	}
	ptr, ok := parseCommunityPointer(event.Tags.Find("A"))
	if !ok || hasTopLevelMarkers(event.Tags, ptr) {
		return nil
	}

	eTag := event.Tags.Find("e")
	if len(eTag) < 2 || !nostr.IsValid32ByteHex(eTag[1]) {
		return nil
	}
	if event.Tags.Find("p") == nil || event.Tags.Find("k") == nil {
		return nil
	}

	parent, err := nostr.EventPointerFromTag(eTag)
	if err != nil {
		return nil
	}
	parent.Kind = parseKindTag(event.Tags.Find("k"))
	if parent.Kind == 0 {
		return nil
	}
	parent.Author, _ = parsePubKeyTag(event.Tags.Find("p"))
	return parent
}

func ClassifyRole(event *nostr.Event) (Role, bool) {
	if GetRootPointer(event) != nil {
		return TopLevelPost, true
	}
	if GetParentPointer(event) != nil {
		return Reply, true
	}
	return Unknown, false
}

func parseCommunityPointer(tag nostr.Tag) (nostr.Pointer, bool) {
	if len(tag) < 2 {
		return nil, false
	}
	ptr, err := nostr.EntityPointerFromTag(tag)
	if err != nil {
		return nil, false
	}
	if !strictCommunityAddr(tag[1]) {
		return nil, false
	}
	return ptr, true
}

func hasRootScope(tags nostr.Tags, ptr nostr.Pointer) bool {
	if ptr == nil || tags.Find("P") == nil || tags.Find("K") == nil {
		return false
	}
	community, ok := ptr.(nostr.EntityPointer)
	if !ok {
		return false
	}
	rootPK, ok := parsePubKeyTag(tags.Find("P"))
	if !ok || rootPK != community.PublicKey {
		return false
	}
	return parseKindTag(tags.Find("K")) == nostr.KindCommunityDefinition
}

func hasTopLevelMarkers(tags nostr.Tags, ptr nostr.Pointer) bool {
	if tags.Find("a") == nil || tags.Find("p") == nil || tags.Find("k") == nil {
		return false
	}
	if !hasRootScope(tags, ptr) {
		return false
	}
	community := ptr.(nostr.EntityPointer)
	parentPtr, err := nostr.EntityPointerFromTag(tags.Find("a"))
	if err != nil || parentPtr.PublicKey != community.PublicKey || parentPtr.Kind != community.Kind || parentPtr.Identifier != community.Identifier {
		return false
	}
	return parseKindTag(tags.Find("k")) == nostr.KindCommunityDefinition && samePubKeyTag(tags.Find("P"), tags.Find("p"))
}

func strictCommunityAddr(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 3 || parts[0] != "34550" {
		return false
	}
	if !nostr.IsValid32ByteHex(parts[1]) || parts[2] == "" {
		return false
	}
	return true
}

func parseKindTag(tag nostr.Tag) nostr.Kind {
	if len(tag) < 2 || tag[1] == "" {
		return 0
	}
	kind, err := strconv.Atoi(tag[1])
	if err != nil {
		return 0
	}
	return nostr.Kind(kind)
}

func parsePubKeyTag(tag nostr.Tag) (nostr.PubKey, bool) {
	if len(tag) < 2 || !nostr.IsValid32ByteHex(tag[1]) {
		return nostr.PubKey{}, false
	}
	pk, err := nostr.PubKeyFromHexCheap(tag[1])
	if err != nil {
		return nostr.PubKey{}, false
	}
	return pk, true
}

func samePubKeyTag(a, b nostr.Tag) bool {
	pa, oka := parsePubKeyTag(a)
	pb, okb := parsePubKeyTag(b)
	return oka && okb && pa == pb
}
