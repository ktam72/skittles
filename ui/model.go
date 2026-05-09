package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ktam/skittles/actions"
	"github.com/ktam/skittles/fs"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

type Mode int

const (
	ModeBrowse Mode = iota
	ModeView
	ModeQuit
)

type Model struct {
	Left       *Pane
	Right      *Pane
	Console    *Console
	Focus      int
	Width      int
	Height     int
	now        time.Time
	mode       Mode
	reg        *actions.Registry
	viewer     string
	viewerBuf  []string
	viewerOff  int
	err        error
}

const (
	focusLeft = iota
	focusRight
	focusConsole
)

const consoleHeight = 8

func NewModel(leftDir, rightDir string) *Model {
	w, h := 120, 40
	paneW := w/2 - 2
	paneH := h - consoleHeight - 9

	if leftDir == "" {
		leftDir, _ = os.Getwd()
	}
	if rightDir == "" {
		rightDir, _ = os.Getwd()
	}

	m := &Model{
		Left:    NewPane(leftDir, paneW, paneH),
		Right:   NewPane(rightDir, paneW, paneH),
		Console: NewConsole(w-2, consoleHeight),
		Focus:   focusLeft,
		Width:   w,
		Height:  h,
		now:     time.Now(),
		mode:    ModeBrowse,
		reg:     actions.DefaultRegistry(),
	}
	m.Left.Active = true
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}

func (m *Model) FocusedPane() *Pane {
	switch m.Focus {
	case focusLeft:
		return m.Left
	case focusRight:
		return m.Right
	default:
		return nil
	}
}

func (m *Model) OppPane() *Pane {
	switch m.Focus {
	case focusLeft:
		return m.Right
	case focusRight:
		return m.Left
	default:
		return m.Left
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case time.Time:
		m.now = msg
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return t
		})

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		paneW := msg.Width/2 - 2
		paneH := msg.Height - consoleHeight - 9
		m.Left.Width = paneW
		m.Left.Height = paneH
		m.Right.Width = paneW
		m.Right.Height = paneH
		m.Console.Width = msg.Width - 2
		m.Console.Height = consoleHeight

	case tea.KeyMsg:
		switch m.mode {
		case ModeQuit:
			if msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}
			m.mode = ModeBrowse

		case ModeView:
			return m.handleViewMode(msg)

		default:
			if m.Focus == focusConsole {
				return m.handleConsoleMode(msg)
			}
			return m.handleBrowseMode(msg)
		}
	}
	return m, nil
}

func (m *Model) handleViewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = ModeBrowse
		m.viewer = ""
		m.viewerBuf = nil
		m.viewerOff = 0
		return m, tea.ClearScreen
	case tea.KeyUp:
		if m.viewerOff > 0 {
			m.viewerOff--
		}
	case tea.KeyDown:
		maxOff := len(m.viewerBuf) - m.Height + 2
		if maxOff < 0 {
			maxOff = 0
		}
		if m.viewerOff < maxOff {
			m.viewerOff++
		}
	case tea.KeyPgUp:
		m.viewerOff -= m.Height - 2
		if m.viewerOff < 0 {
			m.viewerOff = 0
		}
	case tea.KeyPgDown:
		m.viewerOff += m.Height - 2
		maxOff := len(m.viewerBuf) - m.Height + 2
		if maxOff < 0 {
			maxOff = 0
		}
		if m.viewerOff > maxOff {
			m.viewerOff = maxOff
		}
	}
	return m, nil
}

func (m *Model) loadViewerBuffer() {
	data, err := os.ReadFile(m.viewer)
	if err != nil {
		m.viewerBuf = []string{fmt.Sprintf("Error: %v", err)}
		return
	}

	text := string(data)
	ext := strings.ToLower(filepath.Ext(m.viewer))

	if ext == ".md" {
		rendered, err := glamour.Render(text, "dark")
		if err == nil {
			m.viewerBuf = strings.Split(rendered, "\n")
			m.viewerOff = 0
			return
		}
	}

	var buf bytes.Buffer
	lang := ""
	lexer := lexers.Match(filepath.Base(m.viewer))
	if lexer != nil {
		lang = lexer.Config().Name
	}
	err = quick.Highlight(&buf, text, lang, "terminal", "dracula")
	if err == nil && buf.Len() > 0 {
		m.viewerBuf = strings.Split(buf.String(), "\n")
		m.viewerOff = 0
		return
	}

	m.viewerBuf = strings.Split(text, "\n")
	m.viewerOff = 0
}

func (m *Model) handleConsoleMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.Focus = focusLeft
		m.Left.Active = true
		m.Console.Active = false
		m.Console.ClearInput()

	case tea.KeyEnter:
		cmd := strings.TrimSpace(string(m.Console.Input))
		m.Console.ClearInput()
		if cmd != "" {
			m.Console.Exec(cmd)
		}

	case tea.KeyBackspace:
		m.Console.DeleteRune()

	case tea.KeySpace:
		m.Console.InsertRune(' ')

	case tea.KeyUp:
		if m.Console.Output != nil && m.Console.Input == nil {
			m.Console.Scroll--
			if m.Console.Scroll < 0 {
				m.Console.Scroll = 0
			}
		} else {
			m.Console.PrevHistory()
		}

	case tea.KeyDown:
		if m.Console.Input == nil {
			m.Console.Scroll++
			if m.Console.Scroll > len(m.Console.Output) {
				m.Console.Scroll = len(m.Console.Output)
			}
		} else {
			m.Console.NextHistory()
		}

	case tea.KeyTab:
		m.cycleFocus()

	default:
		if !msg.Alt && msg.String() != "" && len(msg.String()) == 1 {
			m.Console.InsertRune(rune(msg.String()[0]))
		}
	}
	return m, nil
}

func (m *Model) cycleFocus() {
	m.Left.Active = false
	m.Right.Active = false
	m.Console.Active = false

	m.Focus = (m.Focus + 1) % 3
	switch m.Focus {
	case focusLeft:
		m.Left.Active = true
	case focusRight:
		m.Right.Active = true
	case focusConsole:
		m.Console.Active = true
	}
}

func (m *Model) handleBrowseMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := m.FocusedPane()
	opp := m.OppPane()

	switch msg.String() {
	case "esc":
		m.mode = ModeQuit
		return m, nil

	case "up", "k":
		p.Up(1)
	case "down", "j":
		p.Down(1)
	case "pgup", "b":
		p.PageUp()
	case "pgdown", " ":
		p.PageDown()
	case "home", "g":
		p.Home()
	case "end", "G":
		p.End()

	case "right", "l":
		cur := p.Current()
		if cur == nil {
			return m, nil
		}
		if cur.IsDir {
			abs, _ := filepath.Abs(cur.Path)
			_ = p.Chdir(abs)
		} else {
			act := m.reg.Resolve(cur.Path)
			if act.Look {
				m.viewer = cur.Path
				m.loadViewerBuffer()
				m.mode = ModeView
			} else if act.Command != "" {
				out := m.runAction(act, cur.Path)
				if out != "" {
					m.Console.AddOutput(out)
				}
			} else if isExecutable(cur) {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vim"
				}
				m.runAction(actions.Action{Command: fmt.Sprintf("%s $P", editor)}, cur.Path)
			} else {
				m.viewer = cur.Path
				m.loadViewerBuffer()
				m.mode = ModeView
			}
		}

	case "enter":
		cur := p.Current()
		if cur == nil {
			return m, nil
		}
		if cur.IsDir {
			abs, _ := filepath.Abs(cur.Path)
			_ = p.Chdir(abs)
		} else {
			act := m.reg.Resolve(cur.Path)
			if act.Look {
				m.viewer = cur.Path
				m.loadViewerBuffer()
				m.mode = ModeView
			} else if act.Command != "" {
				out := m.runAction(act, cur.Path)
				if out != "" {
					m.Console.AddOutput(out)
				}
			} else if isExecutable(cur) {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vim"
				}
				m.runAction(actions.Action{Command: fmt.Sprintf("%s $P", editor)}, cur.Path)
			} else {
				m.viewer = cur.Path
				m.loadViewerBuffer()
				m.mode = ModeView
			}
		}

	case "left", "h":
		if p.Dir != "/" {
			parent := filepath.Dir(strings.TrimRight(p.Dir, "/"))
			_ = p.Chdir(parent)
		}

	case "tab":
		m.cycleFocus()

	case "backspace", "bs":
		p.ToggleMark()
		p.Down(1)

	case "c":
		entries := p.SelectedEntries()
		if len(entries) > 0 {
			_, err := fs.CopyEntries(p.Dir, entries, opp.Dir)
			if err != nil {
				m.err = err
			} else {
				m.err = nil
			}
			p.ClearMarks()
			p.Reload()
			opp.Reload()
		}

	case "m":
		entries := p.SelectedEntries()
		if len(entries) > 0 {
			_, err := fs.MoveEntries(p.Dir, entries, opp.Dir)
			if err != nil {
				m.err = err
			}
			p.ClearMarks()
			p.Reload()
			opp.Reload()
		}

	case "d":
		entries := p.SelectedEntries()
		if len(entries) > 0 {
			for _, e := range entries {
				_ = fs.Delete(e.Path)
			}
			p.ClearMarks()
			p.Reload()
		}

	case "a":
		p.MarkAll()
	case "r":
		p.Reload()

	case "sr":
		cycle := []fs.SortMode{fs.SortName, fs.SortTime, fs.SortExt, fs.SortSize}
		idx := 0
		for i, s := range cycle {
			if p.SortBy == s {
				idx = (i + 1) % len(cycle)
				break
			}
		}
		p.SortBy = cycle[idx]
		p.Reload()

	case "E":
		cur := p.Current()
		if cur != nil {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}
			m.runAction(actions.Action{Command: fmt.Sprintf("%s $P", editor)}, cur.Path)
		}

	case "!":
		m.Focus = focusConsole
		m.Left.Active = false
		m.Right.Active = false
		m.Console.Active = true
	}

	return m, nil
}

func (m *Model) runAction(act actions.Action, path string) string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmdStr := act.Command
	cmdStr = strings.ReplaceAll(cmdStr, "$P", path)
	cmdStr = strings.ReplaceAll(cmdStr, "$F", filepath.Base(path))
	cmdStr = strings.ReplaceAll(cmdStr, "$D", filepath.Dir(path))
	cmdStr = strings.ReplaceAll(cmdStr, "$EDITOR", editor)

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return ""
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("error: %v\n%s", err, string(out))
	}
	return string(out)
}

func isExecutable(e *fs.Entry) bool {
	return e.Mode&0100 != 0
}
