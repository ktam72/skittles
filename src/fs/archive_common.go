package fs

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(target, 0755)
			continue
		}
		_ = os.MkdirAll(filepath.Dir(target), 0755)
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open entry %s: %w", f.Name, err)
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("create %s: %w", f.Name, err)
		}
		_, err = io.Copy(out, rc)
		_ = rc.Close()
		_ = out.Close()
		if err != nil {
			return fmt.Errorf("write %s: %w", f.Name, err)
		}
	}
	return nil
}

func extractTar(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar: %w", err)
	}
	defer func() { _ = f.Close() }()
	return untar(f, dest)
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tgz: %w", err)
	}
	defer func() { _ = f.Close() }()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gr.Close() }()
	return untar(gr, dest)
}

func untar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("create %s: %w", hdr.Name, err)
			}
			_, err = io.Copy(out, tr)
			_ = out.Close()
			if err != nil {
				return fmt.Errorf("write %s: %w", hdr.Name, err)
			}
		}
	}
	return nil
}

func extractGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open gz: %w", err)
	}
	defer func() { _ = f.Close() }()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gr.Close() }()

	outName := filepath.Base(strings.TrimSuffix(src, ".gz"))
	target := filepath.Join(dest, outName)
	out, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create %s: %w", outName, err)
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, gr)
	return err
}

func extractUsing(src, dest, tool string, args ...string) error {
	all := append(args, src)
	cmd := exec.Command(tool, all...)
	cmd.Dir = dest
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %v\n%s", tool, err, string(out))
	}
	return nil
}
