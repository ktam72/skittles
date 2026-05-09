package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
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

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("253"))

	markedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("52"))

	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("24"))

	cursorDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Background(lipgloss.Color("24"))

	cursorMarkedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("52")).
				Bold(true)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236"))

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

	left := m.renderPane(m.Left, m.Focus == focusLeft)
	right := m.renderPane(m.Right, m.Focus == focusRight)
	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	console := m.renderConsole()
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.Width))

	bindings := m.renderBindings()

	body := lipgloss.JoinVertical(lipgloss.Left, top, sep, console, bindings)

	if m.err != nil {
		body = lipgloss.JoinVertical(lipgloss.Left, body,
			lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.err.Error()))
		m.err = nil
	}

	return body
}

func (m *Model) renderPane(p *Pane, active bool) string {
	style := inactivePaneStyle
	if active {
		style = activePaneStyle
	}

	header := p.RenderHeader()
	header = headerStyle.Render(header)

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
	case r.IsCursor && r.IsDir:
		return cursorDirStyle.Render(text)
	case r.IsCursor:
		return cursorStyle.Render(text)
	case r.IsDir:
		return dirStyle.Render(text)
	case r.IsLink:
		return linkStyle.Render(text)
	default:
		return fileStyle.Render(text)
	}
}

func (m *Model) renderBindings() string {
	hints := []string{
		"↑↓:nav", "Enter:open", "Tab:focus(3-pane)",
		"c:copy", "m:move", "d:del", "a:mark all",
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
