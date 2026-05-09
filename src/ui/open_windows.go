//go:build windows

package ui

import (
	"os/exec"
	"strings"
)

func openFile(path string) error {
	cmd := exec.Command("cmd", "/c", "start", "", strings.ReplaceAll(path, "/", "\\"))
	return cmd.Start()
}
