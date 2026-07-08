// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forms

import (
	"net/http"

	"gitea.dev/modules/web/middleware"
	"gitea.dev/services/context"

	"gitea.com/go-chi/binding"
)

// AdminReticulumForm is used to manage Reticulum settings in the admin panel.
type AdminReticulumForm struct {
	Enabled              bool
	DestinationHash      string `binding:"MaxSize(64)"`
	NomadDestinationHash string `binding:"MaxSize(64)"`
	ConfigPath           string `binding:"MaxSize(4096)"`
	RNSConfigPath        string `binding:"MaxSize(4096)"`
	RngitPath            string `binding:"MaxSize(4096)"`
	StartBuiltinServer   bool
	SyncPermissions      bool
	PublicRead           bool
	PublicWrite          bool
	ServeNomadNet        bool
	ShowReticulumThanks  bool
	SyncReticulumThanks  bool
	NodeName             string `binding:"MaxSize(128)"`
	NetworkConfig        string
}

// Validate validates the fields
func (f *AdminReticulumForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}
