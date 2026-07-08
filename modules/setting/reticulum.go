// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"path/filepath"

	"gitea.dev/modules/log"
)

// Reticulum settings for rngit (Git over Reticulum) integration.
var Reticulum = struct {
	Enabled              bool   `ini:"ENABLED"`
	DestinationHash      string `ini:"DESTINATION_HASH"`
	NomadDestinationHash string `ini:"NOMAD_DESTINATION_HASH"`
	ConfigPath           string `ini:"CONFIG_PATH"`
	RNSConfigPath        string `ini:"RNS_CONFIG_PATH"`
	RngitPath            string `ini:"RNGIT_PATH"`
	StartBuiltinServer   bool   `ini:"START_BUILTIN_SERVER"`
	SyncPermissions      bool   `ini:"SYNC_PERMISSIONS"`
	PublicRead           bool   `ini:"PUBLIC_READ"`
	PublicWrite          bool   `ini:"PUBLIC_WRITE"`
	ServeNomadNet        bool   `ini:"SERVE_NOMADNET"`
	NodeName             string `ini:"NODE_NAME"`
}{
	Enabled:            false,
	ConfigPath:         "",
	RNSConfigPath:      "",
	RngitPath:          "rngit",
	StartBuiltinServer: false,
	SyncPermissions:    true,
	PublicRead:         true,
	PublicWrite:        false,
	ServeNomadNet:      false,
	NodeName:           "Gitea",
}

func loadReticulumFrom(rootCfg ConfigProvider) {
	if err := rootCfg.Section("reticulum").MapTo(&Reticulum); err != nil {
		log.Fatal("Failed to map Reticulum settings: %v", err)
	}

	if Reticulum.ConfigPath == "" {
		Reticulum.ConfigPath = filepath.Join(AppDataPath, "reticulum", "rngit")
	}
	if Reticulum.RNSConfigPath == "" {
		Reticulum.RNSConfigPath = filepath.Join(AppDataPath, "reticulum", "network")
	}
}
