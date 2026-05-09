package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ktam72/skittles/src/fs"
	"github.com/charmbracelet/lipgloss"
)

var (
	topBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("17")).
			Bold(true)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Bold(false)

	activeArchivePaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205")).
				Bold(false)

	inactivePaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))

	activeConsoleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63"))

	inactiveConsoleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Bold(true)

	archiveHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("205")).
				Bold(true)

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81"))

	archiveDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("219"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("253"))

	archiveFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("218"))

	markedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("52"))

	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("24"))

	cursorArchiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("53"))

	cursorDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Background(lipgloss.Color("24"))

	cursorArchiveDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("219")).
			Background(lipgloss.Color("53"))

	cursorMarkedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("52")).
				Bold(true)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))
)

func (m *Model) View() string {
	if m.mode == ModeQuit {
		return lipgloss.NewStyle().Padding(2, 4).Render("Press ESC again to quit. Any other key to cancel.")
	}
	if m.mode == ModeView {
		return m.renderViewer()
	}

	topBar := m.renderTopBar()

	if m.mode == ModeConfirm {
		return m.renderWithConfirm(topBar)
	}
	if m.mode == ModeRename {
		return m.renderWithRename(topBar)
	}
	if m.mode == ModeFilter {
		return m.renderWithFilter(topBar)
	}
	if m.mode == ModeChmod {
		return m.renderWithChmod(topBar)
	}
	if m.mode == ModeFileSearch {
		return m.renderWithFileSearch(topBar)
	}
	if m.mode == ModeGrep {
		return m.renderWithGrep(topBar)
	}
	if m.mode == ModePathHistory {
		return m.renderWithPathHistory(topBar)
	}
	if m.mode == ModeCompare {
		return m.renderCompareView()
	}

	return m.renderBrowseBody(topBar)
}

func (m *Model) renderBrowseBody(topBar string) string {
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	console := m.renderConsole()
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))

	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, sep, console, bindings)

	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body,
			lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.err.Error()))
		m.err = nil
	}

	return body
}

func (m *Model) renderWithRename(topBar string) string {
	base := filepath.Base(m.renamePath)
	cursor := " "
	if m.cursorOn {
		cursor = "█"
	}
	current := string(m.renameInput) + cursor
	content := fmt.Sprintf("Rename:\n  %s\n\nto:\n  %s\n", base, current)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(50).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3

	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderWithFilter(topBar string) string {
	cursor := " "
	if m.cursorOn {
		cursor = "█"
	}
	current := string(m.filterInput) + cursor
	prompt := "Filter pattern (glob, empty=clear):"
	if m.filterInput == nil {
		current = cursor
	}
	content := fmt.Sprintf("\n%s\n\n  %s\n", prompt, current)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(50).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3

	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderWithChmod(topBar string) string {
	cursor := " "
	if m.cursorOn {
		cursor = "█"
	}
	current := string(m.chmodInput) + cursor
	content := fmt.Sprintf("\nchmod:\n  %s\n\nMode (octal):\n  %s\n", m.chmodPath, current)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(50).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3

	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderWithFileSearch(topBar string) string {
	cursor := " "
	if m.cursorOn {
		cursor = "█"
	}
	current := string(m.fileSearchPattern) + cursor
	if m.fileSearchPattern == nil {
		current = cursor
	}
	content := fmt.Sprintf("\nFile Search (glob pattern):\n\n  %s\n\nEnter=search  ESC=cancel\n", current)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(50).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3

	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderWithGrep(topBar string) string {
	cursor := " "
	if m.cursorOn {
		cursor = "█"
	}
	current := string(m.grepPattern) + cursor
	if m.grepPattern == nil {
		current = cursor
	}
	content := fmt.Sprintf("\nGrep (case-insensitive):\n\n  %s\n\nEnter=search  ESC=cancel\n", current)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(50).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3

	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderWithConfirm(topBar string) string {
	var content string
	if m.confirmAction == nil {
		content = fmt.Sprintf("\n%s\n\n", m.confirmMessage)
	} else {
		opts := []string{
			" 1. OK（毎回確認する）",
			" 2. OK（以降確認しない）",
			" 3. キャンセル",
		}
		optStr := strings.Join(opts, "\n")
		content = fmt.Sprintf("\n%s\n\n%s\n", m.confirmMessage, optStr)
	}
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, 2).
		Width(50).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3 // + border top/bottom extra spacing

	// reduce pane height
	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderTopBar() string {
	left := fmt.Sprintf(" Skittles v%s by ktam72 ", version)
	right := m.now.Format(" 2006/01/02 (Mon) 15:04:05 ")
	padding := m.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}
	bar := left + strings.Repeat(" ", padding) + right
	return topBarStyle.Render(bar)
}

func (m *Model) renderPane(p *Pane, active bool) string {
	style := inactivePaneStyle
	if active && p.IsArchive {
		style = activeArchivePaneStyle
	} else if active {
		style = activePaneStyle
	}

	hStyle := headerStyle
	if p.IsArchive {
		hStyle = archiveHeaderStyle
	}

	header := p.RenderHeader()
	header = hStyle.Render(header)

	var styledRows []string
	for _, row := range p.RenderRows() {
		styledRows = append(styledRows, m.styleRow(row, p.Width))
	}
	content := strings.Join(styledRows, "\n")
	content = style.Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func (m *Model) renderConsole() string {
	style := inactiveConsoleStyle
	if m.Console.Active {
		style = activeConsoleStyle
	}

	header := m.Console.RenderHeader()
	header = headerStyle.Render(header)

	body := m.Console.RenderBody(m.cursorOn)
	body = style.Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m *Model) styleRow(r RowInfo, paneWidth int) string {
	text := r.Text
	if len(text) > paneWidth {
		text = text[:paneWidth]
	}

	switch {
	case r.IsMarked && r.IsCursor:
		return cursorMarkedStyle.Render(text)
	case r.IsMarked:
		return markedStyle.Render(text)
	case r.IsCursor && r.IsDir && r.IsArchive:
		return cursorArchiveDirStyle.Render(text)
	case r.IsCursor && r.IsDir:
		return cursorDirStyle.Render(text)
	case r.IsCursor && r.IsArchive:
		return cursorArchiveStyle.Render(text)
	case r.IsCursor:
		return cursorStyle.Render(text)
	case r.IsDir && r.IsArchive:
		return archiveDirStyle.Render(text)
	case r.IsDir:
		return dirStyle.Render(text)
	case r.IsArchive:
		return archiveFileStyle.Render(text)
	case r.IsLink:
		return linkStyle.Render(text)
	default:
		return fileStyle.Render(text)
	}
}

func (m *Model) renderWithPathHistory(topBar string) string {
	dirs := m.FocusedPane().PathHistory
	if dirs == nil {
		dirs = []string{}
	}
	n := len(dirs)
	if n > 9 {
		n = 9
	}
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf(" %d. %s", i+1, dirs[len(dirs)-n+i])
	}
	content := strings.Join(lines, "\n")
	if content == "" {
		content = " (no history)"
	}
	content = fmt.Sprintf("\nPath History:\n\n%s\n\n[1-9]:jump  ESC:close\n", content)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(60).
		Render(content)
	dialogH := strings.Count(dialog, "\n") + 3

	savedH := m.Left.Height
	m.Left.Height -= dialogH
	m.Right.Height -= dialogH
	if m.Left.Height < 3 {
		m.Left.Height = 3
		m.Right.Height = 3
	}
	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	m.Left.Height = savedH
	m.Right.Height = savedH

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))
	console := m.renderConsole()
	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, topBar, top, dialog, sep, console, bindings)
	return body
}

func (m *Model) renderCompareView() string {
	viewH := m.Height - 3
	if viewH < 1 {
		viewH = 1
	}

	var visible []string
	end := m.viewerOff + viewH
	if end > len(m.compareResult) {
		end = len(m.compareResult)
	}
	stats := map[fs.DiffKind]int{}
	for _, d := range m.compareResult {
		stats[d.Kind]++
	}
	summary := fmt.Sprintf("Same:%d  LeftOnly:%d  RightOnly:%d  Different:%d",
		stats[fs.DiffSame], stats[fs.DiffLeftOnly], stats[fs.DiffRightOnly], stats[fs.DiffDifferent])

	for i := m.viewerOff; i < end; i++ {
		if i >= len(m.compareResult) {
			break
		}
		d := m.compareResult[i]
		switch d.Kind {
		case fs.DiffSame:
			visible = append(visible, lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(fmt.Sprintf(" = %s", d.Name)))
		case fs.DiffLeftOnly:
			visible = append(visible, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf(" L %s", d.Name)))
		case fs.DiffRightOnly:
			visible = append(visible, lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf(" R %s", d.Name)))
		case fs.DiffDifferent:
			visible = append(visible, lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(fmt.Sprintf(" ! %s  (L:%d  R:%d)", d.Name, d.LeftSize, d.RightSize)))
		}
	}
	for len(visible) < viewH {
		visible = append(visible, "")
	}

	content := strings.Join(visible, "\n")
	style := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("15"))

	info := fmt.Sprintf("─── File Compare ─── %s [ESC:close  ↑↓:scroll  L%d/%d]", summary, m.viewerOff+1, len(m.compareResult))
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Render(info)

	return lipgloss.JoinVertical(lipgloss.Left, header, style.Render(content))
}

func (m *Model) renderBindings() string {
	hints := []string{
		"↑↓:nav", "Enter:open", "Tab:focus(3-pane)",
		"Space:mark", "BS:parent", "r:rename", "c:copy", "m:move", "d:del", "p:preview",
		"F:filter", "f:search", "g:grep", "H:history", "C:chmod", "=:compare",
		"ESC×2:quit",
	}
	return hintStyle.Render(strings.Join(hints, "  |  "))
}

func (m *Model) renderViewer() string {
	viewH := m.Height - 3
	if viewH < 1 {
		viewH = 1
	}

	var visible []string
	end := m.viewerOff + viewH
	if end > len(m.viewerBuf) {
		end = len(m.viewerBuf)
	}
	for i := m.viewerOff; i < end; i++ {
		line := m.viewerBuf[i]
		if lipgloss.Width(line) > m.Width-4 {
			line = lipgloss.NewStyle().Width(m.Width - 7).Render(line) + "..."
		}
		visible = append(visible, line)
	}
	for len(visible) < viewH {
		visible = append(visible, "")
	}

	content := strings.Join(visible, "\n")

	style := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("15"))

	searchInfo := ""
	if len(m.searchResults) > 0 {
		searchInfo = fmt.Sprintf("  Match %d/%d", m.searchIdx+1, len(m.searchResults))
	}
	title := m.viewer
	if m.viewerTitle != "" {
		title = m.viewerTitle
	}
	info := fmt.Sprintf("─── %s ─── [ESC:close ↑↓:scroll /:search n/N:next  L%d/%d%s]",
		title, m.viewerOff+1, len(m.viewerBuf), searchInfo)
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Render(info)

	body := lipgloss.JoinVertical(lipgloss.Left, header, style.Render(content))

	if m.searchActive {
		cursor := " "
		if m.cursorOn {
			cursor = "█"
		}
		query := "/" + string(m.searchQuery) + cursor
		searchBar := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("240")).
			Render(fmt.Sprintf(" %-*s", m.Width-4, query))
		body = lipgloss.JoinVertical(lipgloss.Left, body, searchBar)
	}

	return body
}
