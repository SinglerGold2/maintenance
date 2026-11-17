package gcode

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Usage is what a G-code file says about the tools it uses.
type Usage struct {
	// Extrusion is the filament each tool pushed, in millimetres.
	Extrusion map[int]float64
	// Changes counts the tool changes each tool took part in.
	Changes map[int]int
	// Lines is how many lines carried a motion command.
	Lines int
}

// Parse reads the tool changes and the extrusion of a G-code text.
func Parse(text []byte) Usage {
	return ParseReader(strings.NewReader(string(text)))
}

// ParseReader reads the tool changes and the extrusion line by line, so a print
// of a hundred megabytes costs no more memory than its longest line.
func ParseReader(reader io.Reader) Usage {
	usage := Usage{Extrusion: map[int]float64{}, Changes: map[int]int{}}

	tool := 0
	relative := false
	lastE := 0.0
	seenE := false
	lastLine := ""

	buffered := bufio.NewReaderSize(reader, 1<<20)

	for {
		raw, err := buffered.ReadString('\n')
		line := strings.TrimSpace(raw)

		if line == "" && err != nil {
			break
		}

		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		if code, value, ok := command(line); ok {
			switch code {
			case "M82":
				relative = false
			case "M83":
				relative = true
			case "G92":
				if e, found := numberAfter(value, 'E'); found {
					lastE = e
					seenE = true
				}
			}

			continue
		}

		if line[0] == 'T' {
			// A tool change carries parameters of its own, as "T0 S1 L2 D0"
			// does, so only the digits right after the T name the tool.
			end := 1
			for end < len(line) && line[end] >= '0' && line[end] <= '9' {
				end++
			}

			if end > 1 {
				if next, err := strconv.Atoi(line[1:end]); err == nil {
					if next != tool {
						usage.Changes[next]++
					}

					tool = next
				}
			}

			continue
		}

		if !strings.HasPrefix(line, "G0") && !strings.HasPrefix(line, "G1") &&
			!strings.HasPrefix(line, "G2") && !strings.HasPrefix(line, "G3") {
			continue
		}

		usage.Lines++

		e, found := numberAfter(line, 'E')
		if !found {
			continue
		}

		// A pure extrude or retract carries E alone; the slicer does not count
		// that filament as used, and neither does this reader.
		_, movesX := numberAfter(line, 'X')
		_, movesY := numberAfter(line, 'Y')
		if !movesX && !movesY {
			continue
		}

		if !seenE {
			lastE = 0
			seenE = true
		}

		var pushed float64

		if relative {
			pushed = e
		} else {
			pushed = e - lastE
			lastE = e
		}

		if pushed > 0 {
			usage.Extrusion[tool] += pushed
		}

		lastLine = line
	}

	_ = lastLine

	return usage
}

// command answers whether a line is one of the modal commands this parser reads.
func command(line string) (string, string, bool) {
	for _, code := range []string{"M82", "M83", "G92"} {
		if strings.HasPrefix(line, code) {
			return code, line, true
		}
	}

	return "", "", false
}

// numberAfter reads the number that follows a letter, in millimetres or
// millimetres per second as the G-code writes it.
func numberAfter(line string, letter byte) (float64, bool) {
	for i := 0; i < len(line); i++ {
		if line[i] != letter {
			continue
		}

		start := i + 1
		end := start

		for end < len(line) {
			c := line[end]
			if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
				end++

				continue
			}

			break
		}

		if end == start {
			continue
		}

		value, err := strconv.ParseFloat(line[start:end], 64)
		if err != nil {
			continue
		}

		return value, true
	}

	return 0, false
}
