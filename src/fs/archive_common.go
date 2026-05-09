package fs

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	"github.com/nwaples/rardecode/v2"
)

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if err := extractZipEntry(f, dest); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(f *zip.File, dest string) error {
	target := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
		return nil
	}
	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open %s: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("create %s: %w", f.Name, err)
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, rc)
	return err
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

func extractTarBz2(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tbz2: %w", err)
	}
	defer func() { _ = f.Close() }()
	bzr := bzip2.NewReader(f)
	return untar(bzr, dest)
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

func extractBz2(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open bz2: %w", err)
	}
	defer func() { _ = f.Close() }()
	bzr := bzip2.NewReader(f)

	outName := filepath.Base(strings.TrimSuffix(src, ".bz2"))
	target := filepath.Join(dest, outName)
	out, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create %s: %w", outName, err)
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, bzr)
	return err
}

func extractSevenZip(src, dest string) error {
	r, err := sevenzip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open 7z: %w", err)
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
			return fmt.Errorf("open 7z entry %s: %w", f.Name, err)
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
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

func extractRar(src, dest string) error {
	err := extractRarPure(src, dest)
	if err == nil {
		return nil
	}
	return extractRarFallback(src, dest)
}

func extractRarPure(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open rar: %w", err)
	}
	defer func() { _ = f.Close() }()

	rr, err := rardecode.NewReader(f)
	if err != nil {
		return err
	}

	for {
		hdr, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}
		if hdr.IsDir {
			_ = os.MkdirAll(target, 0755)
			continue
		}
		_ = os.MkdirAll(filepath.Dir(target), 0755)
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.Mode())
		if err != nil {
			return fmt.Errorf("create %s: %w", hdr.Name, err)
		}
		_, err = io.Copy(out, rr)
		_ = out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractRarFallback(src, dest string) error {
	cmd := exec.Command("unar", "-o", dest, "-D", "-q", src)
	cmd.Dir = dest
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unar: %v\n%s", err, string(out))
	}
	return nil
}
