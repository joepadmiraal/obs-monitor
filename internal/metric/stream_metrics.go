package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
)

type StreamMetrics struct {
	client            *goobs.Client
	maxOutputBytes    float64
	prevOutputBytes   float64
	maxSkippedFrames  float64
	prevSkippedFrames float64
	lastActive        bool
	lastError         error
	measurementCount  int
	mu                sync.Mutex
	interval          time.Duration
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

func (s *StreamMetrics) GetAndResetMaxValues() StreamMetricsData {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxBytes := s.maxOutputBytes
	maxSkipped := s.maxSkippedFrames
	active := s.lastActive
	err := s.lastError

	if s.measurementCount < 2 {
		s.prevOutputBytes = maxBytes
		s.prevSkippedFrames = maxSkipped
		s.maxOutputBytes = 0
		s.maxSkippedFrames = 0
		s.lastError = nil
		return StreamMetricsData{
			Timestamp:           time.Now(),
			Active:              active,
			OutputBytes:         0,
			OutputSkippedFrames: 0,
			Error:               err,
		}
	}

	bytesDelta := maxBytes - s.prevOutputBytes
	skippedDelta := maxSkipped - s.prevSkippedFrames
	s.prevOutputBytes = maxBytes
	s.prevSkippedFrames = maxSkipped
	s.maxOutputBytes = 0
	s.maxSkippedFrames = 0
	s.lastError = nil

	return StreamMetricsData{
		Timestamp:           time.Now(),
		Active:              active,
		OutputBytes:         bytesDelta,
		OutputSkippedFrames: skippedDelta,
		Error:               err,
	}
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
			s.measurementCount++
		}
		s.mu.Unlock()

		if err != nil {
			fmt.Printf("Error getting stream status: %v\n", err)
			continue
		}
	}

	return nil
}
