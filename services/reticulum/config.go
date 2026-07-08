// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitea.dev/modules/log"
	"gitea.dev/modules/setting"
)

var (
	repositoriesDestinationPattern = regexp.MustCompile(`Repositories Destination\s*:\s*<([0-9a-fA-F]{32})>`)
	nomadDestinationPattern        = regexp.MustCompile(`Nomad Network Destination\s*:\s*<([0-9a-fA-F]{32})>`)
)

// IdentityInfo holds identity and destination hashes reported by rngit --print-identity.
type IdentityInfo struct {
	RepositoriesDestinationHash string
	NomadDestinationHash        string
}

// WriteConfig generates the rngit configuration file from Gitea's repository layout.
func WriteConfig(ctx context.Context) error {
	if !setting.Reticulum.Enabled {
		return nil
	}

	configDir := setting.Reticulum.ConfigPath
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("create reticulum config directory: %w", err)
	}

	owners, err := listOwnerDirectories()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString("[rngit]\n")
	if setting.Reticulum.NodeName != "" {
		_, _ = fmt.Fprintf(&buf, "node_name = %s\n", setting.Reticulum.NodeName)
	}
	buf.WriteString("announce_interval = 360\n\n")

	buf.WriteString("[repositories]\n")
	for _, owner := range owners {
		ownerPath := filepath.Join(setting.RepoRootPath, owner)
		_, _ = fmt.Fprintf(&buf, "%s = %s\n", owner, ownerPath)
	}
	buf.WriteString("\n")

	buf.WriteString("[pages]\n")
	if setting.Reticulum.ServeNomadNet {
		buf.WriteString("serve_nomadnet = yes\n")
	} else {
		buf.WriteString("serve_nomadnet = no\n")
	}
	buf.WriteString("\n")

	buf.WriteString("[logging]\n")
	buf.WriteString("loglevel = 4\n")

	configPath := filepath.Join(configDir, "config")
	if err := os.WriteFile(configPath, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write rngit config: %w", err)
	}

	log.Info("Reticulum: wrote rngit config to %s with %d owner groups", configPath, len(owners))
	return nil
}

func listOwnerDirectories() ([]string, error) {
	entries, err := os.ReadDir(setting.RepoRootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read repository root: %w", err)
	}

	owners := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			owners = append(owners, entry.Name())
		}
	}
	return owners, nil
}

// ResolveIdentities runs rngit --print-identity and parses repository and Nomad Network destinations.
func ResolveIdentities(ctx context.Context) (*IdentityInfo, error) {
	if err := WriteConfig(ctx); err != nil {
		return nil, err
	}

	args := []string{"--print-identity", "--config", setting.Reticulum.ConfigPath}
	if setting.Reticulum.RNSConfigPath != "" {
		args = append(args, "--rnsconfig", setting.Reticulum.RNSConfigPath)
	}

	stdout, err := runRngit(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("rngit --print-identity: %w", err)
	}

	info := &IdentityInfo{}
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := repositoriesDestinationPattern.FindStringSubmatch(line); len(matches) == 2 {
			info.RepositoriesDestinationHash = matches[1]
		}
		if matches := nomadDestinationPattern.FindStringSubmatch(line); len(matches) == 2 {
			info.NomadDestinationHash = matches[1]
		}
	}
	if info.RepositoriesDestinationHash == "" {
		return nil, errors.New("could not parse repositories destination from rngit output")
	}
	return info, nil
}

// ResolveDestinationHash runs rngit --print-identity and parses the repositories destination hash.
func ResolveDestinationHash(ctx context.Context) (string, error) {
	info, err := ResolveIdentities(ctx)
	if err != nil {
		return "", err
	}
	return info.RepositoriesDestinationHash, nil
}
