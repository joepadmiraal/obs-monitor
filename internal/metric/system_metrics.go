package metric

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMetrics struct {
	maxCpuUsage    float64
	maxMemoryUsage float64
	lastError      error
	mu             sync.Mutex
	interval       time.Duration
}

type SystemMetricsData struct {
	Timestamp   time.Time
	CpuUsage    float64
	MemoryUsage float64
	Error       error
}

func NewSystemMetrics(interval time.Duration) (*SystemMetrics, error) {
	return &SystemMetrics{
		interval: interval,
	}, nil
}

func (s *SystemMetrics) GetAndResetMaxValues() SystemMetricsData {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxCpu := s.maxCpuUsage
	maxMemory := s.maxMemoryUsage
	err := s.lastError

	s.maxCpuUsage = 0
	s.maxMemoryUsage = 0
	s.lastError = nil

	return SystemMetricsData{
		Timestamp:   time.Now(),
		CpuUsage:    maxCpu,
		MemoryUsage: maxMemory,
		Error:       err,
	}
}

func (s *SystemMetrics) updateMetrics(cpuUsage, memUsage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cpuUsage > s.maxCpuUsage {
		s.maxCpuUsage = cpuUsage
	}
	if memUsage > s.maxMemoryUsage {
		s.maxMemoryUsage = memUsage
	}
}

func (s *SystemMetrics) recordError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = err
}

func (s *SystemMetrics) Start() error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		cpuUsage, err := s.getCpuUsage()
		if err != nil {
			s.recordError(err)
			continue
		}

		memUsage, err := s.getMemoryUsage()
		if err != nil {
			s.recordError(err)
			continue
		}

		s.updateMetrics(cpuUsage, memUsage)
	}

	return nil
}

func (s *SystemMetrics) getCpuUsage() (float64, error) {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}

	if len(percentages) == 0 {
		return 0, nil
	}

	return percentages[0], nil
}

func (s *SystemMetrics) getMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return vmStat.UsedPercent, nil
}
