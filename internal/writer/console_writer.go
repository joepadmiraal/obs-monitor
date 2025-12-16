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
		fmt.Println("timestamp                 | rtt_ms   | stream_active | output_bytes | output_skipped_frames | errors")
		fmt.Println("--------------------------|----------|---------------|--------------|----------------------|--------")
		cw.headerPrinted = true
	}

	rttMs := "      -"
	if data.PingError == nil && data.RTT > 0 {
		rttMs = fmt.Sprintf("%7.2f", float64(data.RTT.Microseconds())/1000.0)
	}

	errors := ""
	if data.PingError != nil {
		errors = fmt.Sprintf("ping: %v", data.PingError)
	}
	if data.StreamError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("stream: %v", data.StreamError)
	}

	fmt.Printf("%25s | %8s | %13t | %12.0f | %21.0f | %s\n",
		data.Timestamp.Format(time.RFC3339),
		rttMs,
		data.StreamActive,
		data.OutputBytes,
		data.OutputSkippedFrames,
		errors,
	)

	return nil
}
