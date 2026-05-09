package ui

import (
	"fmt"
	"strings"

	"github.com/ktam/skittles/fs"
)

type RowInfo struct {
	Text     string
	IsDir    bool
	IsLink   bool
	IsMarked bool
	IsCursor bool
}

type Pane struct {
	Dir     string
	Listing *fs.Listing
	Cursor  int
	Offset  int
	Active  bool
	Width   int
	Height  int
	Marked  map[string]bool
	SortBy  fs.SortMode
	ShowDot bool
}

func NewPane(dir string, width, height int) *Pane {
	p := &Pane{
		Dir:    dir,
		Width:  width,
		Height: height,
		Marked: make(map[string]bool),
	}
	p.Reload()
	return p
}

func (p *Pane) Reload() {
	l, err := fs.ReadDir(p.Dir)
	if err != nil {
		return
	}
	if !p.ShowDot {
		filtered := make([]fs.Entry, 0, len(l.Entries))
		for _, e := range l.Entries {
			if !strings.HasPrefix(e.Name, ".") {
				filtered = append(filtered, e)
			}
		}
		l.Entries = filtered
	}
	l.SortBy = p.SortBy
	l.Sort()
	p.Listing = l
	if p.Cursor >= len(l.Entries) {
		p.Cursor = len(l.Entries) - 1
	}
	if p.Cursor < 0 {
		p.Cursor = 0
	}
}

func (p *Pane) Current() *fs.Entry {
	if p.Listing == nil || p.Cursor >= len(p.Listing.Entries) {
		return nil
	}
	return &p.Listing.Entries[p.Cursor]
}

func (p *Pane) Up(n int) {
	p.Cursor -= n
	if p.Cursor < 0 {
		p.Cursor = 0
	}
	p.ensureVisible()
}

func (p *Pane) Down(n int) {
	if p.Listing == nil {
		return
	}
	max := len(p.Listing.Entries) - 1
	p.Cursor += n
	if p.Cursor > max {
		p.Cursor = max
	}
	p.ensureVisible()
}

func (p *Pane) PageUp() {
	p.Up(p.Height - 1)
}

func (p *Pane) PageDown() {
	p.Down(p.Height - 1)
}

func (p *Pane) Home() {
	p.Cursor = 0
	p.Offset = 0
}

func (p *Pane) End() {
	if p.Listing == nil {
		return
	}
	p.Cursor = len(p.Listing.Entries) - 1
	p.ensureVisible()
}

func (p *Pane) ensureVisible() {
	if p.Cursor < p.Offset {
		p.Offset = p.Cursor
	}
	if p.Cursor >= p.Offset+p.Height {
		p.Offset = p.Cursor - p.Height + 1
	}
}

func (p *Pane) Chdir(dir string) error {
	p.Dir = dir
	p.Cursor = 0
	p.Offset = 0
	p.Reload()
	return nil
}

func (p *Pane) ToggleMark() {
	if p.Listing == nil || p.Cursor >= len(p.Listing.Entries) {
		return
	}
	name := p.Listing.Entries[p.Cursor].Name
	p.Marked[name] = !p.Marked[name]
}

func (p *Pane) MarkAll() {
	if p.Listing == nil {
		return
	}
	for i := range p.Listing.Entries {
		p.Marked[p.Listing.Entries[i].Name] = true
	}
}

func (p *Pane) ClearMarks() {
	p.Marked = make(map[string]bool)
}

func (p *Pane) MarkedEntries() []fs.Entry {
	if p.Listing == nil {
		return nil
	}
	var entries []fs.Entry
	for i := range p.Listing.Entries {
		if p.Marked[p.Listing.Entries[i].Name] {
			entries = append(entries, p.Listing.Entries[i])
		}
	}
	return entries
}

func (p *Pane) SelectedEntries() []fs.Entry {
	if p.Listing == nil || len(p.Listing.Entries) == 0 {
		return nil
	}
	if len(p.Marked) > 0 {
		return p.MarkedEntries()
	}
	return p.Listing.Entries[p.Cursor : p.Cursor+1]
}

func (p *Pane) RenderHeader() string {
	dir := p.Dir
	if len(dir) > p.Width-2 {
		dir = "..." + dir[len(dir)-(p.Width-5):]
	}
	return fmt.Sprintf(" %-*s ", p.Width-1, dir)
}

func (p *Pane) RenderRows() []RowInfo {
	if p.Listing == nil {
		return nil
	}
	rows := make([]RowInfo, 0, p.Height)
	entries := p.Listing.Entries
	for y := 0; y < p.Height; y++ {
		idx := p.Offset + y
		if idx >= len(entries) {
			rows = append(rows, RowInfo{Text: strings.Repeat(" ", p.Width)})
			continue
		}
		e := &entries[idx]
		rows = append(rows, RowInfo{
			Text:     p.formatEntryLine(e),
			IsDir:    e.IsDir,
			IsLink:   e.IsLink,
			IsMarked: p.Marked[e.Name],
			IsCursor: idx == p.Cursor,
		})
	}
	return rows
}

func (p *Pane) formatEntryLine(e *fs.Entry) string {
	icon := " "
	if e.IsDir {
		icon = "📁"
	}
	if e.IsLink {
		icon = "🔗"
	}
	size := ""
	if !e.IsDir {
		size = formatSize(e.Size)
	}
	name := e.Name
	if len(name) > p.Width-12 {
		name = name[:p.Width-13] + "…"
	}
	line := fmt.Sprintf(" %s %-*s %8s", icon, p.Width-14, name, size)
	if len(line) > p.Width {
		line = line[:p.Width]
	}
	return line
}

func formatSize(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%dB", n)
	}
}
