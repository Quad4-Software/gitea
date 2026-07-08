// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"testing"

	"gitea.dev/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeCloneURL(t *testing.T) {
	oldEnabled := setting.Reticulum.Enabled
	oldHash := setting.Reticulum.DestinationHash
	t.Cleanup(func() {
		setting.Reticulum.Enabled = oldEnabled
		setting.Reticulum.DestinationHash = oldHash
	})

	setting.Reticulum.Enabled = false
	assert.Empty(t, ComposeCloneURL("Owner", "Repo"))

	setting.Reticulum.Enabled = true
	setting.Reticulum.DestinationHash = ""
	assert.Empty(t, ComposeCloneURL("Owner", "Repo"))

	setting.Reticulum.DestinationHash = "f46dc814df3599844ad2d10b26c0a1a4"
	assert.Equal(t, "rns://f46dc814df3599844ad2d10b26c0a1a4/owner/repo", ComposeCloneURL("Owner", "Repo"))
}

func TestNormalizeIdentityHash(t *testing.T) {
	hash, err := NormalizeIdentityHash("<F46DC814DF3599844AD2D10B26C0A1A4>")
	require.NoError(t, err)
	assert.Equal(t, "f46dc814df3599844ad2d10b26c0a1a4", hash)

	_, err = NormalizeIdentityHash("short")
	assert.Error(t, err)

	_, err = NormalizeIdentityHash("gggggggggggggggggggggggggggggggg")
	assert.Error(t, err)
}
