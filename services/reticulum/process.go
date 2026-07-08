// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"gitea.dev/modules/graceful"
	"gitea.dev/modules/log"
	"gitea.dev/modules/process"
	"gitea.dev/modules/setting"
)

var (
	processMu     sync.Mutex
	rngitCmd      *exec.Cmd
	rngitFinished chan struct{}
)

func runRngit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, setting.Reticulum.RngitPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func startBuiltinServer() error {
	processMu.Lock()
	defer processMu.Unlock()

	if rngitCmd != nil {
		return nil
	}

	args := []string{"-s", "--config", setting.Reticulum.ConfigPath}
	if setting.Reticulum.RNSConfigPath != "" {
		args = append(args, "--rnsconfig", setting.Reticulum.RNSConfigPath)
	}

	// Use the process shutdown context, not an HTTP request context, so rngit keeps running after admin actions.
	rngitCmd = exec.CommandContext(graceful.GetManager().ShutdownContext(), setting.Reticulum.RngitPath, args...)
	rngitCmd.Dir = setting.Reticulum.ConfigPath
	rngitFinished = make(chan struct{})

	if err := rngitCmd.Start(); err != nil {
		rngitCmd = nil
		rngitFinished = nil
		return fmt.Errorf("start rngit: %w", err)
	}

	go func() {
		defer close(rngitFinished)
		if err := rngitCmd.Wait(); err != nil {
			log.Error("Reticulum: rngit process exited: %v", err)
		}
		processMu.Lock()
		rngitCmd = nil
		processMu.Unlock()
	}()

	log.Info("Reticulum: started builtin rngit server (config: %s)", setting.Reticulum.ConfigPath)
	startLogWatcher(graceful.GetManager().ShutdownContext())
	return nil
}

func stopBuiltinServer() {
	stopLogWatcher()

	processMu.Lock()
	cmd := rngitCmd
	finished := rngitFinished
	processMu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		_ = cmd.Process.Kill()
	}
	if finished != nil {
		<-finished
	}
}

// RestartBuiltinServer regenerates config and restarts the builtin rngit process.
func RestartBuiltinServer(ctx context.Context) error {
	if !setting.Reticulum.Enabled || !setting.Reticulum.StartBuiltinServer {
		return nil
	}

	stopBuiltinServer()

	if err := WriteConfig(ctx); err != nil {
		return err
	}

	if setting.Reticulum.DestinationHash == "" {
		info, err := ResolveIdentities(ctx)
		if err != nil {
			log.Warn("Reticulum: could not resolve destination hashes: %v", err)
		} else {
			setting.Reticulum.DestinationHash = info.RepositoriesDestinationHash
			setting.Reticulum.NomadDestinationHash = info.NomadDestinationHash
			log.Info("Reticulum: repositories destination is <%s>", info.RepositoriesDestinationHash)
			if info.NomadDestinationHash != "" {
				log.Info("Reticulum: nomad network destination is <%s>", info.NomadDestinationHash)
			}
		}
	}

	return startBuiltinServer()
}

// IsBuiltinRunning reports whether the builtin rngit subprocess is running.
func IsBuiltinRunning() bool {
	processMu.Lock()
	defer processMu.Unlock()
	return rngitCmd != nil && rngitCmd.Process != nil
}

func ensureConfigDir() error {
	return os.MkdirAll(setting.Reticulum.ConfigPath, 0o700)
}

func initBuiltinShutdown() {
	graceful.GetManager().RunAtShutdown(context.Background(), stopBuiltinServer)
}

func registerProcessContext(ctx context.Context) {
	_, _, _ = process.GetManager().AddTypedContext(ctx, "Service: Reticulum rngit", process.SystemProcessType, true)
}
