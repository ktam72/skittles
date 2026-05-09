package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode())
		}
		return CopyFile(path, target)
	})
}

func Move(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func Delete(path string) error {
	return os.RemoveAll(path)
}

func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}

func Touch(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func CopyEntries(srcDir string, entries []Entry, dstDir string) (int, error) {
	count := 0
	for _, e := range entries {
		src := e.Path
		dst := filepath.Join(dstDir, e.Name)
		var err error
		if e.IsDir {
			err = CopyDir(src, dst)
		} else {
			err = CopyFile(src, dst)
		}
		if err != nil {
			return count, fmt.Errorf("copy %s: %w", e.Name, err)
		}
		count++
	}
	return count, nil
}

func MoveEntries(srcDir string, entries []Entry, dstDir string) (int, error) {
	count := 0
	for _, e := range entries {
		if e.IsDir {
			if err := os.MkdirAll(filepath.Join(dstDir, e.Name), 0755); err != nil {
				return count, err
			}
			count++
			continue
		}
		if err := Move(e.Path, filepath.Join(dstDir, e.Name)); err != nil {
			if err := CopyFile(e.Path, filepath.Join(dstDir, e.Name)); err != nil {
				return count, err
			}
			_ = os.Remove(e.Path)
		}
		count++
	}
	return count, nil
}
