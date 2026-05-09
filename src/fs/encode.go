package fs

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func DecodeText(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// strip BOM
	trimmed := data
	if len(trimmed) >= 3 && trimmed[0] == 0xEF && trimmed[1] == 0xBB && trimmed[2] == 0xBF {
		trimmed = trimmed[3:]
	}

	if utf8.Valid(trimmed) {
		return string(trimmed)
	}

	// try Shift-JIS
	decoded, err := decodeShiftJIS(trimmed)
	if err == nil && len(decoded) > 0 {
		return decoded
	}

	// try EUC-JP
	decoded, err = decodeEUCJP(trimmed)
	if err == nil && len(decoded) > 0 {
		return decoded
	}

	// fallback: replace invalid bytes
	var buf strings.Builder
	for i := 0; i < len(trimmed); {
		r, size := utf8.DecodeRune(trimmed[i:])
		if r == utf8.RuneError && size == 1 {
			buf.WriteRune(rune(trimmed[i]))
			i++
		} else {
			buf.WriteRune(r)
			i += size
		}
	}
	return buf.String()
}

func decodeShiftJIS(data []byte) (string, error) {
	decoded, _, err := transform.Bytes(japanese.ShiftJIS.NewDecoder(), data)
	if err != nil {
		return "", err
	}
	// verify the result is reasonable (mostly valid UTF-8)
	if !utf8.Valid(decoded) {
		return "", nil
	}
	return string(decoded), nil
}

func decodeEUCJP(data []byte) (string, error) {
	decoded, _, err := transform.Bytes(japanese.EUCJP.NewDecoder(), data)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(decoded) {
		return "", nil
	}
	return string(decoded), nil
}
