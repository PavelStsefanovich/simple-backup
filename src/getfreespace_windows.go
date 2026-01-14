//go:build windows

package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)



//////////////  INIT FUNCTIONS  ///////////////////////////////////////////////

func init() {
	// Fixes Virtual Terminal Processing in elevated terminal on Windows.
    if runtime.GOOS == "windows" {
        stdout := windows.Handle(os.Stdout.Fd())
        var originalMode uint32

        // Get the current console mode
        windows.GetConsoleMode(stdout, &originalMode)

        // Add the Virtual Terminal Processing flag
        // 0x0004 is the hex value for ENABLE_VIRTUAL_TERMINAL_PROCESSING
        newMode := originalMode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

        // Set the new mode
        windows.SetConsoleMode(stdout, newMode)
    }
}


// getFreeSpace retrieves the free disk space in bytes for the given path.
// This version is for Windows.
func getFreeSpace(path string) (uint64, string, error) {
	// The Windows API requires a pointer to a string with null termination.
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, "", fmt.Errorf("failed to convert path to UTF16: %w", err)
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
		return 0, "", fmt.Errorf("failed to get free space for %s: %w", path, err)
	}

	return freeBytesAvailableToCaller, formatBytes(freeBytesAvailableToCaller), nil
}
