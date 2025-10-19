package nostr

import (
	"testing"

	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/pkg/config"
)

// TestClient_Subscribe 测试订阅功能
func TestClient_Subscribe(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{}
	cfg.Client.DefaultRelays = []string{"wss://bostr.shop"}

	// 创建客户端
	client, err := NewClientWithoutI2P(cfg)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 测试订阅基本功能
	t.Run("订阅基本功能", func(t *testing.T) {
		// 创建测试过滤器
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote},
			Limit: 10,
		}

		// 由于没有实际连接到 relay，这里主要测试函数调用不会panic
		// 在实际环境中，应该使用 mock 或测试 relay
		subscription, err := client.Subscribe(filter)

		// 由于没有实际连接，应该返回 nil
		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}
	})

	// 测试空过滤器
	t.Run("空过滤器", func(t *testing.T) {
		filter := nostr.Filter{}
		subscription, err := client.Subscribe(filter)

		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}
	})

	// 测试特定类型的过滤器
	t.Run("特定事件类型过滤器", func(t *testing.T) {
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote, nostr.KindRecommendServer},
			Limit: 5,
		}

		subscription, err := client.Subscribe(filter)

		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}
	})
}

// TestClient_SubscribeMultipleFilters 测试多个过滤器的订阅
func TestClient_SubscribeMultipleFilters(t *testing.T) {
	cfg := &config.Config{}
	cfg.Client.DefaultRelays = []string{"ws://localhost:8080"}

	client, err := NewClientWithoutI2P(cfg)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	tests := []struct {
		name   string
		filter nostr.Filter
	}{
		{
			name: "文本笔记过滤器",
			filter: nostr.Filter{
				Kinds: []nostr.Kind{nostr.KindTextNote},
				Limit: 10,
			},
		},
		{
			name: "推荐服务器过滤器",
			filter: nostr.Filter{
				Kinds: []nostr.Kind{nostr.KindRecommendServer},
				Limit: 5,
			},
		},
		{
			name: "作者过滤器",
			filter: nostr.Filter{
				Authors: []nostr.PubKey{},
				Limit:   3,
			},
		},
		{
			name: "标签过滤器",
			filter: nostr.Filter{
				Tags: nostr.TagMap{
					"t": []string{"test-tag"},
				},
				Limit: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription, err := client.Subscribe(tt.filter)

			// 验证函数调用不会panic，并且返回预期的结果
			if subscription != nil {
				t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
			}
			if err != nil {
				t.Errorf("期望 err 为 nil, 但得到 %v", err)
			}
		})
	}
}

// TestClient_SubscribeEdgeCases 测试边界情况
func TestClient_SubscribeEdgeCases(t *testing.T) {
	cfg := &config.Config{}
	cfg.Client.DefaultRelays = []string{"ws://localhost:8080"}

	client, err := NewClientWithoutI2P(cfg)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	t.Run("大限制值", func(t *testing.T) {
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote},
			Limit: 1000, // 大限制值
		}

		subscription, err := client.Subscribe(filter)

		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}
	})

	t.Run("零限制值", func(t *testing.T) {
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote},
			Limit: 0, // 零限制值
		}

		subscription, err := client.Subscribe(filter)

		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}
	})

	t.Run("负限制值", func(t *testing.T) {
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote},
			Limit: -1, // 负限制值
		}

		subscription, err := client.Subscribe(filter)

		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}
	})
}

// TestClient_Integration 测试客户端集成功能
func TestClient_Integration(t *testing.T) {
	cfg := &config.Config{}
	cfg.Client.DefaultRelays = []string{"ws://localhost:8080"}

	client, err := NewClientWithoutI2P(cfg)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	t.Run("订阅后获取relays", func(t *testing.T) {
		// 先获取relays列表
		relays := client.GetRelays()
		if len(relays) != 0 {
			t.Errorf("未连接的客户端应该没有relays, 但得到 %v", relays)
		}

		// 尝试订阅
		filter := nostr.Filter{
			Kinds: []nostr.Kind{nostr.KindTextNote},
		}
		subscription, err := client.Subscribe(filter)

		if subscription != nil {
			t.Errorf("期望 subscription 为 nil, 但得到 %v", subscription)
		}
		if err != nil {
			t.Errorf("期望 err 为 nil, 但得到 %v", err)
		}

		// 再次获取relays列表
		relaysAfter := client.GetRelays()
		if len(relaysAfter) != 0 {
			t.Errorf("订阅后仍然没有连接的relays, 但得到 %v", relaysAfter)
		}
	})
}
