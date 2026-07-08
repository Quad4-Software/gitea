// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"fmt"
	"net/url"
	"strings"

	"gitea.dev/modules/setting"
)

// SettingKeyIdentity is the user_setting key for a Reticulum identity hash.
const SettingKeyIdentity = "reticulum.identity"

// Enabled reports whether Reticulum git clone URLs should be shown.
func Enabled() bool {
	return setting.Reticulum.Enabled && setting.Reticulum.DestinationHash != ""
}

// ComposeCloneURL returns an rns:// clone URL for the given owner and repository.
func ComposeCloneURL(ownerName, repoName string) string {
	if !Enabled() {
		return ""
	}
	hash := strings.Trim(setting.Reticulum.DestinationHash, "<>")
	owner := strings.ToLower(ownerName)
	repo := strings.ToLower(repoName)
	return fmt.Sprintf("rns://%s/%s/%s", hash, url.PathEscape(owner), url.PathEscape(repo))
}

// NormalizeIdentityHash strips angle brackets and validates a 32-char hex identity hash.
func NormalizeIdentityHash(hash string) (string, error) {
	hash = strings.Trim(strings.ToLower(strings.TrimSpace(hash)), "<>")
	if len(hash) != 32 {
		return "", fmt.Errorf("reticulum identity hash must be 32 hexadecimal characters")
	}
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", fmt.Errorf("reticulum identity hash must be hexadecimal")
		}
	}
	return hash, nil
}
