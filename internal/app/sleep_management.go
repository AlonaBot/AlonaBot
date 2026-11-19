package app

import (
	"context"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/core/idle"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

func (c *Core) managementSleepPolicy(ctx context.Context, in state.Install, development *manifest.Manifest) (manage.SleepPolicy, error) {
	m := development
	if m == nil {
		directory, err := installedDirectory(c.dir, in)
		if err != nil {
			return manage.SleepPolicy{}, err
		}
		parsed, err := installedManifest(in, directory)
		if err != nil {
			return manage.SleepPolicy{}, err
		}
		m = parsed.Manifest
	}
	sleep := c.assembly.settings.Settings().Sleep
	// 四类常驻原因先于路由；其余结论必须使用有效的已提交快照。
	policy := idle.Resolve(sleep, m, development == nil && hasInstalledRoutes(in), nil)
	if policy.Reason == "" {
		var err error
		policy, _, err = c.sleepPolicy(ctx, in, sleep)
		if err != nil {
			return manage.SleepPolicy{}, err
		}
	}
	return manage.SleepPolicy{Policy: policy.Policy, EffectivePolicy: policy.EffectivePolicy,
		SleepCapable: policy.SleepCapable, Eligible: policy.Eligible, Reason: policy.Reason, Warnings: policy.Warnings}, nil
}
