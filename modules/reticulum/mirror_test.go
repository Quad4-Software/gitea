// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMirrorURL(t *testing.T) {
	parts, err := ParseMirrorURL("rns://06a54b505bb67b25ef3f8097e8001edc/public/MeshChatX")
	require.NoError(t, err)
	assert.Equal(t, "06a54b505bb67b25ef3f8097e8001edc", parts.DestinationHash)
	assert.Equal(t, "public", parts.Group)
	assert.Equal(t, "meshchatx", parts.Repository)

	_, err = ParseMirrorURL("https://example.com/a/b")
	assert.Error(t, err)
}
