// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cron

import (
	"context"

	user_model "gitea.dev/models/user"
	"gitea.dev/modules/setting"
	reticulum_service "gitea.dev/services/reticulum"
)

func registerSyncReticulumThanks() {
	RegisterTaskFatal("sync_reticulum_thanks", &BaseConfig{
		Enabled:    true,
		RunAtStart: false,
		Schedule:   "@every 5m",
	}, func(ctx context.Context, _ *user_model.User, _ Config) error {
		if !setting.Reticulum.Enabled || !setting.Reticulum.SyncReticulumThanks {
			return nil
		}
		_, err := reticulum_service.SyncAllRepositoryThanks(ctx)
		return err
	})
}

func initReticulumTasks() {
	registerSyncReticulumThanks()
}
