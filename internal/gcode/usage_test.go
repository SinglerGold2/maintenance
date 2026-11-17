package gcode

import (
	"os"
	"testing"
)

func TestParseCountsExtrusionPerTool(t *testing.T) {
	t.Parallel()

	text := []byte("M83\nT0\nG1 X1 E2.5\nG1 X2 E1.5\nT1\nG1 X3 E4\nT0\nG1 X4 E1\n")
	usage := Parse(text)

	if got := usage.Extrusion[0]; got != 5 {
		t.Errorf("tool 0 pushed %v mm, want 5", got)
	}

	if got := usage.Extrusion[1]; got != 4 {
		t.Errorf("tool 1 pushed %v mm, want 4", got)
	}

	if usage.Changes[1] != 1 {
		t.Errorf("tool 1 took %d changes, want 1", usage.Changes[1])
	}
}

func TestParseReadsAbsoluteExtrusion(t *testing.T) {
	t.Parallel()

	text := []byte("M82\nG1 X1 E10\nG1 X2 E12\nG1 X3 E12\nG1 X4 E15\n")
	usage := Parse(text)

	if got := usage.Extrusion[0]; got != 15 {
		t.Errorf("tool 0 pushed %v mm, want 15", got)
	}
}

func TestParseIgnoresAPureExtrude(t *testing.T) {
	t.Parallel()

	// A retract and an unretract carry E alone; they use no filament the slicer
	// counts.
	text := []byte("M83\nG1 X1 E2\nG1 E-1\nG1 E1\nG1 X2 E3\n")

	if got := Parse(text).Extrusion[0]; got != 5 {
		t.Errorf("tool 0 pushed %v mm, want 5", got)
	}
}

// TestSampleFile reads a real file when GCODE_SAMPLE names one.
func TestSampleFile(t *testing.T) {
	path := os.Getenv("GCODE_SAMPLE")
	if path == "" {
		t.Skip("GCODE_SAMPLE is not set")
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	text, err := ReadAll(file, info.Size())
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if len(text) < 1000 {
		t.Fatalf("the text is %d bytes, too short to be G-code", len(text))
	}

	usage := Parse(text)
	t.Logf("lines %d, tools %v, extrusion %v", usage.Lines, usage.Changes, usage.Extrusion)

	if usage.Lines < 100 {
		t.Errorf("%d motion lines is too few", usage.Lines)
	}
}
