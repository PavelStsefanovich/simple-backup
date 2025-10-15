//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

// getFreeSpace retrieves the free disk space in bytes for the given path.
// This version is for Windows.
func getFreeSpace(path string) (uint64, error) {
	// The Windows API requires a pointer to a string with null termination.
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	var freeBytesAvailableToCaller uint64
	
	// Call the Windows function to get disk space info.
	// It's a C-style function, so we need to use unsafe.Pointer.
	_, _, err = syscall.NewLazyDLL("kernel32.dll").NewProc("GetDiskFreeSpaceExW").Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailableToCaller)),
		0,
		0,
	)
	
	if err != nil && err.Error() != "The operation completed successfully." {
		return 0, fmt.Errorf("failed to get free space for %s: %w", path, err)
	}

	return freeBytesAvailableToCaller, nil
}
