package scenario

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type WatchOptions struct {
	SnapshotDir string
	Interval    time.Duration
	Once        bool
	Diff        DiffOptions
}

type WatchEvent struct {
	Source   string      `json:"source"`
	Snapshot string      `json:"snapshot"`
	Previous string      `json:"previous,omitempty"`
	Diff     *DiffReport `json:"diff,omitempty"`
}

// WatchLocal observes one local scenario file. It uses a portable stat-based
// loop so Kit has no OS-specific or remote-machine dependency. Callers can
// select Once for a single snapshot in scripts and tests.
func WatchLocal(path string, opts WatchOptions, emit func(WatchEvent) error) error {
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	if opts.SnapshotDir == "" {
		opts.SnapshotDir = filepath.Join(filepath.Dir(path), ".kit-snapshots")
	}
	if err := os.MkdirAll(opts.SnapshotDir, 0o755); err != nil {
		return err
	}
	var previousPath string
	var previousMod time.Time
	for {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if previousPath == "" || info.ModTime() != previousMod || info.Size() != fileSize(previousPath) {
			snapshot := filepath.Join(opts.SnapshotDir, fmt.Sprintf("%s-%d.aoe2scenario", filepath.Base(path), info.ModTime().UnixNano()))
			if err := copyFile(path, snapshot); err != nil {
				return err
			}
			event := WatchEvent{Source: path, Snapshot: snapshot, Previous: previousPath}
			if previousPath != "" {
				diff, err := DiffFilesWithOptions(previousPath, snapshot, opts.Diff)
				if err != nil {
					return err
				}
				event.Diff = &diff
			}
			if err := emit(event); err != nil {
				return err
			}
			previousPath = snapshot
			previousMod = info.ModTime()
		}
		if opts.Once {
			return nil
		}
		time.Sleep(opts.Interval)
	}
}

func fileSize(path string) int64 {
	if path == "" {
		return -1
	}
	info, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return info.Size()
}

func copyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
