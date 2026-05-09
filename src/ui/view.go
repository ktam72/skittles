package ui

import (
	"fmt"
	"strings"

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

func (m *Model) renderWithConfirm(topBar string) string {
	// dialog
	opts := []string{
		" 1. OK（毎回確認する）",
		" 2. OK（以降確認しない）",
		" 3. キャンセル",
	}
	optStr := strings.Join(opts, "\n")
	content := fmt.Sprintf("\n%s\n\n%s\n", m.confirmMessage, optStr)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, 2).
		Width(40).
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
	left := fmt.Sprintf(" Skittles v%s ", version)
	right := m.now.Format(" 2006/01/02 15:04:05 ")
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

	body := m.Console.RenderBody()
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

func (m *Model) renderBindings() string {
	hints := []string{
		"↑↓:nav", "Enter:open", "Tab:focus(3-pane)",
		"Space:mark", "c:copy", "m:move", "d:del", "p:preview",
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

	info := fmt.Sprintf("─── %s ─── [ESC:close ↑↓:scroll  L%d/%d]",
		m.viewer, m.viewerOff+1, len(m.viewerBuf))
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Render(info)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		style.Render(content))
}
