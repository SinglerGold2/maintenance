package gcode

import "fmt"

// meatpackTable is the 15-character alphabet MeatPack packs into four bits. The
// sixteenth value, 0xF, announces a full-width character.
var meatpackTable = [16]byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', ' ', '\n', 'G', 'X', 0,
}

const (
	meatpackCommand     = 0xFF
	meatpackEnable      = 0xFB
	meatpackDisable     = 0xFA
	meatpackReset       = 0xF9
	meatpackQuery       = 0xF8
	meatpackNoSpacesOn  = 0xF7
	meatpackNoSpacesOff = 0xF6
	meatpackSpaceSlot   = 11
)

// unpackMeatpack decodes Scott Mudge's MeatPack, the encoding a bgcode G-code
// block may use. The stream itself says when a space stands for an E: the
// no-spaces variant is switched on and off by a command.
func unpackMeatpack(data []byte) ([]byte, error) {
	out := make([]byte, 0, len(data)*2)
	packing := true
	noSpaces := false

	var (
		queue  int
		buffer byte
		held   bool
	)

	character := func(nibble byte) byte {
		if noSpaces && nibble == meatpackSpaceSlot {
			return 'E'
		}

		return meatpackTable[nibble]
	}

	for pos := 0; pos < len(data); pos++ {
		current := data[pos]

		if current == meatpackCommand && pos+1 < len(data) && data[pos+1] == meatpackCommand {
			if pos+2 >= len(data) {
				break
			}

			switch data[pos+2] {
			case meatpackEnable:
				packing = true
			case meatpackDisable:
				packing = false
			case meatpackReset:
				queue, held = 0, false
			case meatpackQuery:
				// Nothing to report: this decoder holds no configuration.
			case meatpackNoSpacesOn:
				noSpaces = true
			case meatpackNoSpacesOff:
				noSpaces = false
			default:
				return nil, fmt.Errorf("meatpack: unknown command %#x", data[pos+2])
			}

			pos += 2

			continue
		}

		if !packing {
			out = append(out, current)

			continue
		}

		low := current & 0x0F
		high := (current >> 4) & 0x0F
		firstFull := low == 0x0F
		secondFull := high == 0x0F

		if queue > 0 {
			out = append(out, current)
			if held {
				out = append(out, buffer)
				held = false
			}

			queue--

			continue
		}

		if firstFull {
			queue++
			if secondFull {
				queue++
			} else {
				buffer, held = character(high), true
			}

			continue
		}

		first := character(low)
		out = append(out, first)

		if first == '\n' {
			continue
		}

		if secondFull {
			queue++

			continue
		}

		out = append(out, character(high))
	}

	return out, nil
}
