package app

import "context"

func (a *assembly) commandMetadata(ctx context.Context, id string) (string, string, error) {
	a.core.developmentMu.RLock()
	for _, entry := range a.core.development {
		if entry.manifest.ID == id {
			name, description := entry.manifest.Name, entry.manifest.Description
			a.core.developmentMu.RUnlock()
			return name, description, nil
		}
	}
	a.core.developmentMu.RUnlock()
	in, err := a.store.GetInstall(ctx, id)
	if err != nil {
		return "", "", err
	}
	dir, err := installedDirectory(a.dir, in)
	if err != nil {
		return "", "", err
	}
	parsed, err := installedManifest(in, dir)
	if err != nil {
		return "", "", err
	}
	return parsed.Manifest.Name, parsed.Manifest.Description, nil
}
