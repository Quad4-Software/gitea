// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitea.dev/models/db"
	"gitea.dev/models/perm"
	access_model "gitea.dev/models/perm/access"
	repo_model "gitea.dev/models/repo"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/log"
	reticulum_module "gitea.dev/modules/reticulum"
	"gitea.dev/modules/setting"
)

// allowedFilePath returns the rngit .allowed file path for a repository.
func allowedFilePath(repo *repo_model.Repository) string {
	repoDir := repo.RepoPath()
	base := strings.TrimSuffix(repoDir, ".git")
	return base + ".allowed"
}

// SyncRepositoryPermissions writes the rngit .allowed file for a single repository.
func SyncRepositoryPermissions(ctx context.Context, repo *repo_model.Repository) error {
	if !setting.Reticulum.Enabled || !setting.Reticulum.SyncPermissions {
		return nil
	}

	lines, err := buildAllowedLines(ctx, repo)
	if err != nil {
		return err
	}

	allowedPath := allowedFilePath(repo)
	if len(lines) == 0 {
		if err := os.Remove(allowedPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove reticulum permissions for %s: %w", repo.FullName(), err)
		}
	} else {
		content := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(allowedPath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("write reticulum permissions for %s: %w", repo.FullName(), err)
		}
	}

	return ensureRngitRepoAlias(repo)
}

func ensureRngitRepoAlias(repo *repo_model.Repository) error {
	repoPath := repo.RepoPath()
	aliasPath := strings.TrimSuffix(repoPath, ".git")
	if aliasPath == repoPath {
		return nil
	}

	if _, err := os.Lstat(aliasPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat reticulum repo alias for %s: %w", repo.FullName(), err)
	}

	if err := os.Symlink(filepath.Base(repoPath), aliasPath); err != nil && !os.IsExist(err) {
		return fmt.Errorf("create reticulum repo alias for %s: %w", repo.FullName(), err)
	}
	return nil
}

func removeRngitRepoAlias(repo *repo_model.Repository) error {
	aliasPath := strings.TrimSuffix(repo.RepoPath(), ".git")
	if err := os.Remove(aliasPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove reticulum repo alias for %s: %w", repo.FullName(), err)
	}
	return nil
}

// SyncAllRepositoryPermissions updates rngit permission files for every repository.
func SyncAllRepositoryPermissions(ctx context.Context) error {
	if !setting.Reticulum.Enabled || !setting.Reticulum.SyncPermissions {
		return nil
	}

	var repos []*repo_model.Repository
	if err := db.GetEngine(ctx).Find(&repos); err != nil {
		return fmt.Errorf("list repositories for reticulum sync: %w", err)
	}
	for _, repo := range repos {
		if err := SyncRepositoryPermissions(ctx, repo); err != nil {
			return err
		}
	}
	return nil
}

// RemoveRepositoryPermissions deletes the rngit .allowed file for a repository.
func RemoveRepositoryPermissions(repo *repo_model.Repository) error {
	if !setting.Reticulum.Enabled {
		return nil
	}
	allowedPath := allowedFilePath(repo)
	if err := os.Remove(allowedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove reticulum permissions for %s: %w", repo.FullName(), err)
	}
	return removeRngitRepoAlias(repo)
}

func buildAllowedLines(ctx context.Context, repo *repo_model.Repository) ([]string, error) {
	var lines []string
	var err error
	if repo.IsPrivate {
		lines, err = buildPrivateAllowedLines(ctx, repo)
	} else {
		lines, err = buildPublicAllowedLines(ctx, repo)
	}
	if err != nil {
		return nil, err
	}
	extra, err := ReadExtraAllowedLines(repo)
	if err != nil {
		return nil, err
	}
	if len(extra) == 0 {
		return lines, nil
	}
	return uniqueLines(append(lines, extra...)), nil
}

func buildPublicAllowedLines(ctx context.Context, repo *repo_model.Repository) ([]string, error) {
	if !setting.Reticulum.PublicRead {
		return nil, nil
	}

	lines := []string{"r:all"}
	if setting.Reticulum.PublicWrite {
		lines = append(lines, "w:all")
	} else {
		writeIdentities, err := collectWriteIdentities(ctx, repo)
		if err != nil {
			return nil, err
		}
		for _, identity := range writeIdentities {
			lines = append(lines, "w:"+identity)
		}
	}
	return lines, nil
}

func buildPrivateAllowedLines(ctx context.Context, repo *repo_model.Repository) ([]string, error) {
	readIdentities, err := collectReadIdentities(ctx, repo)
	if err != nil {
		return nil, err
	}
	writeIdentities, err := collectWriteIdentities(ctx, repo)
	if err != nil {
		return nil, err
	}

	if len(readIdentities) == 0 && len(writeIdentities) == 0 {
		return nil, nil
	}

	lines := make([]string, 0, len(readIdentities)+len(writeIdentities))
	for _, identity := range readIdentities {
		lines = append(lines, "r:"+identity)
	}
	for _, identity := range writeIdentities {
		lines = append(lines, "w:"+identity)
	}
	return uniqueLines(lines), nil
}

func collectReadIdentities(ctx context.Context, repo *repo_model.Repository) ([]string, error) {
	return collectIdentities(ctx, repo, perm.AccessModeRead)
}

func collectWriteIdentities(ctx context.Context, repo *repo_model.Repository) ([]string, error) {
	return collectIdentities(ctx, repo, perm.AccessModeWrite)
}

func collectIdentities(ctx context.Context, repo *repo_model.Repository, minMode perm.AccessMode) ([]string, error) {
	identities := make(map[string]struct{})

	if err := repo.LoadOwner(ctx); err != nil {
		return nil, err
	}

	ownerIdentity, err := userIdentity(ctx, repo.OwnerID)
	if err != nil {
		return nil, err
	}
	if ownerIdentity != "" {
		identities[ownerIdentity] = struct{}{}
	}

	var accesses []access_model.Access
	if err := db.GetEngine(ctx).Where("repo_id = ? AND mode >= ?", repo.ID, minMode).Find(&accesses); err != nil {
		return nil, fmt.Errorf("list repository access for reticulum sync: %w", err)
	}
	for _, access := range accesses {
		identity, err := userIdentity(ctx, access.UserID)
		if err != nil {
			return nil, err
		}
		if identity != "" {
			identities[identity] = struct{}{}
		}
	}

	collaborators, _, err := repo_model.GetCollaborators(ctx, &repo_model.FindCollaborationOptions{
		RepoID: repo.ID,
	})
	if err != nil {
		return nil, err
	}

	for _, collab := range collaborators {
		if collab.Collaboration.Mode < minMode {
			continue
		}
		identity, err := userIdentity(ctx, collab.Collaboration.UserID)
		if err != nil {
			return nil, err
		}
		if identity != "" {
			identities[identity] = struct{}{}
		}
	}

	result := make([]string, 0, len(identities))
	for identity := range identities {
		result = append(result, identity)
	}
	return result, nil
}

func userIdentity(ctx context.Context, userID int64) (string, error) {
	value, err := user_model.GetUserSetting(ctx, userID, reticulum_module.SettingKeyIdentity)
	if err != nil || value == "" {
		return "", err
	}
	identity, err := reticulum_module.NormalizeIdentityHash(value)
	if err != nil {
		log.Warn("Reticulum: user %d has invalid identity hash %q: %v", userID, value, err)
		return "", nil
	}
	return identity, nil
}

func uniqueLines(lines []string) []string {
	seen := make(map[string]struct{}, len(lines))
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		result = append(result, line)
	}
	return result
}

// EnsureOwnerGroup is a no-op placeholder kept for owner directory creation hooks.
func EnsureOwnerGroup(ownerName string) error {
	ownerPath := filepath.Join(setting.RepoRootPath, strings.ToLower(ownerName))
	if err := os.MkdirAll(ownerPath, 0o755); err != nil {
		return fmt.Errorf("ensure owner repository directory: %w", err)
	}
	return nil
}
