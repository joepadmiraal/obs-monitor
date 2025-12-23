package metric

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStreamMetrics_GetAndResetMaxValues_NoMeasurements(t *testing.T) {
	sm := &StreamMetrics{
		measurementCount: 0,
	}

	data := sm.GetAndResetMaxValues()

	if data.OutputBytes != 0 {
		t.Errorf("Expected OutputBytes to be 0, got %f", data.OutputBytes)
	}
	if data.OutputSkippedFrames != 0 {
		t.Errorf("Expected OutputSkippedFrames to be 0, got %f", data.OutputSkippedFrames)
	}
	if data.OutputFrames != 0 {
		t.Errorf("Expected OutputFrames to be 0, got %f", data.OutputFrames)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_OneMeasurement(t *testing.T) {
	sm := &StreamMetrics{
		maxOutputBytes:   1000.0,
		maxSkippedFrames: 10.0,
		maxTotalFrames:   100.0,
		measurementCount: 1,
	}

	data := sm.GetAndResetMaxValues()

	if data.OutputBytes != 0 {
		t.Errorf("Expected OutputBytes to be 0 (not enough measurements for delta), got %f", data.OutputBytes)
	}
	if data.OutputSkippedFrames != 0 {
		t.Errorf("Expected OutputSkippedFrames to be 0, got %f", data.OutputSkippedFrames)
	}
	if data.OutputFrames != 0 {
		t.Errorf("Expected OutputFrames to be 0, got %f", data.OutputFrames)
	}

	if sm.prevOutputBytes != 1000.0 {
		t.Errorf("Expected prevOutputBytes to be set to 1000.0, got %f", sm.prevOutputBytes)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_MultipleMeasurements(t *testing.T) {
	sm := &StreamMetrics{
		prevOutputBytes:   1000.0,
		prevSkippedFrames: 10.0,
		prevTotalFrames:   100.0,
		maxOutputBytes:    2500.0,
		maxSkippedFrames:  25.0,
		maxTotalFrames:    250.0,
		measurementCount:  5,
	}

	data := sm.GetAndResetMaxValues()

	expectedBytes := 2500.0 - 1000.0
	if data.OutputBytes != expectedBytes {
		t.Errorf("Expected OutputBytes to be %f, got %f", expectedBytes, data.OutputBytes)
	}

	expectedSkipped := 25.0 - 10.0
	if data.OutputSkippedFrames != expectedSkipped {
		t.Errorf("Expected OutputSkippedFrames to be %f, got %f", expectedSkipped, data.OutputSkippedFrames)
	}

	expectedFrames := 250.0 - 100.0
	if data.OutputFrames != expectedFrames {
		t.Errorf("Expected OutputFrames to be %f, got %f", expectedFrames, data.OutputFrames)
	}

	if sm.prevOutputBytes != 2500.0 {
		t.Errorf("Expected prevOutputBytes to be updated to 2500.0, got %f", sm.prevOutputBytes)
	}
	if sm.prevSkippedFrames != 25.0 {
		t.Errorf("Expected prevSkippedFrames to be updated to 25.0, got %f", sm.prevSkippedFrames)
	}
	if sm.prevTotalFrames != 250.0 {
		t.Errorf("Expected prevTotalFrames to be updated to 250.0, got %f", sm.prevTotalFrames)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_MaxValueTracking(t *testing.T) {
	sm := &StreamMetrics{
		prevOutputBytes:  1000.0,
		maxOutputBytes:   3000.0,
		measurementCount: 3,
	}

	data1 := sm.GetAndResetMaxValues()
	expectedDelta1 := 3000.0 - 1000.0
	if data1.OutputBytes != expectedDelta1 {
		t.Errorf("Expected first call to return delta %f, got %f", expectedDelta1, data1.OutputBytes)
	}

	sm.maxOutputBytes = 3500.0
	sm.measurementCount = 3

	data2 := sm.GetAndResetMaxValues()
	expectedDelta2 := 3500.0 - 3000.0
	if data2.OutputBytes != expectedDelta2 {
		t.Errorf("Expected second call to return delta %f, got %f", expectedDelta2, data2.OutputBytes)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_ConcurrentAccess(t *testing.T) {
	sm := &StreamMetrics{
		prevOutputBytes:  1000.0,
		maxOutputBytes:   2000.0,
		measurementCount: 2,
	}

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sm.GetAndResetMaxValues()
		}()
	}

	wg.Wait()
}

func TestStreamMetrics_NewStreamMetrics(t *testing.T) {
	interval := 100 * time.Millisecond

	sm, err := NewStreamMetrics(nil, interval)

	if err != nil {
		t.Fatalf("NewStreamMetrics returned error: %v", err)
	}
	if sm == nil {
		t.Fatal("NewStreamMetrics returned nil")
	}
	if sm.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, sm.interval)
	}
}

func TestStreamMetrics_ErrorHandlingDuringCollection(t *testing.T) {
	sm := &StreamMetrics{
		maxOutputBytes:   1000.0,
		measurementCount: 5,
	}

	testError := fmt.Errorf("stream status error")

	sm.mu.Lock()
	sm.lastError = testError
	sm.mu.Unlock()

	sm.mu.Lock()
	if sm.lastError != testError {
		t.Error("Expected lastError to be set")
	}
	sm.mu.Unlock()

	data := sm.GetAndResetMaxValues()
	if data.Error != testError {
		t.Error("Expected error to be returned in data")
	}
}

func TestStreamMetrics_ActiveStateTracking(t *testing.T) {
	tests := []struct {
		name           string
		initialActive  bool
		newActive      bool
		expectedActive bool
	}{
		{
			name:           "inactive to active",
			initialActive:  false,
			newActive:      true,
			expectedActive: true,
		},
		{
			name:           "active to inactive",
			initialActive:  true,
			newActive:      false,
			expectedActive: false,
		},
		{
			name:           "remains active",
			initialActive:  true,
			newActive:      true,
			expectedActive: true,
		},
		{
			name:           "remains inactive",
			initialActive:  false,
			newActive:      false,
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StreamMetrics{
				lastActive: tt.initialActive,
			}

			sm.mu.Lock()
			sm.lastActive = tt.newActive
			sm.mu.Unlock()

			data := sm.GetAndResetMaxValues()
			if data.Active != tt.expectedActive {
				t.Errorf("Expected Active to be %v, got %v", tt.expectedActive, data.Active)
			}
		})
	}
}

func TestStreamMetrics_UpdateMetrics(t *testing.T) {
	tests := []struct {
		name                 string
		initialMaxBytes      float64
		initialMaxSkip       float64
		initialMaxFrames     float64
		initialMeasureCount  int
		outputActive         bool
		newBytes             float64
		newSkip              float64
		newFrames            float64
		expectedMaxBytes     float64
		expectedMaxSkip      float64
		expectedMaxFrames    float64
		expectedMeasureCount int
		expectedActive       bool
	}{
		{
			name:                 "first measurement with active stream",
			initialMaxBytes:      0,
			initialMaxSkip:       0,
			initialMaxFrames:     0,
			initialMeasureCount:  0,
			outputActive:         true,
			newBytes:             1000.0,
			newSkip:              5.0,
			newFrames:            100.0,
			expectedMaxBytes:     1000.0,
			expectedMaxSkip:      5.0,
			expectedMaxFrames:    100.0,
			expectedMeasureCount: 1,
			expectedActive:       true,
		},
		{
			name:                 "new max values",
			initialMaxBytes:      1000.0,
			initialMaxSkip:       10.0,
			initialMaxFrames:     100.0,
			initialMeasureCount:  1,
			outputActive:         true,
			newBytes:             2000.0,
			newSkip:              20.0,
			newFrames:            200.0,
			expectedMaxBytes:     2000.0,
			expectedMaxSkip:      20.0,
			expectedMaxFrames:    200.0,
			expectedMeasureCount: 2,
			expectedActive:       true,
		},
		{
			name:                 "values lower than max",
			initialMaxBytes:      5000.0,
			initialMaxSkip:       50.0,
			initialMaxFrames:     500.0,
			initialMeasureCount:  5,
			outputActive:         true,
			newBytes:             3000.0,
			newSkip:              30.0,
			newFrames:            300.0,
			expectedMaxBytes:     5000.0,
			expectedMaxSkip:      50.0,
			expectedMaxFrames:    500.0,
			expectedMeasureCount: 6,
			expectedActive:       true,
		},
		{
			name:                 "inactive stream",
			initialMaxBytes:      2000.0,
			initialMaxSkip:       20.0,
			initialMaxFrames:     200.0,
			initialMeasureCount:  2,
			outputActive:         false,
			newBytes:             2100.0,
			newSkip:              21.0,
			newFrames:            210.0,
			expectedMaxBytes:     2100.0,
			expectedMaxSkip:      21.0,
			expectedMaxFrames:    210.0,
			expectedMeasureCount: 3,
			expectedActive:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StreamMetrics{
				maxOutputBytes:   tt.initialMaxBytes,
				maxSkippedFrames: tt.initialMaxSkip,
				maxTotalFrames:   tt.initialMaxFrames,
				measurementCount: tt.initialMeasureCount,
			}

			sm.updateMetrics(tt.outputActive, tt.newBytes, tt.newSkip, tt.newFrames)

			if sm.maxOutputBytes != tt.expectedMaxBytes {
				t.Errorf("Expected maxOutputBytes %f, got %f", tt.expectedMaxBytes, sm.maxOutputBytes)
			}
			if sm.maxSkippedFrames != tt.expectedMaxSkip {
				t.Errorf("Expected maxSkippedFrames %f, got %f", tt.expectedMaxSkip, sm.maxSkippedFrames)
			}
			if sm.maxTotalFrames != tt.expectedMaxFrames {
				t.Errorf("Expected maxTotalFrames %f, got %f", tt.expectedMaxFrames, sm.maxTotalFrames)
			}
			if sm.measurementCount != tt.expectedMeasureCount {
				t.Errorf("Expected measurementCount %d, got %d", tt.expectedMeasureCount, sm.measurementCount)
			}
			if sm.lastActive != tt.expectedActive {
				t.Errorf("Expected lastActive %v, got %v", tt.expectedActive, sm.lastActive)
			}
		})
	}
}

func TestStreamMetrics_RecordError(t *testing.T) {
	sm := &StreamMetrics{}

	testError := fmt.Errorf("connection error")
	sm.recordError(testError)

	sm.mu.Lock()
	if sm.lastError != testError {
		t.Errorf("Expected lastError to be set to test error")
	}
	sm.mu.Unlock()

	data := sm.GetAndResetMaxValues()
	if data.Error != testError {
		t.Error("Expected error to be returned in GetAndResetMaxValues")
	}
}
