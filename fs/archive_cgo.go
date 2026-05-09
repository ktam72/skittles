//go:build cgo

package fs

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/libarchive/lib -larchive
#cgo CFLAGS: -I/opt/homebrew/opt/libarchive/include

#include <archive.h>
#include <archive_entry.h>
#include <locale.h>
#include <stdlib.h>

void init_locale(void) {
    setlocale(LC_ALL, "en_US.UTF-8");
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func ExtractToTemp(src string) (string, error) {
	C.init_locale()

	base := filepath.Base(src)
	ext := strings.ToLower(filepath.Ext(src))
	trimmed := strings.TrimSuffix(base, ext)
	if trimmed == "" {
		trimmed = "archive"
	}
	dest, err := os.MkdirTemp("", "skittles-"+trimmed+"-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}

	if err := extractWithLibarchive(src, dest); err != nil {
		_ = os.RemoveAll(dest)
		// fallback to Go/external tools
		return extractFallback(src)
	}
	return dest, nil
}

func extractFallback(src string) (string, error) {
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

func extractWithLibarchive(src, dest string) error {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))

	a := C.archive_read_new()
	if a == nil {
		return fmt.Errorf("archive_read_new failed")
	}
	defer C.archive_read_free(a)

	C.archive_read_support_filter_all(a)
	C.archive_read_support_format_all(a)

	if ret := C.archive_read_open_filename(a, csrc, 10240); ret != C.ARCHIVE_OK {
		return fmt.Errorf("archive_read_open_filename: %s", C.GoString(C.archive_error_string(a)))
	}
	defer C.archive_read_close(a)

	var entry *C.struct_archive_entry
	for {
		r := C.archive_read_next_header(a, &entry)
		if r == C.ARCHIVE_EOF {
			break
		}
		if r != C.ARCHIVE_OK {
			return fmt.Errorf("archive_read_next_header: %s", C.GoString(C.archive_error_string(a)))
		}

		name := C.GoString(C.archive_entry_pathname_utf8(entry))
		if name == "" {
			name = C.GoString(C.archive_entry_pathname(entry))
		}
		target := filepath.Join(dest, name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		fileType := C.archive_entry_filetype(entry)
		ft := int(fileType & C.S_IFMT)
		mode := os.FileMode(C.archive_entry_mode(entry))

		switch ft {
		case int(C.S_IFDIR):
			_ = os.MkdirAll(target, mode.Perm())

		case int(C.S_IFLNK):
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			linkTarget := C.GoString(C.archive_entry_symlink(entry))
			if linkTarget != "" {
				_ = os.Symlink(linkTarget, target)
			}

		default:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return fmt.Errorf("create %s: %w", name, err)
			}

			buf := make([]byte, 65536)
			for {
				n := int(C.archive_read_data(a, unsafe.Pointer(&buf[0]), C.size_t(len(buf))))
				if n < 0 {
					_ = out.Close()
					return fmt.Errorf("read data %s: %s", name, C.GoString(C.archive_error_string(a)))
				}
				if n == 0 {
					break
				}
				if _, err := out.Write(buf[:n]); err != nil {
					_ = out.Close()
					return fmt.Errorf("write %s: %w", name, err)
				}
			}
			_ = out.Close()
		}
	}
	return nil
}
