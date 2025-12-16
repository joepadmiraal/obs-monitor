package writer

import (
	"fmt"
	"time"
)

// ConsoleWriter handles writing metrics to the console
type ConsoleWriter struct {
	headerPrinted bool
}

// NewConsoleWriter creates a new console writer
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		headerPrinted: false,
	}
}

// WriteMetrics writes a single metrics data row to the console
func (cw *ConsoleWriter) WriteMetrics(data MetricsData) error {
	// Print header on first call
	if !cw.headerPrinted {
		fmt.Println("timestamp                 | obs_rtt_ms | google_rtt_ms | stream_active | output_bytes | output_skipped_frames | obs_cpu_% | obs_mem_mb | sys_cpu_% | sys_mem_% | errors")
		fmt.Println("--------------------------|------------|---------------|---------------|--------------|-----------------------|-----------|------------|-----------|-----------|--------")
		cw.headerPrinted = true
	}

	obsRttMs := "         -"
	if data.ObsPingError == nil && data.ObsRTT > 0 {
		obsRttMs = fmt.Sprintf("%10.2f", float64(data.ObsRTT.Microseconds())/1000.0)
	}

	googleRttMs := "            -"
	if data.GooglePingError == nil && data.GoogleRTT > 0 {
		googleRttMs = fmt.Sprintf("%13.2f", float64(data.GoogleRTT.Microseconds())/1000.0)
	}

	errors := ""
	if data.ObsPingError != nil {
		errors = fmt.Sprintf("obs_ping: %v", data.ObsPingError)
	}
	if data.GooglePingError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("google_ping: %v", data.GooglePingError)
	}
	if data.StreamError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("stream: %v", data.StreamError)
	}
	if data.ObsStatsError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("obs_stats: %v", data.ObsStatsError)
	}
	if data.SystemMetricsError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("system: %v", data.SystemMetricsError)
	}

	fmt.Printf("%25s | %10s | %13s | %13t | %12.0f | %21.0f | %9.1f | %10.0f | %9.1f | %9.1f | %s\n",
		data.Timestamp.Format(time.RFC3339),
		obsRttMs,
		googleRttMs,
		data.StreamActive,
		data.OutputBytes,
		data.OutputSkippedFrames,
		data.ObsCpuUsage,
		data.ObsMemoryUsage, data.SystemCpuUsage,
		data.SystemMemoryUsage, errors,
	)

	return nil
}
