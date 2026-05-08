package config

type RelayFilter struct {
	Read  *bool `json:"read,omitempty"`
	Write *bool `json:"write,omitempty"`
}

func (f *RelayFilter) matchRelay(relay Relay) bool {
	relayRead := true
	if relay.Read != nil {
		relayRead = *relay.Read
	}

	relayWrite := false
	if relay.Write != nil {
		relayWrite = *relay.Write
	}

	if f.Read != nil && *f.Read != relayRead {
		return false
	}
	if f.Write != nil && *f.Write != relayWrite {
		return false
	}

	return true
}

func (f *RelayFilter) Matches(relayList []Relay) []string {
	res := make([]string, 0)
	for _, v := range relayList {
		if f.matchRelay(v) {
			res = append(res, v.URL)
		}
	}
	return res
}

func BoolPtr(b bool) *bool { return &b }

func GetWritableRelaysFromList(relayList []Relay) []string {
	return (&RelayFilter{Write: BoolPtr(true)}).Matches(relayList)
}

func GetReadableRelaysFromList(relayList []Relay) []string {
	return (&RelayFilter{Read: BoolPtr(true)}).Matches(relayList)
}

func GetRelayFromList(url string, relayList []Relay) (Relay, bool) {
	for _, r := range relayList {
		if r.URL == url {
			return r, true
		}
	}
	return Relay{}, false
}

func AddRelayToList(url string, read, write bool, relayList []Relay) []Relay {
	for i, r := range relayList {
		if r.URL == url {
			relayList[i].Read = &read
			relayList[i].Write = &write
			return relayList
		}
	}
	return append(relayList, Relay{URL: url, Read: &read, Write: &write})
}

func RemoveRelayFromList(url string, relayList []Relay) []Relay {
	newList := make([]Relay, 0)
	for _, r := range relayList {
		if r.URL != url {
			newList = append(newList, r)
		}
	}
	return newList
}

func AddDMRelayToList(url string, dmRelays []string) []string {
	for _, u := range dmRelays {
		if u == url {
			return dmRelays
		}
	}
	return append(dmRelays, url)
}

func RemoveDMRelayFromList(url string, dmRelays []string) []string {
	newList := make([]string, 0)
	for _, u := range dmRelays {
		if u != url {
			newList = append(newList, u)
		}
	}
	return newList
}

func AddSearchRelayToList(url string, searchRelays []string) []string {
	for _, u := range searchRelays {
		if u == url {
			return searchRelays
		}
	}
	return append(searchRelays, url)
}

func RemoveSearchRelayFromList(url string, searchRelays []string) []string {
	newList := make([]string, 0)
	for _, u := range searchRelays {
		if u != url {
			newList = append(newList, u)
		}
	}
	return newList
}