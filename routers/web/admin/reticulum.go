// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"

	"gitea.dev/modules/setting"
	"gitea.dev/modules/templates"
	"gitea.dev/modules/web"
	"gitea.dev/services/context"
	"gitea.dev/services/forms"
	reticulum_service "gitea.dev/services/reticulum"
)

const (
	tplReticulum       templates.TplName = "admin/reticulum"
	tplReticulumStatus templates.TplName = "admin/reticulum_status"
)

func prepareReticulumPage(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.reticulum.title")
	ctx.Data["PageIsAdminReticulum"] = true
	ctx.Data["ReticulumStatus"] = reticulum_service.GetStatus(ctx)
}

// Reticulum shows the Reticulum management page.
func Reticulum(ctx *context.Context) {
	prepareReticulumPage(ctx)

	settings := reticulum_service.CurrentSettings()
	networkConfig, err := reticulum_service.ReadNetworkConfig()
	if err != nil {
		ctx.ServerError("ReadNetworkConfig", err)
		return
	}

	form := forms.AdminReticulumForm{
		Enabled:              settings.Enabled,
		DestinationHash:      settings.DestinationHash,
		NomadDestinationHash: settings.NomadDestinationHash,
		ConfigPath:           settings.ConfigPath,
		RNSConfigPath:        settings.RNSConfigPath,
		RngitPath:            settings.RngitPath,
		StartBuiltinServer:   settings.StartBuiltinServer,
		SyncPermissions:      settings.SyncPermissions,
		PublicRead:           settings.PublicRead,
		PublicWrite:          settings.PublicWrite,
		ServeNomadNet:        settings.ServeNomadNet,
		NodeName:             settings.NodeName,
		NetworkConfig:        networkConfig,
	}
	ctx.Data["ReticulumForm"] = form
	ctx.HTML(http.StatusOK, tplReticulum)
}

// ReticulumPost saves Reticulum settings and network configuration.
func ReticulumPost(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.AdminReticulumForm)

	settings := reticulum_service.Settings{
		Enabled:              ctx.FormBool("enabled"),
		DestinationHash:      form.DestinationHash,
		NomadDestinationHash: form.NomadDestinationHash,
		ConfigPath:           form.ConfigPath,
		RNSConfigPath:        form.RNSConfigPath,
		RngitPath:            form.RngitPath,
		StartBuiltinServer:   ctx.FormBool("start_builtin_server"),
		SyncPermissions:      ctx.FormBool("sync_permissions"),
		PublicRead:           ctx.FormBool("public_read"),
		PublicWrite:          ctx.FormBool("public_write"),
		ServeNomadNet:        ctx.FormBool("serve_nomadnet"),
		NodeName:             form.NodeName,
	}
	if settings.ConfigPath == "" {
		settings.ConfigPath = setting.Reticulum.ConfigPath
	}
	if settings.RNSConfigPath == "" {
		settings.RNSConfigPath = setting.Reticulum.RNSConfigPath
	}
	if settings.RngitPath == "" {
		settings.RngitPath = setting.Reticulum.RngitPath
	}
	if settings.NodeName == "" {
		settings.NodeName = "Gitea"
	}

	if err := reticulum_service.SaveAppIniSettings(settings); err != nil {
		ctx.ServerError("SaveAppIniSettings", err)
		return
	}
	if err := reticulum_service.WriteNetworkConfig(form.NetworkConfig); err != nil {
		ctx.ServerError("WriteNetworkConfig", err)
		return
	}

	reticulum_service.ApplySettingsToRuntime(settings)
	if setting.Reticulum.Enabled && setting.Reticulum.StartBuiltinServer {
		if err := reticulum_service.RestartBuiltinServer(ctx); err != nil {
			ctx.ServerError("RestartBuiltinServer", err)
			return
		}
	}

	ctx.Flash.Success(ctx.Tr("admin.reticulum.saved"))
	ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
}

// ReticulumStatus returns the live status partial for polling.
func ReticulumStatus(ctx *context.Context) {
	ctx.Data["ReticulumStatus"] = reticulum_service.GetStatus(ctx)
	ctx.HTML(http.StatusOK, tplReticulumStatus)
}

// ReticulumSync regenerates rngit config and permission files.
func ReticulumSync(ctx *context.Context) {
	if err := reticulum_service.ReloadSettingsFromAppIni(); err != nil {
		ctx.ServerError("ReloadSettingsFromAppIni", err)
		return
	}
	if err := reticulum_service.WriteConfig(ctx); err != nil {
		ctx.ServerError("WriteConfig", err)
		return
	}
	if err := reticulum_service.SyncAllRepositoryPermissions(ctx); err != nil {
		ctx.ServerError("SyncAllRepositoryPermissions", err)
		return
	}
	if err := reticulum_service.RestartBuiltinServer(ctx); err != nil {
		ctx.ServerError("RestartBuiltinServer", err)
		return
	}
	ctx.Flash.Success(ctx.Tr("admin.reticulum.sync_success"))
	ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
}

// ReticulumRestart restarts the builtin rngit server.
func ReticulumRestart(ctx *context.Context) {
	if err := reticulum_service.ReloadSettingsFromAppIni(); err != nil {
		ctx.ServerError("ReloadSettingsFromAppIni", err)
		return
	}
	if err := reticulum_service.RestartBuiltinServer(ctx); err != nil {
		ctx.ServerError("RestartBuiltinServer", err)
		return
	}
	ctx.Flash.Success(ctx.Tr("admin.reticulum.restart_success"))
	ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
}

// ReticulumResolveHash attempts to resolve repository and Nomad Network destination hashes.
func ReticulumResolveHash(ctx *context.Context) {
	info, err := reticulum_service.ResolveIdentities(ctx)
	if err != nil {
		ctx.ServerError("ResolveIdentities", err)
		return
	}

	settings := reticulum_service.CurrentSettings()
	settings.DestinationHash = info.RepositoriesDestinationHash
	settings.NomadDestinationHash = info.NomadDestinationHash
	if err := reticulum_service.SaveAppIniSettings(settings); err != nil {
		ctx.ServerError("SaveAppIniSettings", err)
		return
	}

	reticulum_service.ApplySettingsToRuntime(settings)
	if info.NomadDestinationHash != "" {
		ctx.Flash.Success(ctx.Tr("admin.reticulum.destinations_resolved", info.RepositoriesDestinationHash, info.NomadDestinationHash))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.reticulum.destination_resolved", info.RepositoriesDestinationHash))
	}
	ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
}

// ReticulumUpdateRNS upgrades the RNS Python package.
func ReticulumUpdateRNS(ctx *context.Context) {
	output, err := reticulum_service.UpdateRNSPackage(ctx)
	if err != nil {
		ctx.Flash.Error(err.Error())
		if output != "" {
			ctx.Flash.Info(output)
		}
		ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
		return
	}
	ctx.Flash.Success(ctx.Tr("admin.reticulum.rns_updated"))
	ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
}

// ReticulumMirror creates or updates an RNS mirror repository.
func ReticulumMirror(ctx *context.Context) {
	owner := ctx.FormTrim("mirror_owner")
	repoName := ctx.FormTrim("mirror_repo")
	sourceURL := ctx.FormTrim("mirror_source")
	if owner == "" || repoName == "" || sourceURL == "" {
		ctx.Flash.Error(ctx.Tr("admin.reticulum.mirror_missing_fields"))
		ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
		return
	}
	if err := reticulum_service.MirrorRepositoryFromRNS(ctx, owner, repoName, sourceURL); err != nil {
		ctx.ServerError("MirrorRepositoryFromRNS", err)
		return
	}
	ctx.Flash.Success(ctx.Tr("admin.reticulum.mirror_success", owner+"/"+repoName))
	ctx.Redirect(setting.AppSubURL + "/-/admin/reticulum")
}

// ReticulumLogs returns rngit log lines for the admin console.
func ReticulumLogs(ctx *context.Context) {
	since := ctx.FormInt("since")
	lines, next := reticulum_service.GetLogs(since)
	ctx.JSON(http.StatusOK, map[string]any{
		"lines": lines,
		"next":  next,
	})
}
