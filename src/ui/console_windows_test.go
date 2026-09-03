//go:build windows

package ui

import (
	"strings"
	"testing"
)

func TestMapUnixCmd(t *testing.T) {
	cases := []struct {
		name    string
		parts   []string
		cmdLine string
		want    string
	}{
		{"ls empty", []string{"ls"}, "ls", "dir"},
		{"ls flags stripped", []string{"ls", "-la"}, "ls -la", "dir"},
		{"ls with path", []string{"ls", "subdir"}, "ls subdir", "dir subdir"},
		{"ls path with space", []string{"ls", "my dir"}, "ls my dir", `dir "my dir"`},
		{"dir passthrough kept", []string{"dir", "/q"}, "dir /q", "dir /q"},
		{"cat", []string{"cat", "f.txt"}, "cat f.txt", "type f.txt"},
		{"grep flag convert", []string{"grep", "-ri", "foo"}, "grep -ri foo", "findstr /s /i foo"},
		{"grep keep pattern", []string{"grep", "pattern", "a.txt"}, "grep pattern a.txt", "findstr pattern a.txt"},
		{"rm", []string{"rm", "a.txt"}, "rm a.txt", "del a.txt"},
		{"cp", []string{"cp", "a.txt", "b.txt"}, "cp a.txt b.txt", "copy a.txt b.txt"},
		{"mv", []string{"mv", "a.txt"}, "mv a.txt", "move a.txt"},
		{"head", []string{"head", "a.txt"}, "head a.txt", "more a.txt"},
		{"touch single", []string{"touch", "new.txt"}, "touch new.txt", "type nul > new.txt"},
		{"touch multi", []string{"touch", "a.txt", "b.txt"}, "touch a.txt b.txt", `type nul > a.txt & type nul > b.txt`},
		{"unknown passthrough", []string{"dri"}, "dri", "dri"},
		{"case insensitive", []string{"LS"}, "LS", "dir"},
		{"no parts", []string{}, "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := mapUnixCmd(c.parts, c.cmdLine); got != c.want {
				t.Errorf("mapUnixCmd(%v,%q) = %q, want %q", c.parts, c.cmdLine, got, c.want)
			}
		})
	}
}

func TestDecodeCP932(t *testing.T) {
	dec := decodeForCP(932)
	// 日本語 in Shift_JIS (CP932) - verified against real cmd.exe output
	got := decode(dec, []byte{0x93, 0xFA, 0x96, 0x7B, 0x8C, 0xEA})
	if got != "日本語" {
		t.Errorf("decode CP932 = %q, want 日本語", got)
	}
}

func TestDecodeUTF8Passthrough(t *testing.T) {
	dec := decodeForCP(65001)
	got := decode(dec, []byte{0xE6, 0x97, 0xA5, 0x61})
	if got != "日a" {
		t.Errorf("decode 65001 = %q, want 日a", got)
	}
}

// runAndCollect は runExec を直列で実行し、outputCh から全行を取り出す。
func runAndCollect(parts []string, cmdLine, dir string) []string {
	c := &Console{}
	ch := make(chan string, 1024)
	c.outputCh = ch
	go func() {
		defer close(ch)
		runExec(parts, cmdLine, dir, c, ch)
	}()
	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}
	return lines
}

func TestRunExecDir(t *testing.T) {
	lines := runAndCollect([]string{"dir"}, "dir", t.TempDir())
	if len(lines) == 0 {
		t.Fatal("dir produced no output")
	}
}

func TestRunExecLsMapsToDir(t *testing.T) {
	// ls は dir にマッピングされ、ネイティブ出力を返す
	lines := runAndCollect([]string{"ls"}, "ls", t.TempDir())
	if len(lines) == 0 {
		t.Fatal("ls produced no output")
	}
}

func TestRunExecUnknownPassthrough(t *testing.T) {
	// dri(打間違い) は passthrough で cmd.exe のエラーを返す
	lines := runAndCollect([]string{"dri"}, "dri", t.TempDir())
	if len(lines) == 0 {
		t.Fatal("dri produced no output")
	}
}

func TestRunExecEchoJapanese(t *testing.T) {
	// CP932 出力が正しく UTF-8 復号される
	lines := runAndCollect([]string{"echo"}, "echo 日本語テスト", t.TempDir())
	found := false
	for _, l := range lines {
		if strings.Contains(l, "日本語") {
			found = true
		}
	}
	if !found {
		t.Errorf("Japanese echo not decoded correctly; lines=%v", lines)
	}
}
