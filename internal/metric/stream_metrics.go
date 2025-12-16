package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
)

type StreamMetrics struct {
	client           *goobs.Client
	metricsChan      chan StreamMetricsData
	maxOutputBytes   float64
	maxSkippedFrames float64
	lastActive       bool
	lastError        error
	mu               sync.Mutex
	interval         time.Duration
}

type StreamMetricsData struct {
	Timestamp           time.Time
	Active              bool
	OutputBytes         float64
	OutputSkippedFrames float64
	Error               error
}

func NewStreamMetrics(client *goobs.Client, interval time.Duration) (*StreamMetrics, error) {
	return &StreamMetrics{
		client:   client,
		interval: interval,
	}, nil
}

func (s *StreamMetrics) GetAndResetMaxValues() (float64, float64, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxBytes := s.maxOutputBytes
	maxSkipped := s.maxSkippedFrames
	active := s.lastActive
	err := s.lastError

	s.maxOutputBytes = 0
	s.maxSkippedFrames = 0
	s.lastError = nil

	return maxBytes, maxSkipped, active, err
}

func (s *StreamMetrics) Start() error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		status, err := s.client.Stream.GetStreamStatus()

		s.mu.Lock()
		if err != nil {
			s.lastError = err
		} else {
			s.lastActive = status.OutputActive
			if status.OutputBytes > s.maxOutputBytes {
				s.maxOutputBytes = status.OutputBytes
			}
			if status.OutputSkippedFrames > s.maxSkippedFrames {
				s.maxSkippedFrames = status.OutputSkippedFrames
			}
		}
		s.mu.Unlock()

		if err != nil {
			fmt.Printf("Error getting stream status: %v\n", err)
			continue
		}
	}

	return nil
}
