package utils

import (
	"bytes"
)

// StripUTF8BOM try to remove a file byte stream's UTF8 BOM header.
func StripUTF8BOM(stream []byte) []byte {
	if len(stream) > 3 && bytes.HasPrefix(stream, []byte{0xef, 0xbb, 0xbf}) {
		return stream[3:]
	}
	return stream
}
