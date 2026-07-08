// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"

	"gitea.dev/models/perm"
	access_model "gitea.dev/models/perm/access"
	repo_model "gitea.dev/models/repo"
	"gitea.dev/models/unit"
	user_model "gitea.dev/models/user"
	reticulum_module "gitea.dev/modules/reticulum"
	"gitea.dev/modules/setting"
)

// CanShowCloneURL reports whether the clone panel should show an RNS URL.
func CanShowCloneURL(ctx context.Context, doer *user_model.User, repo *repo_model.Repository) bool {
	if !reticulum_module.Enabled() {
		return false
	}
	if doer == nil {
		return !repo.IsPrivate && setting.Reticulum.PublicRead
	}
	has, err := access_model.HasAccessUnit(ctx, doer, repo, unit.TypeCode, perm.AccessModeRead)
	return err == nil && has
}
