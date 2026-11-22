package sdk

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

func init() { registry.RequireCapability = requireCapability }

func requireCapability(ctx context.Context, name string) error {
	if _, err := currentSession(ctx); err != nil {
		return err
	}
	a := componentFor(ctx).active.Load()
	for _, capability := range a.ack.GetCapabilityIntersection() {
		if capability.GetName() == name && capability.GetVersion() == 1 {
			return nil
		}
	}
	return errcode.Unsupported.Errorf("protocol capability %s@1 is not negotiated", name)
}
