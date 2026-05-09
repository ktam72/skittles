package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type consoleOutputMsg string

type Console struct {
	Output       []string
	Input        []rune
	Cursor       int
	Dir          string
	History      []string
	HistoryPos   int
	Scroll       int
	Active       bool
	Width        int
	Height       int
	outputCh     chan string
	outputBytes  int
	maxOutputBytes int
}

func NewConsole(width, height int) *Console {
	dir, _ := os.Getwd()
	return &Console{
		Width:         width,
		Height:        height,
		Dir:           dir,
		maxOutputBytes: 102400,
	}
}

func (c *Console) Exec(cmdLine string) {
	c.History = append(c.History, cmdLine)
	c.HistoryPos = len(c.History)

	c.AddOutput(fmt.Sprintf("$ %s", cmdLine))

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return
	}

	// handle clear builtin
	if parts[0] == "clear" {
		c.Output = nil
		c.outputBytes = 0
		c.Scroll = 0
		return
	}

	// handle cd builtin
	if parts[0] == "cd" {
		target := c.Dir
		if len(parts) >= 2 {
			if parts[1] == "-" {
			} else if strings.HasPrefix(parts[1], "/") {
				target = parts[1]
			} else {
				target = c.Dir + "/" + parts[1]
			}
		} else {
			target, _ = os.UserHomeDir()
		}
		target = filepath.Clean(target)
		if fi, err := os.Stat(target); err == nil && fi.IsDir() {
			c.Dir = target
		} else if len(parts) >= 2 {
			c.AddOutput(fmt.Sprintf("cd: %s: No such directory", parts[1]))
		}
		return
	}

	outputCh := make(chan string, 100)
	c.outputCh = outputCh

	go func() {
		defer close(outputCh)

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = c.Dir

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

func (c *Console) AddOutput(line string) {
	c.Output = append(c.Output, line)
	c.outputBytes += len(line) + 1
	c.Scroll = len(c.Output)

	for c.outputBytes > c.maxOutputBytes && len(c.Output) > 1 {
		c.outputBytes -= len(c.Output[0]) + 1
		c.Output = c.Output[1:]
		c.Scroll--
		if c.Scroll < 0 {
			c.Scroll = 0
		}
	}
}

func (c *Console) RenderHeader() string {
	return "Console"
}

func (c *Console) RenderBody(cursorOn bool) string {
	bodyH := c.Height - 2
	if bodyH < 1 {
		bodyH = 1
	}
	outH := bodyH - 1
	if outH < 0 {
		outH = 0
	}

	cursor := " "
	if cursorOn {
		cursor = "█"
	}

	w := c.Width - 2
	lineFmt := fmt.Sprintf(" %%-%ds\n", w-1)

	var body strings.Builder
	start := c.Scroll - outH
	if start < 0 {
		start = 0
	}
	for i := start; i < len(c.Output) && i < start+outH; i++ {
		line := c.Output[i]
		if len(line) > w-2 {
			line = line[:w-5] + "..."
		}
		fmt.Fprintf(&body, lineFmt, line)
	}
	remain := outH - (len(c.Output) - start)
	for i := 0; i < remain; i++ {
		body.WriteString(strings.Repeat(" ", w) + "\n")
	}

	text := string(c.Input)
	prompt := fmt.Sprintf(" %s $ %s%s", c.Dir, text, cursor)
	if len(prompt) > w-1 {
		short := fmt.Sprintf(" $ %s%s", text, cursor)
		if len(short) > w-1 {
			prompt = short[:w-1]
		} else {
			prompt = short
		}
	}
	body.WriteString(prompt)
	return body.String()
}

func (c *Console) InsertRune(r rune) {
	c.Input = append(c.Input[:c.Cursor], append([]rune{r}, c.Input[c.Cursor:]...)...)
	c.Cursor++
}

func (c *Console) DeleteRune() {
	if c.Cursor > 0 {
		c.Input = append(c.Input[:c.Cursor-1], c.Input[c.Cursor:]...)
		c.Cursor--
	}
}

func (c *Console) ClearInput() {
	c.Input = nil
	c.Cursor = 0
}

func (c *Console) PrevHistory() {
	if len(c.History) == 0 {
		return
	}
	c.HistoryPos--
	if c.HistoryPos < 0 {
		c.HistoryPos = 0
	}
	c.Input = []rune(c.History[c.HistoryPos])
	c.Cursor = len(c.Input)
}

func (c *Console) NextHistory() {
	c.HistoryPos++
	if c.HistoryPos >= len(c.History) {
		c.HistoryPos = len(c.History)
		c.Input = nil
		c.Cursor = 0
		return
	}
	c.Input = []rune(c.History[c.HistoryPos])
	c.Cursor = len(c.Input)
}
