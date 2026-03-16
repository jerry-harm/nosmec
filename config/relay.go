package config

import (
	"fmt"

	"fiatjaf.com/nostr"
	"github.com/spf13/viper"
)

// RelayFilter 用于过滤具有特定属性的 relay
type RelayFilter struct {
	Read   *bool `json:"read,omitempty"`
	Write  *bool `json:"write,omitempty"`
	Search *bool `json:"search,omitempty"`
	DM     *bool `json:"dm,omitempty"`
}

// Matches 检查 relay 是否匹配过滤条件
func (f *RelayFilter) matche(relay Relay) bool {
	// 获取 relay 的实际值，处理 nil 指针
	relayRead := true // 默认只可读
	if relay.Read != nil {
		relayRead = *relay.Read
	}

	relayWrite := false // 默认不可写
	if relay.Write != nil {
		relayWrite = *relay.Write
	}

	relaySearch := false // 默认不可搜索
	if relay.Search != nil {
		relaySearch = *relay.Search
	}

	relayDM := false // 默认不可 DM
	if relay.DM != nil {
		relayDM = *relay.DM
	}

	// 检查过滤条件
	// 如果过滤条件不为 nil，则必须匹配
	if f.Read != nil && *f.Read != relayRead {
		return false
	}
	if f.Write != nil && *f.Write != relayWrite {
		return false
	}
	if f.Search != nil && *f.Search != relaySearch {
		return false
	}
	if f.DM != nil && *f.DM != relayDM {
		return false
	}

	return true
}

func (f *RelayFilter) Matches() []string {
	res := make([]string, 0)
	for k, v := range globalConfig.RelayList {
		if f.matche(v) {
			res = append(res, k)
		}
	}
	return res
}

func SaveRelays() {
	relays := make([]string, 0)
	globalPool.Relays.Range(func(key string, value *nostr.Relay) bool {
		relays = append(relays, key)
		return true
	})
	for k, _ := range globalConfig.RelayList {
		relays = append(relays, k)
	}
	temp := append(globalConfig.KnownRelays, relays...)
	set := make(map[string]struct{})
	res := make([]string, 0)
	for _, v := range temp {
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			res = append(res, v)
		}
	}
	globalConfig.KnownRelays = res
	// 保存配置
	if err := viper.WriteConfig(); err != nil {
		fmt.Printf("Warning: Failed to save config: %v\n", err)
	}
}
