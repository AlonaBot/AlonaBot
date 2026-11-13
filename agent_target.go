package alonabot

import (
	"context"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

func (a *assembly) agentID(ctx context.Context) (string, error) {
	installed, err := a.store.Installs(ctx)
	if err != nil {
		return "", err
	}
	for _, component := range installed {
		if component.Role == state.RoleAgent && !component.Uninstalled {
			return component.PluginID, nil
		}
	}
	return "", nil
}
