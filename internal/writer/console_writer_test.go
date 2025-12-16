package writer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConsoleWriter_WriteMetrics_HighRTT(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	// Test with 1001ms RTT (1001000 microseconds)
	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 16, 10, 0, 0, 0, time.UTC),
		RTT:                 1001 * time.Millisecond,
		PingError:           nil,
		StreamActive:        true,
		OutputBytes:         123456.0,
		OutputSkippedFrames: 10.0,
		StreamError:         nil,
	}

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	// Close writer and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (header, separator, data), got %d", len(lines))
	}

	// Check that the data line has properly formatted RTT
	dataLine := lines[2]

	// The RTT value should be formatted as a 6-character string (right-aligned)
	// For 1001.00ms, it should appear in the output
	if !strings.Contains(dataLine, "1001.00") {
		t.Errorf("Expected RTT value '1001.00' in output, got: %s", dataLine)
	}

	// Verify the columns are aligned by checking pipe positions
	headerLine := lines[0]

	// All lines should have pipes at the same positions
	headerPipes := findPipePositions(headerLine)
	dataPipes := findPipePositions(dataLine)

	if len(headerPipes) != len(dataPipes) {
		t.Errorf("Number of columns mismatch: header has %d pipes, data has %d pipes", len(headerPipes), len(dataPipes))
	}

	for i := range headerPipes {
		if i < len(dataPipes) && headerPipes[i] != dataPipes[i] {
			t.Errorf("Column %d misaligned: header pipe at %d, data pipe at %d", i, headerPipes[i], dataPipes[i])
			t.Logf("Header: %s", headerLine)
			t.Logf("Data:   %s", dataLine)
		}
	}
}

func findPipePositions(line string) []int {
	positions := []int{}
	for i, ch := range line {
		if ch == '|' {
			positions = append(positions, i)
		}
	}
	return positions
}
