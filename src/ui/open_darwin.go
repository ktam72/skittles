//go:build darwin

package ui

import "os/exec"

func openFile(path string) error {
	return exec.Command("open", path).Start()
}
