package fs

import "fmt"

func RenderHexView(data []byte) []string {
	const perLine = 16
	var lines []string
	for off := 0; off < len(data); off += perLine {
		offset := fmt.Sprintf("%08x", off)
		hexB := make([]byte, 0, 49)
		hexB = append(hexB, offset...)
		hexB = append(hexB, ' ')

		for i := 0; i < perLine; i++ {
			if off+i >= len(data) {
				hexB = append(hexB, "   "...)
			} else {
				b := data[off+i]
				hexB = append(hexB, hexDigit[b>>4], hexDigit[b&0xf], ' ')
			}
			if i == 7 {
				hexB = append(hexB, ' ')
			}
		}

		hexB = append(hexB, "  "...)

		for i := 0; i < perLine; i++ {
			if off+i >= len(data) {
				hexB = append(hexB, ' ')
			} else {
				b := data[off+i]
				if b >= 0x20 && b <= 0x7e {
					hexB = append(hexB, b)
				} else {
					hexB = append(hexB, '.')
				}
			}
		}

		lines = append(lines, string(hexB))
	}
	return lines
}

var hexDigit = [16]byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'a', 'b', 'c', 'd', 'e', 'f',
}
