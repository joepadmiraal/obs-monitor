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
		"obs_rtt_ms",
		"obs_ping_error",
		"google_rtt_ms",
		"google_ping_error",
		"stream_active",
		"output_bytes",
		"output_skipped_frames",
		"stream_error",
		"obs_cpu_percent",
		"obs_memory_mb",
		"obs_stats_error",
		"system_cpu_percent",
		"system_memory_percent",
		"system_metrics_error",
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
	obsRttMs := ""
	if data.ObsPingError == nil && data.ObsRTT > 0 {
		obsRttMs = fmt.Sprintf("%.2f", float64(data.ObsRTT.Microseconds())/1000.0)
	}

	googleRttMs := ""
	if data.GooglePingError == nil && data.GoogleRTT > 0 {
		googleRttMs = fmt.Sprintf("%.2f", float64(data.GoogleRTT.Microseconds())/1000.0)
	}

	obsPingError := ""
	if data.ObsPingError != nil {
		obsPingError = data.ObsPingError.Error()
	}

	googlePingError := ""
	if data.GooglePingError != nil {
		googlePingError = data.GooglePingError.Error()
	}

	streamError := ""
	if data.StreamError != nil {
		streamError = data.StreamError.Error()
	}

	obsStatsError := ""
	if data.ObsStatsError != nil {
		obsStatsError = data.ObsStatsError.Error()
	}

	systemMetricsError := ""
	if data.SystemMetricsError != nil {
		systemMetricsError = data.SystemMetricsError.Error()
	}

	row := []string{
		data.Timestamp.Format(time.RFC3339),
		obsRttMs,
		obsPingError,
		googleRttMs,
		googlePingError,
		fmt.Sprintf("%t", data.StreamActive),
		fmt.Sprintf("%.0f", data.OutputBytes),
		fmt.Sprintf("%.0f", data.OutputSkippedFrames),
		streamError,
		fmt.Sprintf("%.2f", data.ObsCpuUsage),
		fmt.Sprintf("%.2f", data.ObsMemoryUsage),
		obsStatsError,
		fmt.Sprintf("%.2f", data.SystemCpuUsage),
		fmt.Sprintf("%.2f", data.SystemMemoryUsage),
		systemMetricsError,
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
