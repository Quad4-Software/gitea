// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	repo_model "gitea.dev/models/repo"
	reticulum_module "gitea.dev/modules/reticulum"
)

func extraAllowedPath(repo *repo_model.Repository) string {
	return strings.TrimSuffix(repo.RepoPath(), ".git") + ".rns_extra"
}

func mirrorSourcePath(repo *repo_model.Repository) string {
	return strings.TrimSuffix(repo.RepoPath(), ".git") + ".rns_mirror"
}

// ReadExtraAllowedLines returns manually configured permission lines for a repository.
func ReadExtraAllowedLines(repo *repo_model.Repository) ([]string, error) {
	path := extraAllowedPath(repo)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read reticulum extra permissions for %s: %w", repo.FullName(), err)
	}
	return parseAllowedLines(string(data)), nil
}

// WriteExtraAllowedLines stores manually configured permission lines for a repository.
func WriteExtraAllowedLines(repo *repo_model.Repository, lines []string) error {
	path := extraAllowedPath(repo)
	lines = normalizeAllowedLines(lines)
	if len(lines) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove reticulum extra permissions for %s: %w", repo.FullName(), err)
		}
		return nil
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write reticulum extra permissions for %s: %w", repo.FullName(), err)
	}
	return nil
}

// ReadMirrorSource returns the configured RNS mirror upstream URL for a repository.
func ReadMirrorSource(repo *repo_model.Repository) (string, error) {
	path := mirrorSourcePath(repo)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read reticulum mirror source for %s: %w", repo.FullName(), err)
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteMirrorSource stores the RNS mirror upstream URL for a repository.
func WriteMirrorSource(repo *repo_model.Repository, sourceURL string) error {
	path := mirrorSourcePath(repo)
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove reticulum mirror source for %s: %w", repo.FullName(), err)
		}
		return nil
	}
	if err := os.WriteFile(path, []byte(sourceURL+"\n"), 0o600); err != nil {
		return fmt.Errorf("write reticulum mirror source for %s: %w", repo.FullName(), err)
	}
	return nil
}

func parseAllowedLines(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func normalizeAllowedLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, line)
	}
	return uniqueLines(result)
}

func ParseExtraAllowedInput(content string) []string {
	lines := parseAllowedLines(content)
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, ":") {
			result = append(result, line)
			continue
		}
		if normalized, err := reticulum_module.NormalizeIdentityHash(line); err == nil {
			result = append(result, "r:"+normalized)
		}
	}
	return uniqueLines(result)
}
