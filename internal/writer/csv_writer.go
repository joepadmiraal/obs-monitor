package writer

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// CSVWriter handles writing metrics to a CSV file
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

// NewCSVWriter creates a new CSV writer and writes the header
func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write header
	header := []string{
		"timestamp",
		"rtt_ms",
		"ping_error",
		"stream_active",
		"output_bytes",
		"output_skipped_frames",
		"stream_error",
	}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}
	writer.Flush()

	return &CSVWriter{
		file:   file,
		writer: writer,
	}, nil
}

// WriteMetrics writes a single metrics data row to the CSV file
func (cw *CSVWriter) WriteMetrics(data MetricsData) error {
	rttMs := ""
	if data.PingError == nil && data.RTT > 0 {
		rttMs = fmt.Sprintf("%.2f", float64(data.RTT.Microseconds())/1000.0)
	}

	pingError := ""
	if data.PingError != nil {
		pingError = data.PingError.Error()
	}

	streamError := ""
	if data.StreamError != nil {
		streamError = data.StreamError.Error()
	}

	row := []string{
		data.Timestamp.Format(time.RFC3339),
		rttMs,
		pingError,
		fmt.Sprintf("%t", data.StreamActive),
		fmt.Sprintf("%.0f", data.OutputBytes),
		fmt.Sprintf("%.0f", data.OutputSkippedFrames),
		streamError,
	}

	if err := cw.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	cw.writer.Flush()

	return cw.writer.Error()
}

// Close closes the CSV file
func (cw *CSVWriter) Close() error {
	cw.writer.Flush()
	return cw.file.Close()
}
