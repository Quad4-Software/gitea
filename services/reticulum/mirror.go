// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	repo_model "gitea.dev/models/repo"
	reticulum_module "gitea.dev/modules/reticulum"
	"gitea.dev/modules/setting"
)

// ConfigureRepositoryMirror configures an existing repository to mirror an upstream rns:// URL.
func ConfigureRepositoryMirror(ctx context.Context, repo *repo_model.Repository, sourceURL string) error {
	if !setting.Reticulum.Enabled {
		return errors.New("reticulum is not enabled")
	}
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return errors.New("mirror source URL is required")
	}
	if _, err := reticulum_module.ParseMirrorURL(sourceURL); err != nil {
		return err
	}

	repoPath := repo.RepoPath()
	if err := runGitConfig(ctx, repoPath, "repository.rngit.type", "mirror"); err != nil {
		return err
	}
	if err := runGitConfig(ctx, repoPath, "repository.rngit.upstream.source", sourceURL); err != nil {
		return err
	}
	if err := WriteMirrorSource(repo, sourceURL); err != nil {
		return err
	}
	if err := ensureRngitRepoAlias(repo); err != nil {
		return err
	}
	if err := SyncRepositoryPermissions(ctx, repo); err != nil {
		return err
	}
	if err := syncRepositoryMirror(ctx, repoPath, sourceURL); err != nil {
		return err
	}
	return RestartBuiltinServer(ctx)
}

// SyncRepositoryMirror fetches the latest content for a configured RNS mirror repository.
func SyncRepositoryMirror(ctx context.Context, repo *repo_model.Repository) error {
	sourceURL, err := ReadMirrorSource(repo)
	if err != nil {
		return err
	}
	if sourceURL == "" {
		sourceURL = readMirrorSourceFromGit(ctx, repo.RepoPath())
	}
	if sourceURL == "" {
		return fmt.Errorf("repository %s has no RNS mirror source configured", repo.FullName())
	}
	if err := syncRepositoryMirror(ctx, repo.RepoPath(), sourceURL); err != nil {
		return err
	}
	return RestartBuiltinServer(ctx)
}

// MirrorRepositoryFromRNS creates a new mirrored repository using rngit mirror.
func MirrorRepositoryFromRNS(ctx context.Context, ownerName, repoName, sourceURL string) error {
	if !setting.Reticulum.Enabled {
		return errors.New("reticulum is not enabled")
	}
	if _, err := reticulum_module.ParseMirrorURL(sourceURL); err != nil {
		return err
	}
	targetURL := reticulum_module.ComposeLocalMirrorTarget(ownerName, repoName)
	if targetURL == "" {
		return errors.New("reticulum destination hash is not configured")
	}

	args := []string{"mirror", sourceURL, targetURL, "--config", setting.Reticulum.ConfigPath}
	if setting.Reticulum.RNSConfigPath != "" {
		args = append(args, "--rnsconfig", setting.Reticulum.RNSConfigPath)
	}
	if _, err := runRngit(ctx, args...); err != nil {
		return err
	}
	return RestartBuiltinServer(ctx)
}

func syncRepositoryMirror(ctx context.Context, repoPath, sourceURL string) error {
	cmd := exec.CommandContext(ctx, "git", "fetch", sourceURL, "+refs/*:refs/*")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(),
		"RNGIT_CONFIG="+setting.Reticulum.ConfigPath,
		"RNS_CONFIG="+setting.Reticulum.RNSConfigPath,
		"GIT_ALLOW_PROTOCOL=rns",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch %s: %w: %s", sourceURL, err, strings.TrimSpace(string(out)))
	}
	return runGitConfig(ctx, repoPath, "repository.rngit.upstream.sync", strconv.FormatInt(time.Now().Unix(), 10))
}

func readMirrorSourceFromGit(ctx context.Context, repoPath string) string {
	cmd := exec.CommandContext(ctx, "git", "config", "--get", "repository.rngit.upstream.source")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runGitConfig(ctx context.Context, repoPath, key, value string) error {
	cmd := exec.CommandContext(ctx, "git", "config", key, value)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config %s: %w: %s", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}
