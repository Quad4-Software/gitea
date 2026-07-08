// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"net/http"
	"strings"

	"gitea.dev/modules/setting"
	"gitea.dev/services/context"
	reticulum_service "gitea.dev/services/reticulum"
)

func prepareReticulumRepoSettings(ctx *context.Context) {
	if !setting.Reticulum.Enabled {
		return
	}
	repo := ctx.Repo.Repository
	extra, err := reticulum_service.ReadExtraAllowedLines(repo)
	if err != nil {
		ctx.ServerError("ReadExtraAllowedLines", err)
		return
	}
	mirrorSource, err := reticulum_service.ReadMirrorSource(repo)
	if err != nil {
		ctx.ServerError("ReadMirrorSource", err)
		return
	}
	ctx.Data["ReticulumEnabled"] = true
	ctx.Data["ReticulumExtraAllowed"] = strings.Join(extra, "\n")
	ctx.Data["ReticulumMirrorSource"] = mirrorSource
}

func handleSettingsPostReticulum(ctx *context.Context) {
	if !setting.Reticulum.Enabled {
		ctx.HTTPError(http.StatusNotFound)
		return
	}
	repo := ctx.Repo.Repository
	extraInput := ctx.FormString("reticulum_extra_allowed")
	lines := reticulum_service.ParseExtraAllowedInput(extraInput)
	if err := reticulum_service.WriteExtraAllowedLines(repo, lines); err != nil {
		ctx.ServerError("WriteExtraAllowedLines", err)
		return
	}

	mirrorSource := strings.TrimSpace(ctx.FormString("reticulum_mirror_source"))
	if mirrorSource != "" {
		if err := reticulum_service.ConfigureRepositoryMirror(ctx, repo, mirrorSource); err != nil {
			ctx.ServerError("ConfigureRepositoryMirror", err)
			return
		}
	} else {
		if err := reticulum_service.WriteMirrorSource(repo, ""); err != nil {
			ctx.ServerError("WriteMirrorSource", err)
			return
		}
		if err := reticulum_service.SyncRepositoryPermissions(ctx, repo); err != nil {
			ctx.ServerError("SyncRepositoryPermissions", err)
			return
		}
		if err := reticulum_service.RestartBuiltinServer(ctx); err != nil {
			ctx.ServerError("RestartBuiltinServer", err)
			return
		}
	}

	ctx.Flash.Success(ctx.Tr("repo.settings.reticulum_saved"))
	ctx.Redirect(ctx.Repo.RepoLink + "/settings")
}

func handleSettingsPostReticulumMirrorSync(ctx *context.Context) {
	if !setting.Reticulum.Enabled {
		ctx.HTTPError(http.StatusNotFound)
		return
	}
	if err := reticulum_service.SyncRepositoryMirror(ctx, ctx.Repo.Repository); err != nil {
		ctx.ServerError("SyncRepositoryMirror", err)
		return
	}
	ctx.Flash.Success(ctx.Tr("repo.settings.reticulum_mirror_synced"))
	ctx.Redirect(ctx.Repo.RepoLink + "/settings")
}
