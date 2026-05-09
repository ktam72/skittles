//go:build darwin

package main

import "os/exec"

func switchToEnglishInput() {
	_ = exec.Command("osascript", "-e",
		`tell application "System Events" to tell first process where frontmost is true to set input source to language code "en"`,
	).Run()
}
