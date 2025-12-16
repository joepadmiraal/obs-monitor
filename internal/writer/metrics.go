package writer

import "time"

// MetricsData holds all metrics data for a single measurement
type MetricsData struct {
	Timestamp           time.Time
	RTT                 time.Duration
	PingError           error
	StreamActive        bool
	OutputBytes         float64
	OutputSkippedFrames float64
	StreamError         error
}
