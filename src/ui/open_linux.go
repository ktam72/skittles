//go:build linux

package ui

import "os/exec"

func openFile(path string) error {
	return exec.Command("xdg-open", path).Start()
}
