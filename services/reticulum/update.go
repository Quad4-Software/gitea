// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"gitea.dev/modules/setting"
)

// PackageInfo describes the installed RNS Python package.
type PackageInfo struct {
	RngitVersion string
	PipVersion   string
}

// GetPackageInfo returns version information for rngit and the RNS pip package.
func GetPackageInfo(ctx context.Context) *PackageInfo {
	info := &PackageInfo{}
	if path, err := exec.LookPath(setting.Reticulum.RngitPath); err == nil {
		if out, err := exec.CommandContext(ctx, path, "--version").CombinedOutput(); err == nil {
			info.RngitVersion = strings.TrimSpace(string(out))
		}
	}
	if out, err := exec.CommandContext(ctx, "pip3", "show", "rns").CombinedOutput(); err == nil {
		for line := range strings.SplitSeq(string(out), "\n") {
			if version, ok := strings.CutPrefix(line, "Version: "); ok {
				info.PipVersion = strings.TrimSpace(version)
				break
			}
		}
	}
	return info
}

// UpdateRNSPackage upgrades the RNS Python package and restarts rngit.
func UpdateRNSPackage(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "pip3", "install", "--upgrade", "--break-system-packages", "rns")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return output, fmt.Errorf("pip install --upgrade rns: %w", err)
	}
	if setting.Reticulum.StartBuiltinServer {
		if err := RestartBuiltinServer(ctx); err != nil {
			return output, err
		}
	}
	return output, nil
}
