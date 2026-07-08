// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"

	user_model "gitea.dev/models/user"
	reticulum_module "gitea.dev/modules/reticulum"
)

// GetUserIdentity returns the normalized Reticulum identity hash for a user.
func GetUserIdentity(ctx context.Context, userID int64) (string, error) {
	return userIdentity(ctx, userID)
}

// SetUserIdentity stores a Reticulum identity hash for a user.
func SetUserIdentity(ctx context.Context, userID int64, identity string) error {
	if identity == "" {
		return user_model.DeleteUserSetting(ctx, userID, reticulum_module.SettingKeyIdentity)
	}
	normalized, err := reticulum_module.NormalizeIdentityHash(identity)
	if err != nil {
		return err
	}
	return user_model.SetUserSetting(ctx, userID, reticulum_module.SettingKeyIdentity, normalized)
}
