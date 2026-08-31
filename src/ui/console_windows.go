//go:build windows

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
		} else if strings.HasPrefix(arg, "~") {
			home, _ := os.UserHomeDir()
			target = filepath.Join(home, arg[1:])
		} else if strings.Contains(arg, ":") {
			target = filepath.Clean(arg)
		} else {
			target = filepath.Join(c.Dir, arg)
		}
	} else {
		home, _ := os.UserHomeDir()
		target = filepath.Clean(home)
	}
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		c.Dir = target
	} else if len(parts) >= 2 {
		c.AddOutput(fmt.Sprintf("cd: %s: No such directory", parts[1]))
	}
	return true
}

func runExec(parts []string, cmdLine string, dir string, c *Console) {
	outputCh := make(chan string, 100)
	c.outputCh = outputCh

	go func() {
		defer close(outputCh)

		cmd := exec.Command("cmd.exe", "/c", cmdLine)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "PROMPT=$P$G")

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
	}()
}
