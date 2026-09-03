//go:build windows

package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

func handleBuiltin(parts []string, c *Console) bool {
	switch parts[0] {
	case "sudo":
		c.AddOutput("sudo: 権限昇格は禁止されています")
		return true
	case "clear":
		c.Output = nil
		c.outputBytes = 0
		c.Scroll = 0
		return true
	case "cd":
		return handleCdBuiltin(parts, c)
	default:
		return false
	}
}

func handleCdBuiltin(parts []string, c *Console) bool {
	target := c.Dir
	if len(parts) >= 2 {
		arg := parts[1]
		if arg == "-" {
		} else if strings.HasPrefix(arg, "~") {
			home, _ := os.UserHomeDir()
			target = filepath.Join(home, arg[1:])
		} else if strings.Contains(arg, ":") {
			target = filepath.Clean(arg)
		} else {
			target = filepath.Join(c.Dir, arg)
		}
	} else {
		home, _ := os.UserHomeDir()
		target = filepath.Clean(home)
	}
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		c.Dir = target
	} else if len(parts) >= 2 {
		c.AddOutput(fmt.Sprintf("cd: %s: No such directory", parts[1]))
	}
	return true
}

// REQ-WIN-001: Windows 版コンソールの子プロセス出力の文字化け修正
// winDecoder は子プロセス(cmd.exe)の出力バイトを復号するデコーダーを構築する。
// Windows コンソールはシステム依存のコードページ(日本語環境では CP932)で出力する為、
// 既定では UTF-8 として扱われると文字化けする。アクティブコードページを取得し、
// 対応するデコーダーへ変換することで正しく表示する。
func winDecoder() *encoding.Decoder {
	cp, _ := windows.GetConsoleOutputCP()
	if cp == 0 {
		cp = uint32(windows.GetACP())
	}
	return decodeForCP(cp)
}

// decodeForCP はコードページ番号に対応する復号 Decoder を返す。
// 未対応番号は日本語環境を想定して ShiftJIS(CP932)にフォールバックする。
func decodeForCP(cp uint32) *encoding.Decoder {
	switch cp {
	case 437:
		return charmap.CodePage437.NewDecoder()
	case 932:
		return japanese.ShiftJIS.NewDecoder()
	case 936:
		return simplifiedchinese.GBK.NewDecoder()
	case 949:
		return korean.EUCKR.NewDecoder()
	case 950:
		return traditionalchinese.Big5.NewDecoder()
	case 1250:
		return charmap.Windows1250.NewDecoder()
	case 1251:
		return charmap.Windows1251.NewDecoder()
	case 1252:
		return charmap.Windows1252.NewDecoder()
	case 1253:
		return charmap.Windows1253.NewDecoder()
	case 1254:
		return charmap.Windows1254.NewDecoder()
	case 1255:
		return charmap.Windows1255.NewDecoder()
	case 1256:
		return charmap.Windows1256.NewDecoder()
	case 1257:
		return charmap.Windows1257.NewDecoder()
	case 1258:
		return charmap.Windows1258.NewDecoder()
	case 65001:
		return encoding.Nop.NewDecoder()
	default:
		return japanese.ShiftJIS.NewDecoder()
	}
}

// decode はバイト列を子プロセスのコードページから UTF-8 へ変換する。
func decode(dec *encoding.Decoder, b []byte) string {
	s, _ := dec.String(string(b))
	return s
}

// mapUnixCmd は unix 流コマンドを Windows ネイティブコマンドに書き換える。
// 不明コマンド(dri 等の打間違い含む)は original の cmdLine をそのまま返す。
// REQ-WIN-002: unix 流コマンドを Windows ネティブコマンドに書き換える。
func mapUnixCmd(parts []string, cmdLine string) string {
	if len(parts) == 0 {
		return cmdLine
	}
	name := strings.ToLower(parts[0])
	rest := parts[1:]
	switch name {
	case "ls", "dir":
		var args []string
		for _, a := range rest {
			if strings.HasPrefix(a, "-") {
				continue
			}
			args = append(args, a)
		}
		if len(args) == 0 {
			return "dir"
		}
		return "dir " + joinArgs(args)
	case "cat":
		return wrapWith("type", rest)
	case "grep", "find":
		return wrapWith("findstr", convertGrepArgs(rest))
	case "rm":
		return wrapWith("del", rest)
	case "cp":
		return wrapWith("copy", rest)
	case "mv":
		return wrapWith("move", rest)
	case "head":
		return wrapWith("more", rest)
	case "touch":
		if len(rest) == 0 {
			return cmdLine
		}
		if len(rest) == 1 {
			return "type nul > " + joinArgs(rest)
		}
		var b strings.Builder
		for i, a := range rest {
			if i > 0 {
				b.WriteString(" & ")
			}
			b.WriteString("type nul > ")
			b.WriteString(joinArgs([]string{a}))
		}
		return b.String()
	default:
		return cmdLine
	}
}

func wrapWith(cmd string, rest []string) string {
	if len(rest) == 0 {
		return cmd
	}
	return cmd + " " + joinArgs(rest)
}

func convertGrepArgs(rest []string) []string {
	// single-char flag -> findstr switch
	single := map[byte]string{
		'i': "/i", 'r': "/s", 'R': "/s", 'l': "/l",
		'n': "/n", 'v': "/v", 'c': "/c",
	}
	out := make([]string, 0, len(rest))
	for _, a := range rest {
		if len(a) >= 2 && a[0] == '-' && a[1] != '-' {
			// combined short flags (e.g. -ri) -> individual findstr switches
			var extra string
			var expanded []string
			for _, ch := range a[1:] {
				if sw, ok := single[byte(ch)]; ok {
					expanded = append(expanded, sw)
				} else {
					extra += string(ch)
				}
			}
			if extra != "" {
				expanded = append(expanded, "-"+extra)
			}
			out = append(out, expanded...)
			continue
		}
		out = append(out, a)
	}
	return out
}

func joinArgs(args []string) string {
	var b strings.Builder
	for i, a := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		if strings.ContainsAny(a, " \t\"") {
			b.WriteByte('"')
			b.WriteString(a)
			b.WriteByte('"')
		} else {
			b.WriteString(a)
		}
	}
	return b.String()
}

func runExec(parts []string, cmdLine string, dir string, c *Console, outputCh chan string) {
	cmdLine = mapUnixCmd(parts, cmdLine)
	dec := winDecoder()

	cmd := exec.Command("cmd.exe", "/c", cmdLine)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PROMPT=$P$G")

	// CombinedOutput は失敗時も出力を返す為、err より先に出力を流す。
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		s, _ := dec.String(string(out))
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		lines := strings.Split(s, "\n")
		// 末尾の改行による空行は落とす
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		for _, line := range lines {
			outputCh <- line
		}
	}
	if err != nil {
		outputCh <- fmt.Sprintf("error: %v", err)
	}
}
