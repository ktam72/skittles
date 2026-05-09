package ui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

func renderMarkdown(text string) string {
	rendered, err := glamour.Render(text, "dark")
	if err != nil {
		return renderMarkdownPlain(text)
	}
	return rendered
}

func renderMarkdownPlain(text string) string {
	var out strings.Builder
	sc := bufio.NewScanner(strings.NewReader(text))
	inCode := false
	for sc.Scan() {
		line := sc.Text()

		if strings.HasPrefix(line, "```") {
			inCode = !inCode
			if inCode {
				fmt.Fprintln(&out, "  code:")
			}
			continue
		}

		if inCode {
			fmt.Fprintf(&out, "  %s\n", line)
			continue
		}

		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			fmt.Fprintln(&out)
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			fmt.Fprintf(&out, "### %s\n", trimmed[4:])
		case strings.HasPrefix(trimmed, "## "):
			fmt.Fprintf(&out, "## %s\n", trimmed[3:])
		case strings.HasPrefix(trimmed, "# "):
			fmt.Fprintf(&out, "# %s\n", trimmed[2:])
		case strings.HasPrefix(trimmed, "- "), strings.HasPrefix(trimmed, "* "):
			fmt.Fprintf(&out, "  • %s\n", trimmed[2:])
		case strings.HasPrefix(trimmed, "> "):
			fmt.Fprintf(&out, "  │ %s\n", trimmed[2:])
		case strings.HasPrefix(trimmed, "---"):
			fmt.Fprintln(&out, strings.Repeat("─", 40))
		default:
			fmt.Fprintln(&out, trimBullet(line))
		}
	}
	return out.String()
}

func trimBullet(s string) string {
	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimPrefix(s, "* ")
	s = strings.TrimPrefix(s, "+ ")
	return s
}
