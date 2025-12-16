package metric

import (
	"fmt"
	"time"

	"github.com/andreykaipov/goobs"
)

type StreamMetrics struct {
	client      *goobs.Client
	metricsChan chan StreamMetricsData
}

type StreamMetricsData struct {
	Timestamp           time.Time
	Active              bool
	OutputBytes         float64
	OutputSkippedFrames float64
	Error               error
}

func NewStreamMetrics(client *goobs.Client) (*StreamMetrics, error) {
	return &StreamMetrics{
		client:      client,
		metricsChan: make(chan StreamMetricsData, 10),
	}, nil
}

func (s *StreamMetrics) GetMetricsChan() <-chan StreamMetricsData {
	return s.metricsChan
}

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

		select {
		case s.metricsChan <- metricsData:
		default:
		}

		if err != nil {
			fmt.Printf("Error getting stream status: %v\n", err)
			continue
		}
	}

	return nil
}
