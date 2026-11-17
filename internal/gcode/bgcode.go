package gcode

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Block types of the bgcode container.
const (
	blockFileMetadata = 0
	blockGCode        = 1
	blockSlicerData   = 2
	blockPrinterData  = 3
	blockPrintData    = 4
	blockThumbnail    = 5
)

// Compression values of a block.
const (
	compressionNone       = 0
	compressionDeflate    = 1
	compressionHeat11     = 2
	compressionHeat12     = 3
	encodingPlain         = 0
	encodingMeatpack      = 1
	encodingMeatpackNotes = 2
)

var errNotBgcode = errors.New("gcode: not a bgcode file")

// isBgcode answers whether a file starts with the bgcode magic.
func isBgcode(head []byte) bool {
	return len(head) >= 4 && string(head[:4]) == "GCDE"
}

// readBgcode walks the container and returns the G-code text of every G-code
// block, decompressed and decoded.
func readBgcode(data []byte) ([]byte, error) {
	if !isBgcode(data) {
		return nil, errNotBgcode
	}

	if len(data) < 10 {
		return nil, fmt.Errorf("gcode: bgcode header is %d bytes", len(data))
	}

	checksum := binary.LittleEndian.Uint16(data[8:10])
	blockChecksum := 0
	if checksum != 0 {
		blockChecksum = 4
	}

	var out bytes.Buffer
	pos := 10

	for pos+8 <= len(data) {
		blockType := binary.LittleEndian.Uint16(data[pos:])
		compression := binary.LittleEndian.Uint16(data[pos+2:])

		var (
			size uint32
			head int
		)

		if compression == compressionNone {
			size = binary.LittleEndian.Uint32(data[pos+4:])
			head = 8
		} else {
			if pos+12 > len(data) {
				break
			}

			size = binary.LittleEndian.Uint32(data[pos+8:])
			head = 12
		}

		params := 2
		if blockType == blockThumbnail {
			params = 6
		}

		body := pos + head + params
		if body+int(size) > len(data) {
			return nil, fmt.Errorf("gcode: block of %d bytes runs past the file", size)
		}

		payload := data[body : body+int(size)]

		if blockType == blockGCode {
			encoding := binary.LittleEndian.Uint16(data[pos+head:])
			text, err := gcodeBlockText(payload, compression, encoding)
			if err != nil {
				return nil, err
			}

			out.Write(text)
		}

		pos = body + int(size) + blockChecksum
	}

	return out.Bytes(), nil
}

// gcodeBlockText decompresses and decodes one G-code block.
func gcodeBlockText(payload []byte, compression uint16, encoding uint16) ([]byte, error) {
	var raw []byte

	switch compression {
	case compressionNone:
		raw = payload
	case compressionDeflate:
		unpacked, err := unpackDeflate(payload)
		if err != nil {
			return nil, err
		}

		raw = unpacked
	case compressionHeat11:
		unpacked, err := unpackHeatshrink(payload, 11, 4)
		if err != nil {
			return nil, err
		}

		raw = unpacked
	case compressionHeat12:
		unpacked, err := unpackHeatshrink(payload, 12, 4)
		if err != nil {
			return nil, err
		}

		raw = unpacked
	default:
		return nil, fmt.Errorf("gcode: compression %d is not one this reader knows", compression)
	}

	switch encoding {
	case encodingPlain:
		return raw, nil
	case encodingMeatpack, encodingMeatpackNotes:
		return unpackMeatpack(raw)
	}

	return nil, fmt.Errorf("gcode: encoding %d is not one this reader knows", encoding)
}

// ReadAll reads a G-code file, text or bgcode, into one text buffer.
func ReadAll(reader io.ReaderAt, size int64) ([]byte, error) {
	if size > 256<<20 {
		return nil, fmt.Errorf("gcode: %d bytes is larger than this reader takes", size)
	}

	data := make([]byte, size)
	if _, err := reader.ReadAt(data, 0); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	if isBgcode(data) {
		return readBgcode(data)
	}

	return data, nil
}
