//go:build darwin

package metric

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
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

func (s *SystemMetrics) Start() error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		cpuUsage, err := s.getCpuUsage()
		if err != nil {
			s.mu.Lock()
			s.lastError = err
			s.mu.Unlock()
			continue
		}

		memUsage, err := s.getMemoryUsage()
		if err != nil {
			s.mu.Lock()
			s.lastError = err
			s.mu.Unlock()
			continue
		}

		s.mu.Lock()
		if cpuUsage > s.maxCpuUsage {
			s.maxCpuUsage = cpuUsage
		}
		if memUsage > s.maxMemoryUsage {
			s.maxMemoryUsage = memUsage
		}
		s.mu.Unlock()
	}

	return nil
}

func (s *SystemMetrics) getCpuUsage() (float64, error) {
	cmd := exec.Command("ps", "-A", "-o", "%cpu")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	var totalCpu float64
	lines := strings.Split(out.String(), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cpu, err := strconv.ParseFloat(line, 64)
		if err == nil {
			totalCpu += cpu
		}
	}

	return totalCpu, nil
}

func (s *SystemMetrics) getMemoryUsage() (float64, error) {
	cmd := exec.Command("vm_stat")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	var pagesFree, pagesActive, pagesInactive, pagesSpeculative, pagesWired uint64
	pageSize := uint64(4096)

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		valueStr := strings.TrimSuffix(fields[len(fields)-1], ".")
		value, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			continue
		}

		switch {
		case strings.HasPrefix(line, "Pages free:"):
			pagesFree = value
		case strings.HasPrefix(line, "Pages active:"):
			pagesActive = value
		case strings.HasPrefix(line, "Pages inactive:"):
			pagesInactive = value
		case strings.HasPrefix(line, "Pages speculative:"):
			pagesSpeculative = value
		case strings.HasPrefix(line, "Pages wired down:"):
			pagesWired = value
		}
	}

	totalPages := pagesFree + pagesActive + pagesInactive + pagesSpeculative + pagesWired
	usedPages := pagesActive + pagesWired

	if totalPages == 0 {
		return 0, fmt.Errorf("failed to calculate memory usage")
	}

	memUsage := (float64(usedPages*pageSize) / float64(totalPages*pageSize)) * 100.0

	return memUsage, nil
}
