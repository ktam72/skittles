//go:build windows

package fs

import "os"

func lookupStat(fi os.FileInfo) (string, string) {
	return "", ""
}
