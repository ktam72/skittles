//go:build darwin

package ui

import "os/exec"

func switchToEnglishInput() {
	_ = exec.Command("osascript", "-e",
		`tell application "System Events" to key code 102`,
	).Run()
}
