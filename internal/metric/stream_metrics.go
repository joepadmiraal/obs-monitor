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
	maxTotalFrames    float64
	prevTotalFrames   float64
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
	OutputFrames        float64
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
		maxTotalFrames := s.maxTotalFrames
		s.prevOutputBytes = maxBytes
		s.prevSkippedFrames = maxSkipped
		s.prevTotalFrames = maxTotalFrames
		s.maxOutputBytes = 0
		s.maxSkippedFrames = 0
		s.maxTotalFrames = 0
		s.lastError = nil
		return StreamMetricsData{
			Timestamp:           time.Now(),
			Active:              active,
			OutputBytes:         0,
			OutputSkippedFrames: 0,
			OutputFrames:        0,
			Error:               err,
		}
	}

	maxTotalFrames := s.maxTotalFrames
	bytesDelta := maxBytes - s.prevOutputBytes
	skippedDelta := maxSkipped - s.prevSkippedFrames
	framesDelta := maxTotalFrames - s.prevTotalFrames
	s.prevOutputBytes = maxBytes
	s.prevSkippedFrames = maxSkipped
	s.prevTotalFrames = maxTotalFrames
	s.lastError = nil

	return StreamMetricsData{
		Timestamp:           time.Now(),
		Active:              active,
		OutputBytes:         bytesDelta,
		OutputSkippedFrames: skippedDelta,
		OutputFrames:        framesDelta,
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
			if status.OutputTotalFrames > s.maxTotalFrames {
				s.maxTotalFrames = status.OutputTotalFrames
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
