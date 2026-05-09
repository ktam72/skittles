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
	case ".tbz2", ".bz2":
		if strings.HasSuffix(strings.ToLower(src), ".tar.bz2") ||
			strings.HasSuffix(strings.ToLower(src), ".tbz2") {
			err = extractTarBz2(src, dest)
		} else {
			err = extractBz2(src, dest)
		}
	case ".7z":
		err = extractSevenZip(src, dest)
	case ".lzh", ".lha":
		err = extractUsing(src, dest, "lha", "x")
case ".rar":
	err = extractRar(src, dest)
	default:
		err = fmt.Errorf("unsupported archive: %s", ext)
	}

	if err != nil {
		_ = os.RemoveAll(dest)
		return "", err
	}
	return dest, nil
}

func ExtractCmdFor(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".zip":
		return "unzip -o $P"
	case ".tar":
		return "tar xf $P"
	case ".tgz", ".gz":
		if strings.HasSuffix(strings.ToLower(path), ".tar.gz") {
			return "tar xzf $P"
		}
		return "gunzip -c $P > ${P%.*}"
	case ".bz2":
		return "bunzip2 -c $P > ${P%.*}"
	case ".lzh", ".lha":
		return "lha x $P"
case ".rar":
	return "unar -o . -D -q $P"
	case ".7z":
		return "7z x $P"
	default:
		return ""
	}
}
