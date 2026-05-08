package utils

import (
	"context"
	"fmt"

	"github.com/jerry-harm/nosmec/config"
)

func SyncAll(ctx context.Context, app *config.AppContext) error {
	if err := SyncProfile(ctx, app); err != nil {
		return fmt.Errorf("failed to sync profile: %w", err)
	}

	if err := SyncSubscriptionsFromNetwork(ctx, app); err != nil {
		return fmt.Errorf("failed to sync subscriptions: %w", err)
	}

	if err := SyncRelaysFromNetwork(ctx, app); err != nil {
		return fmt.Errorf("failed to sync relays: %w", err)
	}

	return nil
}

func PublishAll(ctx context.Context, app *config.AppContext) error {
	if _, err := SetProfile(ctx, app, true, "", "", "", "", "", "", "", "", "", "", ""); err != nil {
		return fmt.Errorf("failed to publish profile: %w", err)
	}

	if err := PublishSubscriptions(ctx, app); err != nil {
		return fmt.Errorf("failed to publish subscriptions: %w", err)
	}

	if err := PublishRelayList(ctx, app); err != nil {
		return fmt.Errorf("failed to publish relay list: %w", err)
	}

	return nil
}
