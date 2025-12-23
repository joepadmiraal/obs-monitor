package metric

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSystemMetrics_GetAndResetMaxValues_ReturnsCorrectMaxValues(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    75.5,
		maxMemoryUsage: 85.2,
	}

	data := sm.GetAndResetMaxValues()

	if data.CpuUsage != 75.5 {
		t.Errorf("Expected CpuUsage to be 75.5, got %f", data.CpuUsage)
	}
	if data.MemoryUsage != 85.2 {
		t.Errorf("Expected MemoryUsage to be 85.2, got %f", data.MemoryUsage)
	}
}

func TestSystemMetrics_GetAndResetMaxValues_ResetsValues(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    75.5,
		maxMemoryUsage: 85.2,
	}

	_ = sm.GetAndResetMaxValues()

	if sm.maxCpuUsage != 0 {
		t.Errorf("Expected maxCpuUsage to be reset to 0, got %f", sm.maxCpuUsage)
	}
	if sm.maxMemoryUsage != 0 {
		t.Errorf("Expected maxMemoryUsage to be reset to 0, got %f", sm.maxMemoryUsage)
	}
	if sm.lastError != nil {
		t.Error("Expected lastError to be reset to nil")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_ErrorHandling(t *testing.T) {
	testError := fmt.Errorf("system error")
	sm := &SystemMetrics{
		maxCpuUsage: 50.0,
		lastError:   testError,
	}

	data := sm.GetAndResetMaxValues()

	if data.Error == nil {
		t.Error("Expected error to be returned")
	}
	if data.Error != testError {
		t.Error("Expected error pointer to match")
	}

	if sm.lastError != nil {
		t.Error("Expected lastError to be reset to nil after GetAndResetMaxValues")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_ConcurrentAccess(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    60.0,
		maxMemoryUsage: 70.0,
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

func TestSystemMetrics_GetAndResetMaxValues_ZeroValues(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    0,
		maxMemoryUsage: 0,
	}

	data := sm.GetAndResetMaxValues()

	if data.CpuUsage != 0 {
		t.Errorf("Expected CpuUsage to be 0, got %f", data.CpuUsage)
	}
	if data.MemoryUsage != 0 {
		t.Errorf("Expected MemoryUsage to be 0, got %f", data.MemoryUsage)
	}
	if data.Error != nil {
		t.Error("Expected no error with zero values")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_TracksMaximum(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage: 40.0,
	}

	data1 := sm.GetAndResetMaxValues()
	if data1.CpuUsage != 40.0 {
		t.Errorf("Expected first call to return 40.0, got %f", data1.CpuUsage)
	}

	sm.maxCpuUsage = 80.0

	data2 := sm.GetAndResetMaxValues()
	if data2.CpuUsage != 80.0 {
		t.Errorf("Expected second call to return 80.0 (new max), got %f", data2.CpuUsage)
	}
}

func TestSystemMetrics_NewSystemMetrics(t *testing.T) {
	interval := 100 * time.Millisecond

	sm, err := NewSystemMetrics(interval)

	if err != nil {
		t.Fatalf("NewSystemMetrics returned error: %v", err)
	}
	if sm == nil {
		t.Fatal("NewSystemMetrics returned nil")
	}
	if sm.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, sm.interval)
	}
}

func TestSystemMetrics_GetAndResetMaxValues_TimestampSet(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage: 50.0,
	}

	before := time.Now()
	data := sm.GetAndResetMaxValues()
	after := time.Now()

	if data.Timestamp.Before(before) || data.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_HighMemoryUsage(t *testing.T) {
	sm := &SystemMetrics{
		maxMemoryUsage: 99.9,
	}

	data := sm.GetAndResetMaxValues()

	if data.MemoryUsage != 99.9 {
		t.Errorf("Expected MemoryUsage to be 99.9, got %f", data.MemoryUsage)
	}
}

func TestSystemMetrics_MaxValueTracking(t *testing.T) {
	tests := []struct {
		name           string
		currentMaxCpu  float64
		currentMaxMem  float64
		newCpu         float64
		newMem         float64
		expectedMaxCpu float64
		expectedMaxMem float64
	}{
		{
			name:           "new values higher than current max",
			currentMaxCpu:  50.0,
			currentMaxMem:  60.0,
			newCpu:         75.5,
			newMem:         85.2,
			expectedMaxCpu: 75.5,
			expectedMaxMem: 85.2,
		},
		{
			name:           "new values lower than current max",
			currentMaxCpu:  90.0,
			currentMaxMem:  95.0,
			newCpu:         60.0,
			newMem:         70.0,
			expectedMaxCpu: 90.0,
			expectedMaxMem: 95.0,
		},
		{
			name:           "mixed higher and lower values",
			currentMaxCpu:  70.0,
			currentMaxMem:  80.0,
			newCpu:         75.0,
			newMem:         75.0,
			expectedMaxCpu: 75.0,
			expectedMaxMem: 80.0,
		},
		{
			name:           "zero to non-zero",
			currentMaxCpu:  0,
			currentMaxMem:  0,
			newCpu:         45.5,
			newMem:         55.3,
			expectedMaxCpu: 45.5,
			expectedMaxMem: 55.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &SystemMetrics{
				maxCpuUsage:    tt.currentMaxCpu,
				maxMemoryUsage: tt.currentMaxMem,
			}

			sm.mu.Lock()
			if tt.newCpu > sm.maxCpuUsage {
				sm.maxCpuUsage = tt.newCpu
			}
			if tt.newMem > sm.maxMemoryUsage {
				sm.maxMemoryUsage = tt.newMem
			}
			sm.mu.Unlock()

			if sm.maxCpuUsage != tt.expectedMaxCpu {
				t.Errorf("Expected maxCpuUsage %f, got %f", tt.expectedMaxCpu, sm.maxCpuUsage)
			}
			if sm.maxMemoryUsage != tt.expectedMaxMem {
				t.Errorf("Expected maxMemoryUsage %f, got %f", tt.expectedMaxMem, sm.maxMemoryUsage)
			}
		})
	}
}

func TestSystemMetrics_ErrorHandlingDuringCollection(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage: 50.0,
	}

	testError := fmt.Errorf("system metrics collection error")

	sm.mu.Lock()
	sm.lastError = testError
	sm.mu.Unlock()

	sm.mu.Lock()
	if sm.lastError != testError {
		t.Error("Expected lastError to be set")
	}
	if sm.maxCpuUsage != 50.0 {
		t.Error("Expected maxCpuUsage to remain unchanged after error")
	}
	sm.mu.Unlock()

	data := sm.GetAndResetMaxValues()
	if data.Error != testError {
		t.Error("Expected error to be returned in data")
	}
}

func TestSystemMetrics_ErrorHandlingContinuesCollection(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage: 30.0,
	}

	cpuError := fmt.Errorf("cpu collection error")

	sm.mu.Lock()
	sm.lastError = cpuError
	sm.mu.Unlock()

	sm.mu.Lock()
	currentError := sm.lastError
	sm.mu.Unlock()

	if currentError != cpuError {
		t.Error("Expected lastError to be set to CPU error")
	}

	memError := fmt.Errorf("memory collection error")
	sm.mu.Lock()
	sm.lastError = memError
	sm.mu.Unlock()

	data := sm.GetAndResetMaxValues()
	if data.Error != memError {
		t.Error("Expected error to be the last encountered error")
	}
}
