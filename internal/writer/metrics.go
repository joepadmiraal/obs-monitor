package writer

import "time"

// MetricsData holds all metrics data for a single measurement
type MetricsData struct {
	Timestamp           time.Time
	ObsRTT              time.Duration
	ObsPingError        error
	GoogleRTT           time.Duration
	GooglePingError     error
	StreamActive        bool
	OutputBytes         float64
	OutputSkippedFrames float64
	StreamError         error
	ObsCpuUsage         float64
	ObsMemoryUsage      float64
	ObsStatsError       error
	SystemCpuUsage      float64
	SystemMemoryUsage   float64
	SystemMetricsError  error
}
