// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	repo_model "gitea.dev/models/repo"
	"gitea.dev/modules/setting"
	"gitea.dev/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadThanksCountFromFile(t *testing.T) {
	defer test.MockVariableValue(&setting.Reticulum.Enabled, true)()

	dir := t.TempDir()
	setting.RepoRootPath = dir

	ownerDir := filepath.Join(dir, "owner")
	require.NoError(t, os.MkdirAll(ownerDir, 0o755))
	repoPath := filepath.Join(ownerDir, "repo.git")
	require.NoError(t, os.MkdirAll(repoPath, 0o755))

	repo := &repo_model.Repository{
		OwnerName: "owner",
		Name:      "repo",
	}

	count, err := ReadThanksCountFromFile(repo)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	thanksPath := strings.TrimSuffix(repoPath, ".git") + ".thanks"
	require.NoError(t, os.WriteFile(thanksPath, []byte{0x81, 0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x05}, 0o600))

	count, err = ReadThanksCountFromFile(repo)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}
