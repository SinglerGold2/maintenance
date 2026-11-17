package gcode

import (
	"bytes"
	"compress/zlib"
	"io"
)

// unpackDeflate inflates a block compressed with the deflate algorithm.
func unpackDeflate(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
