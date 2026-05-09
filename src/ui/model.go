package ui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ktam72/skittles/src/actions"
	"github.com/ktam72/skittles/src/fs"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type editCmd struct{ *exec.Cmd }

func (c *editCmd) SetStdin(r io.Reader)  { c.Stdin = r }
func (c *editCmd) SetStdout(w io.Writer) { c.Stdout = w }
func (c *editCmd) SetStderr(w io.Writer) { c.Stderr = w }

type archiveExtractedMsg struct {
	tmp  string
	err  error
	path string
}

func extractArchiveCmd(src string) tea.Cmd {
	return func() tea.Msg {
		tmp, err := fs.ExtractToTemp(src)
		return archiveExtractedMsg{tmp: tmp, err: err, path: src}
	}
}

const version = "1.2.0"

type Mode int

const (
	ModeBrowse Mode = iota
	ModeView
	ModeQuit
	ModeConfirm
	ModeRename
)

type Model struct {
	Left      *Pane
	Right     *Pane
	Console   *Console
	Focus     int
	Width     int
	Height    int
	now       time.Time
	mode      Mode
	reg       *actions.Registry
	editor    string
	viewer    string
	viewerBuf []string
	viewerOff int
	err       error
	cursorOn  bool

	confirmAction   func()
	confirmMessage  string
	noDeleteConfirm bool

	renamePath  string
	renameInput []rune
}

const (
	focusLeft = iota
	focusRight
	focusConsole
)

const consoleHeight = 10

func NewModel(leftDir, rightDir string, editor string) *Model {
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
		editor:  editor,
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
		m.cursorOn = !m.cursorOn
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

	case consoleOutputMsg:
		m.Console.AddOutput(string(msg))
		return m, consoleOutputCmd(m.Console.outputCh)

	case archiveExtractedMsg:
		p := m.FocusedPane()
		if p == nil {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err
			m.Console.AddOutput(fmt.Sprintf("error: %v", msg.err))
		} else {
			m.Console.AddOutput(fmt.Sprintf("done: %s", msg.path))
			p.SavedCursor = p.Cursor
			p.ArchivePath = msg.path
			p.RealDir = p.Dir
			p.IsArchive = true
			p.ArchiveRoot = msg.tmp
			_ = p.Chdir(msg.tmp)
		}

	case tea.KeyMsg:
		switch m.mode {
		case ModeQuit:
			if msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}
			m.mode = ModeBrowse

		case ModeView:
			return m.handleViewMode(msg)

		case ModeConfirm:
			return m.handleConfirmMode(msg)

		case ModeRename:
			return m.handleRenameMode(msg)

		default:
			if m.Focus == focusConsole {
				return m.handleConsoleMode(msg)
			}
			return m.handleBrowseMode(msg)
		}
	}
	return m, nil
}

func (m *Model) handleConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.noDeleteConfirm = false
		m.mode = ModeBrowse
		if m.confirmAction != nil {
			m.confirmAction()
		}
	case "2":
		m.noDeleteConfirm = true
		m.mode = ModeBrowse
		if m.confirmAction != nil {
			m.confirmAction()
		}
	case "3", "esc":
		m.mode = ModeBrowse
	}
	return m, nil
}

func (m *Model) handleRenameMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = ModeBrowse
		m.renameInput = nil
	case tea.KeyEnter:
		newName := strings.TrimSpace(string(m.renameInput))
		if newName != "" {
			dir := filepath.Dir(m.renamePath)
			newPath := filepath.Join(dir, newName)
			if err := os.Rename(m.renamePath, newPath); err != nil {
				m.err = err
			}
			m.FocusedPane().Reload()
		}
		m.mode = ModeBrowse
		m.renameInput = nil
	case tea.KeyBackspace:
		if len(m.renameInput) > 0 {
			m.renameInput = m.renameInput[:len(m.renameInput)-1]
		}
	default:
		if !msg.Alt && len(msg.String()) == 1 {
			m.renameInput = append(m.renameInput, []rune(msg.String())...)
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
	case tea.KeyLeft:
		m.viewerOff -= m.Height - 2
		if m.viewerOff < 0 {
			m.viewerOff = 0
		}
	case tea.KeyRight:
		m.viewerOff += m.Height - 2
		maxOff := len(m.viewerBuf) - m.Height + 2
		if maxOff < 0 {
			maxOff = 0
		}
		if m.viewerOff > maxOff {
			m.viewerOff = maxOff
		}
	case tea.KeyRunes:
		if string(msg.Runes) == "e" || string(msg.Runes) == "E" {
			switchToEnglishInput()
			cmd := &editCmd{Cmd: exec.Command(m.editor, m.viewer)}
			return m, tea.Exec(cmd, func(err error) tea.Msg {
				m.loadViewerBuffer()
				return nil
			})
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

	text := fs.DecodeText(data)
	ext := strings.ToLower(filepath.Ext(m.viewer))

	if ext == ".md" {
		m.viewerBuf = strings.Split(renderMarkdown(text), "\n")
		m.viewerOff = 0
		return
	}

	if isBinaryData(data) {
		m.viewerBuf = renderHexView(data)
		m.viewerOff = 0
		return
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
			return m, consoleOutputCmd(m.Console.outputCh)
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

	case tea.KeyPgUp:
		page := m.Console.Height - 3
		if page < 1 {
			page = 1
		}
		m.Console.Scroll -= page
		if m.Console.Scroll < 0 {
			m.Console.Scroll = 0
		}

	case tea.KeyPgDown:
		page := m.Console.Height - 3
		if page < 1 {
			page = 1
		}
		m.Console.Scroll += page
		if m.Console.Scroll > len(m.Console.Output) {
			m.Console.Scroll = len(m.Console.Output)
		}

	case tea.KeyTab:
		m.cycleFocus()

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "y":
			if len(m.Console.Input) == 0 {
				m.copyConsoleLines(false)
			} else {
				m.Console.InsertRune('y')
			}
		case "Y":
			if len(m.Console.Input) == 0 {
				m.copyConsoleLines(true)
			} else {
				m.Console.InsertRune('Y')
			}
		default:
			for _, r := range msg.Runes {
				m.Console.InsertRune(r)
			}
		}

	default:
		if !msg.Alt && msg.String() != "" && len(msg.String()) == 1 {
			m.Console.InsertRune(rune(msg.String()[0]))
		}
	}
	return m, nil
}

func (m *Model) copyConsoleLines(all bool) {
	var lines []string
	if all {
		lines = m.Console.Output
	} else {
		start := m.Console.Scroll - (m.Console.Height - 3)
		if start < 0 {
			start = 0
		}
		end := m.Console.Scroll
		if end > len(m.Console.Output) {
			end = len(m.Console.Output)
		}
		lines = m.Console.Output[start:end]
	}
	text := strings.Join(lines, "\n")
	if err := clipboard.WriteAll(text); err != nil {
		m.Console.AddOutput(fmt.Sprintf("clipboard error: %v", err))
	} else {
		m.Console.AddOutput(fmt.Sprintf("copied %d lines", len(lines)))
	}
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
		switchToEnglishInput()
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
	case "pgdown":
		p.PageDown()
	case "home", "g":
		p.Home()
	case "end", "G":
		p.End()

	case "enter":
		cur := p.Current()
		if cur == nil {
			return m, nil
		}
		if cur.IsDir {
			if p.IsArchive && cur.Name == ".." && p.Dir == p.ArchiveRoot {
				_ = os.RemoveAll(p.Dir)
				p.IsArchive = false
				p.ArchivePath = ""
				p.ArchiveRoot = ""
				_ = p.Chdir(p.RealDir)
				p.Cursor = p.SavedCursor
				p.RealDir = ""
			} else {
				abs, _ := filepath.Abs(cur.Path)
				_ = p.Chdir(abs)
			}
		} else {
			act := m.reg.Resolve(cur.Path)
			if act.Browse {
				m.Console.AddOutput(fmt.Sprintf("Extracting %s ...", cur.Name))
				p.SavedCursor = p.Cursor
				p.ArchivePath = cur.Path
				p.RealDir = p.Dir
				return m, extractArchiveCmd(cur.Path)
			} else if act.Look {
				m.viewer = cur.Path
				m.loadViewerBuffer()
				m.mode = ModeView
			} else if act.Command != "" {
				out := m.runAction(act, cur.Path)
				if out != "" {
					for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
						m.Console.AddOutput(line)
					}
				}
			} else if isExecutable(cur) {
				m.Console.ClearInput()
				m.Console.Input = []rune(cur.Path)
				m.Console.Cursor = len(cur.Path)
				m.Focus = focusConsole
				m.Left.Active = false
				m.Right.Active = false
				m.Console.Active = true
				switchToEnglishInput()
			} else {
				m.viewer = cur.Path
				m.loadViewerBuffer()
				m.mode = ModeView
			}
		}

	case "left", "h":
		if p.Dir != "/" {
			if p.IsArchive && p.Dir == p.ArchiveRoot {
				_ = os.RemoveAll(p.Dir)
				p.IsArchive = false
				p.ArchivePath = ""
				p.ArchiveRoot = ""
				_ = p.Chdir(p.RealDir)
				p.Cursor = p.SavedCursor
				p.RealDir = ""
			} else if p.IsArchive {
				parent := filepath.Dir(strings.TrimRight(p.Dir, "/"))
				_ = p.Chdir(parent)
			} else {
				parent := filepath.Dir(strings.TrimRight(p.Dir, "/"))
				_ = p.Chdir(parent)
			}
		}

	case "tab":
		m.cycleFocus()

	case " ":
		p.ToggleMark()
		p.Down(1)
	case "backspace", "bs":
		if p.Dir != "/" {
			if p.IsArchive && p.Dir == p.ArchiveRoot {
				_ = os.RemoveAll(p.Dir)
				p.IsArchive = false
				p.ArchivePath = ""
				p.ArchiveRoot = ""
				_ = p.Chdir(p.RealDir)
				p.Cursor = p.SavedCursor
				p.RealDir = ""
			} else if p.IsArchive {
				parent := filepath.Dir(strings.TrimRight(p.Dir, "/"))
				_ = p.Chdir(parent)
			} else {
				parent := filepath.Dir(strings.TrimRight(p.Dir, "/"))
				_ = p.Chdir(parent)
			}
		}

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
			if m.noDeleteConfirm {
				for _, e := range entries {
					_ = fs.Delete(e.Path)
				}
				p.ClearMarks()
				p.Reload()
			} else {
				names := make([]string, len(entries))
				for i, e := range entries {
					names[i] = e.Name
				}
				msg := fmt.Sprintf("Delete %d file(s)?\n  %s", len(entries), strings.Join(names, "\n  "))
				m.confirmMessage = msg
				m.confirmAction = func() {
					for _, e := range entries {
						_ = fs.Delete(e.Path)
					}
					p.ClearMarks()
					p.Reload()
				}
				m.mode = ModeConfirm
			}
		}

	case "a":
		p.MarkAll()
	case "p":
		cur := p.Current()
		if cur != nil && !cur.IsDir {
			_ = openFile(cur.Path)
		}
	case "x":
		cur := p.Current()
		if cur != nil && !cur.IsDir {
			cmd := fs.ExtractCmdFor(cur.Path)
			if cmd != "" {
				out := m.runAction(actions.Action{Command: fmt.Sprintf("%s $P", cmd)}, cur.Path)
				if out != "" {
					for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
						m.Console.AddOutput(line)
					}
				}
			}
		}
	case "r":
		cur := p.Current()
		if cur != nil {
			m.renamePath = cur.Path
			m.renameInput = []rune(cur.Name)
			m.mode = ModeRename
		}
	case "R":
		p.Reload()
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
			switchToEnglishInput()
			m.runAction(actions.Action{Command: fmt.Sprintf("%s $P", m.editor)}, cur.Path)
		}

	case "!":
		m.Focus = focusConsole
		m.Left.Active = false
		m.Right.Active = false
		m.Console.Active = true
		switchToEnglishInput()
	}

	return m, nil
}

func (m *Model) runAction(act actions.Action, path string) string {
	editor := m.editor
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

func isBinaryData(data []byte) bool {
	n := len(data)
	if n > 512 {
		n = 512
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

func renderHexView(data []byte) []string {
	return fs.RenderHexView(data)
}

func consoleOutputCmd(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return consoleOutputMsg(line)
	}
}
