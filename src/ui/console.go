package ui

import (
	"fmt"
	"os"
	"strings"
)

type consoleOutputMsg string

type Console struct {
	Output         []string
	Input          []rune
	Cursor         int
	Dir            string
	History        []string
	HistoryPos     int
	Scroll         int
	Active         bool
	Width          int
	Height         int
	outputCh       chan string
	outputBytes    int
	maxOutputBytes int
}

func NewConsole(width, height int) *Console {
	dir, _ := os.Getwd()
	return &Console{
		Width:          width,
		Height:         height,
		Dir:            dir,
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

	if handleBuiltin(parts, c) {
		return
	}

	outputCh := make(chan string, 100)
	c.outputCh = outputCh

	go func() {
		defer close(outputCh)
		runExec(parts, cmdLine, c.Dir, c, outputCh)
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

	w := c.Width - 3

	// scrollbar
	totalH := len(c.Output)
	barH := outH
	thumbH := 1
	if totalH > barH {
		thumbH = max(1, barH*barH/totalH)
	}
	thumbPos := 0
	if totalH > barH {
		thumbPos = (c.Scroll - outH) * (barH - thumbH) / (totalH - barH)
		if thumbPos < 0 {
			thumbPos = 0
		}
	}

	lineFmt := fmt.Sprintf(" %%-%ds", w-1)

	var body strings.Builder

	// output lines (outH-1 rows; 最終行はプロンプト行が使う)
	maxOut := outH - 1
	if maxOut < 0 {
		maxOut = 0
	}
	// 表示可能行数は maxOut。start を outH 基準にすると最新の 1 行が
	// 常に描画範囲外へ落ちるので maxOut 基準で算出する。
	start := c.Scroll - maxOut
	if start < 0 {
		start = 0
	}
	for i := 0; i < maxOut; i++ {
		idx := start + i
		if idx < len(c.Output) {
			line := c.Output[idx]
			if len(line) > w-2 {
				line = line[:w-5] + "..."
			}
			fmt.Fprintf(&body, lineFmt, line)
		} else {
			fmt.Fprintf(&body, lineFmt, "")
		}
		if i >= thumbPos && i < thumbPos+thumbH {
			body.WriteString("█\n")
		} else {
			body.WriteString("│\n")
		}
	}

	// prompt line (last row)
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
	fmt.Fprintf(&body, "%-*s", w-1, prompt)
	totalH = len(c.Output)
	if totalH > 0 {
		ratio := float64(c.Scroll) / float64(totalH)
		barPos := int(ratio * float64(outH))
		if barPos >= outH {
			barPos = outH - 1
		}
		if barPos == maxOut {
			body.WriteString("█")
		} else {
			body.WriteString("│")
		}
	} else {
		body.WriteString("│")
	}
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
