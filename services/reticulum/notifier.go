// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"

	repo_model "gitea.dev/models/repo"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/graceful"
	"gitea.dev/modules/log"
	"gitea.dev/modules/setting"
	notify_service "gitea.dev/services/notify"
)

func init() {
	notify_service.RegisterNotifier(NewNotifier())
}

type reticulumNotifier struct {
	notify_service.NullNotifier
}

var _ notify_service.Notifier = &reticulumNotifier{}

// NewNotifier creates a notifier that keeps rngit in sync with Gitea repositories.
func NewNotifier() notify_service.Notifier {
	return &reticulumNotifier{}
}

func (n *reticulumNotifier) onRepoChanged(ctx context.Context, repo *repo_model.Repository) {
	if !setting.Reticulum.Enabled {
		return
	}
	if err := SyncRepositoryPermissions(ctx, repo); err != nil {
		log.Error("Reticulum: sync permissions for %s: %v", repo.FullName(), err)
	}
	if err := RestartBuiltinServer(ctx); err != nil {
		log.Error("Reticulum: restart builtin server after %s change: %v", repo.FullName(), err)
	}
}

func (n *reticulumNotifier) CreateRepository(ctx context.Context, doer, u *user_model.User, repo *repo_model.Repository) {
	n.onRepoChanged(ctx, repo)
}

func (n *reticulumNotifier) AdoptRepository(ctx context.Context, doer, u *user_model.User, repo *repo_model.Repository) {
	n.onRepoChanged(ctx, repo)
}

func (n *reticulumNotifier) MigrateRepository(ctx context.Context, doer, u *user_model.User, repo *repo_model.Repository) {
	n.onRepoChanged(ctx, repo)
}

func (n *reticulumNotifier) ForkRepository(ctx context.Context, doer *user_model.User, oldRepo, repo *repo_model.Repository) {
	n.onRepoChanged(ctx, repo)
}

func (n *reticulumNotifier) DeleteRepository(ctx context.Context, doer *user_model.User, repo *repo_model.Repository) {
	if !setting.Reticulum.Enabled {
		return
	}
	if err := RemoveRepositoryPermissions(repo); err != nil {
		log.Error("Reticulum: remove permissions for %s: %v", repo.FullName(), err)
	}
	if err := RestartBuiltinServer(graceful.GetManager().ShutdownContext()); err != nil {
		log.Error("Reticulum: restart builtin server after %s deletion: %v", repo.FullName(), err)
	}
}

func (n *reticulumNotifier) RenameRepository(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, oldRepoName string) {
	n.onRepoChanged(ctx, repo)
}

func (n *reticulumNotifier) TransferRepository(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, oldOwnerName string) {
	n.onRepoChanged(ctx, repo)
}

// OnRepositoryUpdated should be called when repository metadata such as visibility changes.
func OnRepositoryUpdated(ctx context.Context, repo *repo_model.Repository) {
	(&reticulumNotifier{}).onRepoChanged(ctx, repo)
}

// OnCollaborationUpdated should be called when repository collaborator access changes.
func OnCollaborationUpdated(ctx context.Context, repo *repo_model.Repository) {
	(&reticulumNotifier{}).onRepoChanged(ctx, repo)
}
