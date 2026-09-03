//go:build unix

package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func handleBuiltin(parts []string, c *Console) bool {
	switch parts[0] {
	case "sudo":
		c.AddOutput("sudo: 権限昇格は禁止されています")
		return true
	case "clear":
		c.Output = nil
		c.outputBytes = 0
		c.Scroll = 0
		return true
	case "cd":
		return handleCdBuiltin(parts, c)
	default:
		return false
	}
}

func handleCdBuiltin(parts []string, c *Console) bool {
	target := c.Dir
	if len(parts) >= 2 {
		arg := parts[1]
		if arg == "-" {
		} else if strings.HasPrefix(arg, "~/") {
			home, _ := os.UserHomeDir()
			target = home + arg[1:]
		} else if arg == "~" {
			home, _ := os.UserHomeDir()
			target = home
		} else if strings.HasPrefix(arg, "/") {
			target = parts[1]
		} else {
			target = c.Dir + "/" + parts[1]
		}
	} else {
		home, _ := os.UserHomeDir()
		target = home
	}
	target = filepath.Clean(target)
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		c.Dir = target
	} else if len(parts) >= 2 {
		c.AddOutput(fmt.Sprintf("cd: %s: No such directory", parts[1]))
	}
	return true
}

func runExec(parts []string, cmdLine string, dir string, c *Console, outputCh chan string) {
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		outputCh <- fmt.Sprintf("error: %v", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		outputCh <- fmt.Sprintf("error: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		outputCh <- fmt.Sprintf("error: %v", err)
		return
	}

	done := make(chan struct{})
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			outputCh <- sc.Text()
		}
		close(done)
	}()

	scErr := bufio.NewScanner(stderr)
	for scErr.Scan() {
		outputCh <- scErr.Text()
	}

	<-done
	_ = cmd.Wait()
}
