package ui

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Console struct {
	Output     []string
	Input      []rune
	Cursor     int
	History    []string
	HistoryPos int
	Scroll     int
	Active     bool
	Width      int
	Height     int
}

func NewConsole(width, height int) *Console {
	return &Console{
		Width:  width,
		Height: height,
	}
}

func (c *Console) Exec(cmdLine string) {
	c.History = append(c.History, cmdLine)
	c.HistoryPos = len(c.History)

	c.AddOutput(fmt.Sprintf("> %s", cmdLine))

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		c.AddOutput(fmt.Sprintf("error: %v", err))
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	for _, line := range lines {
		c.AddOutput(line)
	}
}

func (c *Console) AddOutput(line string) {
	c.Output = append(c.Output, line)
	c.Scroll = len(c.Output)
}

func (c *Console) RenderHeader() string {
	return "Console"
}

func (c *Console) RenderBody() string {
	bodyH := c.Height - 2
	if bodyH < 1 {
		bodyH = 1
	}
	outH := bodyH - 1
	if outH < 0 {
		outH = 0
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

	prompt := fmt.Sprintf(" > %s", string(c.Input))
	if len(prompt) > w {
		prompt = prompt[:w]
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
