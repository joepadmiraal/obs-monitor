package metric

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestObsStats_GetAndResetMaxValues_ReturnsCorrectMaxValues(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    25.5,
		maxObsMemoryUsage: 1024.0,
		measurementCount:  5,
	}

	data := obs.GetAndResetMaxValues()

	if data.ObsCpuUsage != 25.5 {
		t.Errorf("Expected ObsCpuUsage to be 25.5, got %f", data.ObsCpuUsage)
	}
	if data.ObsMemoryUsage != 1024.0 {
		t.Errorf("Expected ObsMemoryUsage to be 1024.0, got %f", data.ObsMemoryUsage)
	}
}

func TestObsStats_GetAndResetMaxValues_ResetsValues(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    25.5,
		maxObsMemoryUsage: 1024.0,
		measurementCount:  5,
	}

	_ = obs.GetAndResetMaxValues()

	if obs.maxObsCpuUsage != 0 {
		t.Errorf("Expected maxObsCpuUsage to be reset to 0, got %f", obs.maxObsCpuUsage)
	}
	if obs.maxObsMemoryUsage != 0 {
		t.Errorf("Expected maxObsMemoryUsage to be reset to 0, got %f", obs.maxObsMemoryUsage)
	}
	if obs.lastError != nil {
		t.Error("Expected lastError to be reset to nil")
	}
}

func TestObsStats_GetAndResetMaxValues_ErrorHandling(t *testing.T) {
	testError := fmt.Errorf("test error")
	obs := &ObsStats{
		maxObsCpuUsage:   10.0,
		lastError:        testError,
		measurementCount: 3,
	}

	data := obs.GetAndResetMaxValues()

	if data.Error == nil {
		t.Error("Expected error to be returned")
	}
	if data.Error != testError {
		t.Error("Expected error pointer to match")
	}

	if obs.lastError != nil {
		t.Error("Expected lastError to be reset to nil after GetAndResetMaxValues")
	}
}

func TestObsStats_GetAndResetMaxValues_ConcurrentAccess(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    15.0,
		maxObsMemoryUsage: 512.0,
		measurementCount:  10,
	}

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = obs.GetAndResetMaxValues()
		}()
	}

	wg.Wait()
}

func TestObsStats_GetAndResetMaxValues_ZeroValues(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    0,
		maxObsMemoryUsage: 0,
		measurementCount:  0,
	}

	data := obs.GetAndResetMaxValues()

	if data.ObsCpuUsage != 0 {
		t.Errorf("Expected ObsCpuUsage to be 0, got %f", data.ObsCpuUsage)
	}
	if data.ObsMemoryUsage != 0 {
		t.Errorf("Expected ObsMemoryUsage to be 0, got %f", data.ObsMemoryUsage)
	}
	if data.Error != nil {
		t.Error("Expected no error with zero values")
	}
}

func TestObsStats_GetAndResetMaxValues_TracksMaximum(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:   10.0,
		measurementCount: 2,
	}

	data1 := obs.GetAndResetMaxValues()
	if data1.ObsCpuUsage != 10.0 {
		t.Errorf("Expected first call to return 10.0, got %f", data1.ObsCpuUsage)
	}

	obs.maxObsCpuUsage = 20.0
	obs.measurementCount = 3

	data2 := obs.GetAndResetMaxValues()
	if data2.ObsCpuUsage != 20.0 {
		t.Errorf("Expected second call to return 20.0 (new max), got %f", data2.ObsCpuUsage)
	}
}

func TestObsStats_NewObsStats(t *testing.T) {
	interval := 100 * time.Millisecond

	obs, err := NewObsStats(nil, interval)

	if err != nil {
		t.Fatalf("NewObsStats returned error: %v", err)
	}
	if obs == nil {
		t.Fatal("NewObsStats returned nil")
	}
	if obs.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, obs.interval)
	}
}

func TestObsStats_GetAndResetMaxValues_SetsTimestamp(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage: 15.0,
	}

	before := time.Now()
	data := obs.GetAndResetMaxValues()
	after := time.Now()

	if data.Timestamp.Before(before) || data.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}

func TestObsStats_ErrorHandlingDuringCollection(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:   20.0,
		measurementCount: 5,
	}

	testError := fmt.Errorf("stats collection error")

	obs.mu.Lock()
	obs.lastError = testError
	obs.mu.Unlock()

	obs.mu.Lock()
	if obs.lastError != testError {
		t.Error("Expected lastError to be set")
	}
	if obs.maxObsCpuUsage != 20.0 {
		t.Error("Expected maxObsCpuUsage to remain unchanged after error")
	}
	obs.mu.Unlock()

	data := obs.GetAndResetMaxValues()
	if data.Error != testError {
		t.Error("Expected error to be returned in data")
	}
}

func TestObsStats_MeasurementCountIncrement(t *testing.T) {
	obs := &ObsStats{
		measurementCount: 5,
	}

	obs.mu.Lock()
	obs.maxObsCpuUsage = 10.0
	obs.measurementCount++
	obs.mu.Unlock()

	if obs.measurementCount != 6 {
		t.Errorf("Expected measurementCount to be 6, got %d", obs.measurementCount)
	}
}

func TestObsStats_UpdateStats(t *testing.T) {
	tests := []struct {
		name                 string
		initialMaxCpu        float64
		initialMaxMem        float64
		initialMeasureCount  int
		newCpu               float64
		newMem               float64
		expectedMaxCpu       float64
		expectedMaxMem       float64
		expectedMeasureCount int
	}{
		{
			name:                 "first measurement",
			initialMaxCpu:        0,
			initialMaxMem:        0,
			initialMeasureCount:  0,
			newCpu:               25.5,
			newMem:               512.0,
			expectedMaxCpu:       25.5,
			expectedMaxMem:       512.0,
			expectedMeasureCount: 1,
		},
		{
			name:                 "new max values",
			initialMaxCpu:        20.0,
			initialMaxMem:        400.0,
			initialMeasureCount:  3,
			newCpu:               35.5,
			newMem:               600.0,
			expectedMaxCpu:       35.5,
			expectedMaxMem:       600.0,
			expectedMeasureCount: 4,
		},
		{
			name:                 "values lower than max",
			initialMaxCpu:        50.0,
			initialMaxMem:        800.0,
			initialMeasureCount:  5,
			newCpu:               30.0,
			newMem:               500.0,
			expectedMaxCpu:       50.0,
			expectedMaxMem:       800.0,
			expectedMeasureCount: 6,
		},
		{
			name:                 "mixed higher and lower",
			initialMaxCpu:        40.0,
			initialMaxMem:        700.0,
			initialMeasureCount:  2,
			newCpu:               45.0,
			newMem:               600.0,
			expectedMaxCpu:       45.0,
			expectedMaxMem:       700.0,
			expectedMeasureCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := &ObsStats{
				maxObsCpuUsage:    tt.initialMaxCpu,
				maxObsMemoryUsage: tt.initialMaxMem,
				measurementCount:  tt.initialMeasureCount,
			}

			obs.updateStats(tt.newCpu, tt.newMem)

			if obs.maxObsCpuUsage != tt.expectedMaxCpu {
				t.Errorf("Expected maxObsCpuUsage %f, got %f", tt.expectedMaxCpu, obs.maxObsCpuUsage)
			}
			if obs.maxObsMemoryUsage != tt.expectedMaxMem {
				t.Errorf("Expected maxObsMemoryUsage %f, got %f", tt.expectedMaxMem, obs.maxObsMemoryUsage)
			}
			if obs.measurementCount != tt.expectedMeasureCount {
				t.Errorf("Expected measurementCount %d, got %d", tt.expectedMeasureCount, obs.measurementCount)
			}
		})
	}
}

func TestObsStats_RecordError(t *testing.T) {
	obs := &ObsStats{}

	testError := fmt.Errorf("stats fetch error")
	obs.recordError(testError)

	obs.mu.Lock()
	if obs.lastError != testError {
		t.Errorf("Expected lastError to be set to test error")
	}
	obs.mu.Unlock()

	data := obs.GetAndResetMaxValues()
	if data.Error != testError {
		t.Error("Expected error to be returned in GetAndResetMaxValues")
	}
}
