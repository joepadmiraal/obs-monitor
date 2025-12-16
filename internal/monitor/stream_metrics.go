package monitor

import (
	"fmt"
	"time"

	"github.com/andreykaipov/goobs"
)

type StreamMetrics struct {
	client      *goobs.Client
	metricsChan chan StreamMetricsData
}

// StreamMetricsData holds stream measurement data
type StreamMetricsData struct {
	Timestamp           time.Time
	Active              bool
	OutputBytes         float64
	OutputSkippedFrames float64
	Error               error
}

// NewStreamMetrics creates a new stream metrics monitor
func NewStreamMetrics(client *goobs.Client) (*StreamMetrics, error) {
	return &StreamMetrics{
		client:      client,
		metricsChan: make(chan StreamMetricsData, 10),
	}, nil
}

// GetMetricsChan returns the channel for receiving stream metrics
func (s *StreamMetrics) GetMetricsChan() <-chan StreamMetricsData {
	return s.metricsChan
}

// Start begins monitoring stream metrics every second
func (s *StreamMetrics) Start() error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		timestamp := time.Now()
		status, err := s.client.Stream.GetStreamStatus()

		metricsData := StreamMetricsData{
			Timestamp: timestamp,
			Error:     err,
		}

		if err == nil {
			metricsData.Active = status.OutputActive
			metricsData.OutputBytes = status.OutputBytes
			metricsData.OutputSkippedFrames = status.OutputSkippedFrames
		}

		// Send metrics to channel
		select {
		case s.metricsChan <- metricsData:
		default:
			// Channel full, skip this measurement
		}

		if err != nil {
			fmt.Printf("Error getting stream status: %v\n", err)
			continue
		}
	}

	return nil
}
