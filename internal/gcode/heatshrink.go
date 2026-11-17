package gcode

import "fmt"

// heatReader reads bits, the most significant first, from a byte slice.
type heatReader struct {
	data []byte
	pos  int
	bit  uint8
}

func (r *heatReader) readBit() (uint8, bool) {
	if r.pos >= len(r.data) {
		return 0, false
	}

	value := (r.data[r.pos] >> (7 - r.bit)) & 1
	r.bit++
	if r.bit == 8 {
		r.bit = 0
		r.pos++
	}

	return value, true
}

func (r *heatReader) readBits(count int) (uint16, bool) {
	var out uint16

	for range count {
		bit, ok := r.readBit()
		if !ok {
			return 0, false
		}

		out = out<<1 | uint16(bit)
	}

	return out, true
}

// unpackHeatshrink decodes Scott Mudge's heatshrink, the compression PrusaSlicer
// uses for the G-code blocks of a bgcode file. windowBits and lookaheadBits are
// the parameters the block's compression names.
func unpackHeatshrink(data []byte, windowBits, lookaheadBits int) ([]byte, error) {
	if windowBits < 4 || windowBits > 15 || lookaheadBits < 3 || lookaheadBits > 15 {
		return nil, fmt.Errorf("heatshrink: window %d and lookahead %d are out of range", windowBits, lookaheadBits)
	}

	reader := &heatReader{data: data}
	out := make([]byte, 0, len(data)*4)

	// The window starts as zeros: a back reference may reach into it before
	// the stream has written that far, and the reference decoder copies zeros
	// from there.
	windowSize := 1 << windowBits
	window := make([]byte, windowSize)
	head := 0

	for {
		tag, ok := reader.readBit()
		if !ok {
			break
		}

		if tag == 1 {
			literal, literalOK := reader.readBits(8)
			if !literalOK {
				break
			}

			value := byte(literal)
			out = append(out, value)
			window[head] = value
			head = (head + 1) % windowSize

			continue
		}

		index, indexOK := reader.readBits(windowBits)
		if !indexOK {
			break
		}

		count, countOK := reader.readBits(lookaheadBits)
		if !countOK {
			break
		}

		distance := int(index) + 1
		length := int(count) + 1
		start := (head - distance + windowSize) % windowSize

		for i := range length {
			value := window[(start+i)%windowSize]
			out = append(out, value)
			window[head] = value
			head = (head + 1) % windowSize
		}
	}

	return out, nil
}
