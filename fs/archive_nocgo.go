//go:build !cgo

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ExtractToTemp(src string) (string, error) {
	ext := strings.ToLower(filepath.Ext(src))
	base := filepath.Base(src)
	trimmed := strings.TrimSuffix(base, ext)
	if trimmed == "" {
		trimmed = "archive"
	}
	dest, err := os.MkdirTemp("", "skittles-"+trimmed+"-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}

	switch ext {
	case ".zip":
		err = extractZip(src, dest)
	case ".tar":
		err = extractTar(src, dest)
	case ".tgz", ".gz":
		if strings.HasSuffix(strings.ToLower(src), ".tar.gz") ||
			strings.HasSuffix(strings.ToLower(src), ".tgz") {
			err = extractTarGz(src, dest)
		} else {
			err = extractGz(src, dest)
		}
	case ".bz2":
		err = extractUsing(src, dest, "bunzip2", "-c")
	case ".lzh", ".lha":
		err = extractUsing(src, dest, "lha", "x")
	case ".rar":
		err = extractUsing(src, dest, "unrar", "x")
	case ".7z":
		err = extractUsing(src, dest, "7z", "x")
	default:
		err = fmt.Errorf("unsupported archive: %s", ext)
	}

	if err != nil {
		_ = os.RemoveAll(dest)
		return "", err
	}
	return dest, nil
}
