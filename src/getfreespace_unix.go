//go:build !windows

package main

import (
	"fmt"
	"syscall"
)

// getFreeSpace retrieves the free disk space in bytes for the given path.
// This version is for Unix-like systems (Linux, macOS).
func getFreeSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to get free space for %s: %w", path, err)
	}

	// Bsize is int64 on some platforms, so convert it to uint64 for multiplication
	freeSpace := uint64(stat.Bsize) * stat.Bavail

	return freeSpace, nil
}