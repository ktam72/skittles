package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/ktam/skittles/fs"
)

type RowInfo struct {
	Text      string
	IsDir     bool
	IsLink    bool
	IsMarked  bool
	IsCursor  bool
	IsArchive bool
}

type Pane struct {
	Dir         string
	Listing     *fs.Listing
	Cursor      int
	Offset      int
	Active      bool
	Width       int
	Height      int
	Marked      map[string]bool
	SortBy      fs.SortMode
	ShowDot     bool
	IsArchive   bool
	ArchivePath string
	RealDir     string
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
			if e.Name == ".." || !strings.HasPrefix(e.Name, ".") {
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
			Text:      p.formatEntryLine(e),
			IsDir:     e.IsDir,
			IsLink:    e.IsLink,
			IsMarked:  p.Marked[e.Name],
			IsCursor:  idx == p.Cursor,
			IsArchive: p.IsArchive,
		})
	}
	return rows
}

func (p *Pane) formatEntryLine(e *fs.Entry) string {
	icon := "  "
	if e.IsDir {
		icon = "📁"
	}
	if e.IsLink {
		icon = "🔗"
	}

	perm := formatPerm(e.Mode)
	owner := e.Owner
	if owner == "" {
		owner = "?"
	}
	if len(owner) > 8 {
		owner = owner[:8]
	}
	group := e.Group
	if group == "" {
		group = "?"
	}
	if len(group) > 8 {
		group = group[:8]
	}
	size := ""
	if !e.IsDir {
		size = formatSize(e.Size)
	}

	fixed := len(icon) + 1 + 10 + 1 + 8 + 1 + 8 + 1 + 1 + 8
	nameWidth := p.Width - fixed
	if nameWidth < 3 {
		nameWidth = 3
	}

	name := e.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-1] + "…"
	}

	return fmt.Sprintf("%s %s %-8s %-8s %-*s %8s", icon, perm, owner, group, nameWidth, name, size)
}

func formatPerm(mode os.FileMode) string {
	var b [10]byte
	if mode.IsDir() {
		b[0] = 'd'
	} else if mode&os.ModeSymlink != 0 {
		b[0] = 'l'
	} else {
		b[0] = '-'
	}

	perm := []struct {
		bit  os.FileMode
		char byte
	}{
		{0400, 'r'}, {0200, 'w'}, {0100, 'x'},
		{0040, 'r'}, {0020, 'w'}, {0010, 'x'},
		{0004, 'r'}, {0002, 'w'}, {0001, 'x'},
	}

	for i, p := range perm {
		if mode&p.bit != 0 {
			b[i+1] = p.char
		} else {
			b[i+1] = '-'
		}
	}

	if mode&os.ModeSetuid != 0 {
		if b[3] == 'x' {
			b[3] = 's'
		} else {
			b[3] = 'S'
		}
	}
	if mode&os.ModeSetgid != 0 {
		if b[6] == 'x' {
			b[6] = 's'
		} else {
			b[6] = 'S'
		}
	}
	if mode&os.ModeSticky != 0 {
		if b[9] == 'x' {
			b[9] = 't'
		} else {
			b[9] = 'T'
		}
	}

	return string(b[:])
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
		return fmt.Sprintf("%d", n)
	}
}
