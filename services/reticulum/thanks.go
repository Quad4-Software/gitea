// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gitea.dev/models/db"
	repo_model "gitea.dev/models/repo"
	"gitea.dev/modules/log"
	reticulum_module "gitea.dev/modules/reticulum"
	"gitea.dev/modules/setting"
)

// ShowReticulumThanks reports whether rngit Thanks counts should be shown in the UI.
func ShowReticulumThanks() bool {
	return reticulum_module.Enabled() && setting.Reticulum.ShowReticulumThanks
}

func thanksFilePath(repo *repo_model.Repository) string {
	base := strings.TrimSuffix(repo.RepoPath(), ".git")
	return base + ".thanks"
}

// ReadThanksCountFromFile reads the rngit Thanks counter for a repository.
func ReadThanksCountFromFile(repo *repo_model.Repository) (int, error) {
	data, err := os.ReadFile(thanksFilePath(repo))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read reticulum thanks for %s: %w", repo.FullName(), err)
	}
	count, err := reticulum_module.ParseThanksCount(data)
	if err != nil {
		return 0, fmt.Errorf("parse reticulum thanks for %s: %w", repo.FullName(), err)
	}
	if count < 0 {
		return 0, nil
	}
	return count, nil
}

// RepositoryThanksCount returns the Thanks count for display.
func RepositoryThanksCount(repo *repo_model.Repository) int {
	if !ShowReticulumThanks() {
		return 0
	}
	if setting.Reticulum.SyncReticulumThanks {
		return repo.NumReticulumThanks
	}
	count, err := ReadThanksCountFromFile(repo)
	if err != nil {
		log.Warn("Reticulum: %v", err)
		return 0
	}
	return count
}

// SyncRepositoryThanks updates the cached Thanks count from the rngit .thanks file.
func SyncRepositoryThanks(ctx context.Context, repo *repo_model.Repository) (bool, error) {
	if !reticulum_module.Enabled() || !setting.Reticulum.SyncReticulumThanks {
		return false, nil
	}

	count, err := ReadThanksCountFromFile(repo)
	if err != nil {
		return false, err
	}
	if count == repo.NumReticulumThanks {
		return false, nil
	}

	_, err = db.GetEngine(ctx).ID(repo.ID).Cols("num_reticulum_thanks").Update(&repo_model.Repository{
		NumReticulumThanks: count,
	})
	if err != nil {
		return false, fmt.Errorf("update reticulum thanks for %s: %w", repo.FullName(), err)
	}
	repo.NumReticulumThanks = count
	return true, nil
}

// SyncAllRepositoryThanks refreshes cached Thanks counts for every repository.
func SyncAllRepositoryThanks(ctx context.Context) (updated int, err error) {
	if !reticulum_module.Enabled() || !setting.Reticulum.SyncReticulumThanks {
		return 0, nil
	}

	var repos []*repo_model.Repository
	if err := db.GetEngine(ctx).Cols("id", "owner_name", "name", "num_reticulum_thanks").Find(&repos); err != nil {
		return 0, fmt.Errorf("list repositories for reticulum thanks sync: %w", err)
	}

	for _, repo := range repos {
		changed, err := SyncRepositoryThanks(ctx, repo)
		if err != nil {
			log.Warn("Reticulum: sync thanks for %s: %v", repo.FullName(), err)
			continue
		}
		if changed {
			updated++
		}
	}
	return updated, nil
}

// RefreshRepositoryThanks syncs a single repository and returns the current count.
func RefreshRepositoryThanks(ctx context.Context, repo *repo_model.Repository) int {
	if !ShowReticulumThanks() {
		return 0
	}
	if setting.Reticulum.SyncReticulumThanks {
		if _, err := SyncRepositoryThanks(ctx, repo); err != nil {
			log.Warn("Reticulum: sync thanks for %s: %v", repo.FullName(), err)
		}
		return repo.NumReticulumThanks
	}
	return RepositoryThanksCount(repo)
}
