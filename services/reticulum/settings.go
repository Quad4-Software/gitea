// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"strconv"

	"gitea.dev/modules/setting"
)

// Settings holds editable Reticulum options stored in app.ini.
type Settings struct {
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
	ShowReticulumThanks  bool   `ini:"SHOW_RETICULUM_THANKS"`
	SyncReticulumThanks  bool   `ini:"SYNC_RETICULUM_THANKS"`
	NodeName             string `ini:"NODE_NAME"`
}

// CurrentSettings returns the active Reticulum settings.
func CurrentSettings() Settings {
	return Settings{
		Enabled:              setting.Reticulum.Enabled,
		DestinationHash:      setting.Reticulum.DestinationHash,
		NomadDestinationHash: setting.Reticulum.NomadDestinationHash,
		ConfigPath:           setting.Reticulum.ConfigPath,
		RNSConfigPath:        setting.Reticulum.RNSConfigPath,
		RngitPath:            setting.Reticulum.RngitPath,
		StartBuiltinServer:   setting.Reticulum.StartBuiltinServer,
		SyncPermissions:      setting.Reticulum.SyncPermissions,
		PublicRead:           setting.Reticulum.PublicRead,
		PublicWrite:          setting.Reticulum.PublicWrite,
		ServeNomadNet:        setting.Reticulum.ServeNomadNet,
		ShowReticulumThanks:  setting.Reticulum.ShowReticulumThanks,
		SyncReticulumThanks:  setting.Reticulum.SyncReticulumThanks,
		NodeName:             setting.Reticulum.NodeName,
	}
}

// ApplySettingsToProvider writes Reticulum settings into a config provider.
func ApplySettingsToProvider(cfg setting.ConfigProvider, s Settings) {
	sec := cfg.Section("reticulum")
	sec.Key("ENABLED").SetValue(strconv.FormatBool(s.Enabled))
	sec.Key("DESTINATION_HASH").SetValue(s.DestinationHash)
	sec.Key("NOMAD_DESTINATION_HASH").SetValue(s.NomadDestinationHash)
	if s.ConfigPath != "" {
		sec.Key("CONFIG_PATH").SetValue(s.ConfigPath)
	}
	if s.RNSConfigPath != "" {
		sec.Key("RNS_CONFIG_PATH").SetValue(s.RNSConfigPath)
	}
	if s.RngitPath != "" {
		sec.Key("RNGIT_PATH").SetValue(s.RngitPath)
	}
	sec.Key("START_BUILTIN_SERVER").SetValue(strconv.FormatBool(s.StartBuiltinServer))
	sec.Key("SYNC_PERMISSIONS").SetValue(strconv.FormatBool(s.SyncPermissions))
	sec.Key("PUBLIC_READ").SetValue(strconv.FormatBool(s.PublicRead))
	sec.Key("PUBLIC_WRITE").SetValue(strconv.FormatBool(s.PublicWrite))
	sec.Key("SERVE_NOMADNET").SetValue(strconv.FormatBool(s.ServeNomadNet))
	sec.Key("SHOW_RETICULUM_THANKS").SetValue(strconv.FormatBool(s.ShowReticulumThanks))
	sec.Key("SYNC_RETICULUM_THANKS").SetValue(strconv.FormatBool(s.SyncReticulumThanks))
	if s.NodeName != "" {
		sec.Key("NODE_NAME").SetValue(s.NodeName)
	}
}

// SaveAppIniSettings persists Reticulum settings to app.ini.
func SaveAppIniSettings(s Settings) error {
	cfg, err := setting.NewConfigProviderFromFile(setting.CustomConf)
	if err != nil {
		return err
	}
	ApplySettingsToProvider(cfg, s)
	return cfg.SaveTo(setting.CustomConf)
}

// ApplySettingsToRuntime updates the in-process Reticulum settings.
func ApplySettingsToRuntime(s Settings) {
	setting.Reticulum.Enabled = s.Enabled
	setting.Reticulum.DestinationHash = s.DestinationHash
	setting.Reticulum.NomadDestinationHash = s.NomadDestinationHash
	if s.ConfigPath != "" {
		setting.Reticulum.ConfigPath = s.ConfigPath
	}
	if s.RNSConfigPath != "" {
		setting.Reticulum.RNSConfigPath = s.RNSConfigPath
	}
	if s.RngitPath != "" {
		setting.Reticulum.RngitPath = s.RngitPath
	}
	setting.Reticulum.StartBuiltinServer = s.StartBuiltinServer
	setting.Reticulum.SyncPermissions = s.SyncPermissions
	setting.Reticulum.PublicRead = s.PublicRead
	setting.Reticulum.PublicWrite = s.PublicWrite
	setting.Reticulum.ServeNomadNet = s.ServeNomadNet
	setting.Reticulum.ShowReticulumThanks = s.ShowReticulumThanks
	setting.Reticulum.SyncReticulumThanks = s.SyncReticulumThanks
	if s.NodeName != "" {
		setting.Reticulum.NodeName = s.NodeName
	}
}

// ReloadSettingsFromAppIni reloads Reticulum settings from app.ini into memory.
func ReloadSettingsFromAppIni() error {
	cfg, err := setting.NewConfigProviderFromFile(setting.CustomConf)
	if err != nil {
		return err
	}
	var stored Settings
	if err := cfg.Section("reticulum").MapTo(&stored); err != nil {
		return err
	}
	ApplySettingsToRuntime(stored)
	return nil
}
