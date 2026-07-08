// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"strings"
	"testing"
)

func TestParseIdentitiesOutput(t *testing.T) {
	stdout := strings.Join([]string{
		"Git Peer Identity         : <aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa>",
		"Repository Node Identity  : <bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb>",
		"Repositories Destination  : <cccccccccccccccccccccccccccccccc>",
		"Nomad Network Destination : <dddddddddddddddddddddddddddddddd>",
	}, "\n")

	info := &IdentityInfo{}
	for _, line := range strings.Split(stdout, "\n") {
		if matches := repositoriesDestinationPattern.FindStringSubmatch(line); len(matches) == 2 {
			info.RepositoriesDestinationHash = matches[1]
		}
		if matches := nomadDestinationPattern.FindStringSubmatch(line); len(matches) == 2 {
			info.NomadDestinationHash = matches[1]
		}
	}

	if info.RepositoriesDestinationHash != "cccccccccccccccccccccccccccccccc" {
		t.Fatalf("unexpected repositories destination: %q", info.RepositoriesDestinationHash)
	}
	if info.NomadDestinationHash != "dddddddddddddddddddddddddddddddd" {
		t.Fatalf("unexpected nomad destination: %q", info.NomadDestinationHash)
	}
}
