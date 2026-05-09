//go:build windows

package ui

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	shell32   = syscall.NewLazyDLL("shell32.dll")
	shellExec = shell32.NewProc("ShellExecuteW")
)

func openFile(path string) error {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("convert path: %w", err)
	}
	openPtr, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return fmt.Errorf("convert verb: %w", err)
	}
	// ShellExecuteW(hwnd, lpVerb, lpFile, lpParameters, lpDirectory, nShowCmd)
	ret, _, _ := shellExec.Call(
		0,
		uintptr(unsafe.Pointer(openPtr)),
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		0,
		5, // SW_SHOW
	)
	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed: %d", ret)
	}
	return nil
}
