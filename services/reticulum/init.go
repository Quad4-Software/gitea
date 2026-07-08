// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"fmt"

	"gitea.dev/modules/graceful"
	"gitea.dev/modules/log"
	reticulum_module "gitea.dev/modules/reticulum"
	"gitea.dev/modules/setting"
)

// Init prepares rngit integration and optionally starts the builtin server.
func Init(ctx context.Context) error {
	if !setting.Reticulum.Enabled {
		return nil
	}

	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("prepare reticulum config directory: %w", err)
	}

	if err := EnsureNetworkConfig(); err != nil {
		return fmt.Errorf("prepare reticulum network config: %w", err)
	}

	if err := WriteConfig(ctx); err != nil {
		return err
	}

	if setting.Reticulum.DestinationHash == "" {
		info, err := ResolveIdentities(ctx)
		if err != nil {
			log.Warn("Reticulum: could not resolve destination hashes (set DESTINATION_HASH in app.ini): %v", err)
		} else {
			setting.Reticulum.DestinationHash = info.RepositoriesDestinationHash
			setting.Reticulum.NomadDestinationHash = info.NomadDestinationHash
			log.Info("Reticulum: repositories destination is <%s>", info.RepositoriesDestinationHash)
			if info.NomadDestinationHash != "" {
				log.Info("Reticulum: nomad network destination is <%s>", info.NomadDestinationHash)
			}
		}
	}

	if err := SyncAllRepositoryPermissions(ctx); err != nil {
		return fmt.Errorf("sync reticulum permissions: %w", err)
	}

	if setting.Reticulum.SyncReticulumThanks {
		if updated, err := SyncAllRepositoryThanks(ctx); err != nil {
			log.Warn("Reticulum: initial thanks sync: %v", err)
		} else if updated > 0 {
			log.Info("Reticulum: synchronized thanks counts for %d repositories", updated)
		}
	}

	if setting.Reticulum.StartBuiltinServer {
		registerProcessContext(ctx)
		initBuiltinShutdown()
		if err := startBuiltinServer(); err != nil {
			return err
		}
	}

	log.Info("Reticulum: rngit integration enabled")
	return nil
}

// Enabled reports whether Reticulum git support is active.
func Enabled() bool {
	return reticulum_module.Enabled()
}

// Shutdown stops the builtin rngit server if it is running.
func Shutdown() {
	stopBuiltinServer()
}

// ShutdownContext returns the graceful shutdown context for background reticulum work.
func ShutdownContext() context.Context {
	return graceful.GetManager().ShutdownContext()
}
