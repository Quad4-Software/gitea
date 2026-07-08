// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitea.dev/modules/setting"
)

const maxLogLines = 1000

type logLine struct {
	seq  int
	text string
}

var (
	logMu          sync.Mutex
	logEntries     []logLine
	logSeq         int
	logWatchCancel context.CancelFunc
)

func appendLogLine(text string) {
	if text == "" {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	appendLogLineLocked(text)
}

func appendLogLineLocked(text string) {
	logSeq++
	logEntries = append(logEntries, logLine{seq: logSeq, text: text})
	if len(logEntries) > maxLogLines {
		logEntries = logEntries[len(logEntries)-maxLogLines:]
	}
}

// GetLogs returns rngit log lines with sequence numbers greater than since.
func GetLogs(since int) (lines []string, next int) {
	logMu.Lock()
	defer logMu.Unlock()
	if len(logEntries) == 0 {
		seedLogFileLocked(serverLogPath())
	}
	for _, entry := range logEntries {
		if entry.seq > since {
			lines = append(lines, entry.text)
		}
	}
	return lines, logSeq
}

func seedLogFileLocked(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil || stat.Size() == 0 {
		return
	}

	startOffset := int64(0)
	if stat.Size() > 256*1024 {
		startOffset = stat.Size() - 256*1024
	}
	_, _ = f.Seek(startOffset, io.SeekStart)
	if startOffset > 0 {
		_, _ = bufio.NewReader(f).ReadString('\n')
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		appendLogLineLocked(scanner.Text())
	}
}

func serverLogPath() string {
	return filepath.Join(setting.Reticulum.ConfigPath, "server_log")
}

func startLogWatcher(ctx context.Context) {
	stopLogWatcher()

	watchCtx, cancel := context.WithCancel(ctx)
	logWatchCancel = cancel

	go func() {
		path := serverLogPath()
		logMu.Lock()
		logEntries = nil
		logSeq = 0
		seedLogFileLocked(path)
		logMu.Unlock()
		appendLogLine("[gitea] following rngit server log")

		var offset int64
		for {
			select {
			case <-watchCtx.Done():
				return
			default:
			}

			offset = followLogFile(watchCtx, path, offset)
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

func followLogFile(ctx context.Context, path string, offset int64) int64 {
	f, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			appendLogLine(fmt.Sprintf("[gitea] cannot open rngit log: %v", err))
		}
		return offset
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return offset
	}
	if stat.Size() < offset {
		offset = 0
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return offset
	}

	reader := bufio.NewReader(f)
	for {
		select {
		case <-ctx.Done():
			return offset
		default:
		}

		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			appendLogLine(strings.TrimRight(line, "\r\n"))
			offset += int64(len(line))
		}
		if err == nil {
			continue
		}
		if err != io.EOF {
			return offset
		}

		select {
		case <-ctx.Done():
			return offset
		case <-time.After(500 * time.Millisecond):
		}

		newStat, err := os.Stat(path)
		if err != nil || newStat.Size() <= offset {
			continue
		}
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return offset
		}
		reader = bufio.NewReader(f)
	}
}

func stopLogWatcher() {
	if logWatchCancel != nil {
		logWatchCancel()
		logWatchCancel = nil
	}
}
