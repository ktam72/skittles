package fs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func SearchFiles(root, pattern string) ([]string, error) {
	var results []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		name := fi.Name()
		if matched, _ := filepath.Match(pattern, name); matched {
			rel, _ := filepath.Rel(root, path)
			results = append(results, rel)
		}
		return nil
	})
	return results, err
}

func GrepFiles(root, pattern string) ([]string, error) {
	pat := strings.ToLower(pattern)
	var results []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if isBinaryFile(path) {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer func() { _ = f.Close() }()
		rel, _ := filepath.Rel(root, path)
		scanner := bufio.NewScanner(f)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), pat) {
				results = append(results, formatGrepResult(rel, lineNum, line))
			}
			lineNum++
		}
		return nil
	})
	return results, err
}

func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

func formatGrepResult(path string, lineNum int, line string) string {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) > 120 {
		trimmed = trimmed[:120]
	}
	return strings.ReplaceAll(path, " ", "_") + ":" + itoa(lineNum) + ":" + trimmed
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
