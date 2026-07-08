// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	reticulum_module "gitea.dev/modules/reticulum"
	"gitea.dev/modules/setting"
)

// Status describes the current Reticulum / rngit integration state.
type Status struct {
	Enabled              bool
	CloneURLsAvailable   bool
	DestinationHash      string
	NomadDestinationHash string
	NodeName             string
	RngitPath            string
	RngitAvailable       bool
	RngitVersion         string
	RNSPackageVersion    string
	BuiltinServer        bool
	BuiltinRunning       bool
	SyncPermissions      bool
	PublicRead           bool
	PublicWrite          bool
	ServeNomadNet        bool
	ConfigPath           string
	RNSConfigPath        string
	NetworkConfigExists  bool
	OwnerGroupCount      int
}

// GetStatus returns a snapshot of the Reticulum integration state.
func GetStatus(ctx context.Context) *Status {
	status := &Status{
		Enabled:              setting.Reticulum.Enabled,
		CloneURLsAvailable:   Enabled(),
		DestinationHash:      setting.Reticulum.DestinationHash,
		NomadDestinationHash: setting.Reticulum.NomadDestinationHash,
		NodeName:             setting.Reticulum.NodeName,
		RngitPath:            setting.Reticulum.RngitPath,
		BuiltinServer:        setting.Reticulum.StartBuiltinServer,
		BuiltinRunning:       IsBuiltinRunning(),
		SyncPermissions:      setting.Reticulum.SyncPermissions,
		PublicRead:           setting.Reticulum.PublicRead,
		PublicWrite:          setting.Reticulum.PublicWrite,
		ServeNomadNet:        setting.Reticulum.ServeNomadNet,
		ConfigPath:           setting.Reticulum.ConfigPath,
		RNSConfigPath:        setting.Reticulum.RNSConfigPath,
		NetworkConfigExists:  networkConfigExists(),
	}

	if path, err := exec.LookPath(setting.Reticulum.RngitPath); err == nil {
		status.RngitAvailable = true
		status.RngitPath = path
		if out, err := exec.CommandContext(ctx, path, "--version").CombinedOutput(); err == nil {
			status.RngitVersion = strings.TrimSpace(string(out))
		}
	}
	if pkg := GetPackageInfo(ctx); pkg != nil {
		status.RNSPackageVersion = pkg.PipVersion
	}

	if owners, err := listOwnerDirectories(); err == nil {
		status.OwnerGroupCount = len(owners)
	}

	return status
}

func networkConfigExists() bool {
	_, err := os.Stat(filepath.Join(setting.Reticulum.RNSConfigPath, "config"))
	return err == nil
}

// ReadNetworkConfig returns the Reticulum network configuration file contents.
func ReadNetworkConfig() (string, error) {
	path := filepath.Join(setting.Reticulum.RNSConfigPath, "config")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return reticulum_module.DefaultNetworkConfig, nil
		}
		return "", fmt.Errorf("read reticulum network config: %w", err)
	}
	return string(data), nil
}

// WriteNetworkConfig writes the Reticulum network configuration file.
func WriteNetworkConfig(content string) error {
	dir := setting.Reticulum.RNSConfigPath
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create reticulum network config directory: %w", err)
	}
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write reticulum network config: %w", err)
	}
	return nil
}

// EnsureNetworkConfig creates the default Reticulum network config if missing.
func EnsureNetworkConfig() error {
	path := filepath.Join(setting.Reticulum.RNSConfigPath, "config")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat reticulum network config: %w", err)
	}
	return WriteNetworkConfig(reticulum_module.DefaultNetworkConfig)
}
