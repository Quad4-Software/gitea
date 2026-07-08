// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"gitea.dev/modules/setting"
)

var rnsMirrorURLPattern = regexp.MustCompile(`^rns://([0-9a-fA-F]{32})/([^/]+)/([^/]+)$`)

// MirrorURLParts holds the parsed components of an rns:// mirror source URL.
type MirrorURLParts struct {
	DestinationHash string
	Group           string
	Repository      string
}

// ParseMirrorURL validates and parses an rns:// repository URL.
func ParseMirrorURL(raw string) (*MirrorURLParts, error) {
	raw = strings.TrimSpace(raw)
	matches := rnsMirrorURLPattern.FindStringSubmatch(raw)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid rns mirror URL, expected rns://<hash>/<group>/<repo>")
	}
	group, err := url.PathUnescape(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid rns mirror group: %w", err)
	}
	repo, err := url.PathUnescape(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid rns mirror repository: %w", err)
	}
	return &MirrorURLParts{
		DestinationHash: strings.ToLower(matches[1]),
		Group:           strings.ToLower(group),
		Repository:      strings.ToLower(repo),
	}, nil
}

// ComposeLocalMirrorTarget returns the local rns:// URL for a repository on this node.
func ComposeLocalMirrorTarget(ownerName, repoName string) string {
	if !Enabled() {
		return ""
	}
	hash := strings.Trim(setting.Reticulum.DestinationHash, "<>")
	owner := strings.ToLower(ownerName)
	repo := strings.ToLower(repoName)
	return fmt.Sprintf("rns://%s/%s/%s", hash, url.PathEscape(owner), url.PathEscape(repo))
}
