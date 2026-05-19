package nostr_sdk

import (
	"fmt"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip10"
	"github.com/jerry-harm/nosmec/nip72"
)

func GetThreadParentPointer(event *nostr.Event) nostr.Pointer {
	if event == nil {
		return nil
	}

	if ptr := nip72.GetRootPointer(event); ptr != nil {
		return nil
	}

	if ptr := nip72.GetParentPointer(event); ptr != nil {
		return ptr
	}

	return nip10.GetImmediateParent(event.Tags)
}

func GetThreadRootID(event *nostr.Event) (rootID nostr.ID, isRoot bool, err error) {
	if event == nil {
		return nostr.ID{}, false, fmt.Errorf("nil event")
	}

	if ptr := nip72.GetRootPointer(event); ptr != nil {
		return event.ID, true, nil
	}
	if ptr := nip72.GetParentPointer(event); ptr != nil {
		if ep, ok := ptr.(nostr.EventPointer); ok {
			return ep.ID, ep.ID == event.ID, nil
		}
	}

	ptr := nip10.GetThreadRoot(event.Tags)
	if ptr == nil {
		return event.ID, true, nil
	}

	id, err := nostr.IDFromHex(ptr.AsTagReference())
	if err != nil {
		return nostr.ID{}, false, err
	}

	return id, id == event.ID, nil
}
