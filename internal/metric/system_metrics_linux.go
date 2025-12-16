//go:build linux

package metric

import (
	"bufio"
	"fmt"
	"os"
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
	lastCpuStats   cpuStats
}

type cpuStats struct {
	user    uint64
	nice    uint64
	system  uint64
	idle    uint64
	iowait  uint64
	irq     uint64
	softirq uint64
	steal   uint64
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
	s.lastCpuStats, _ = readCpuStats()

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
	current, err := readCpuStats()
	if err != nil {
		return 0, err
	}

	prevIdle := s.lastCpuStats.idle + s.lastCpuStats.iowait
	currIdle := current.idle + current.iowait

	prevTotal := s.lastCpuStats.user + s.lastCpuStats.nice + s.lastCpuStats.system +
		s.lastCpuStats.idle + s.lastCpuStats.iowait + s.lastCpuStats.irq +
		s.lastCpuStats.softirq + s.lastCpuStats.steal
	currTotal := current.user + current.nice + current.system + current.idle +
		current.iowait + current.irq + current.softirq + current.steal

	totalDiff := currTotal - prevTotal
	idleDiff := currIdle - prevIdle

	s.lastCpuStats = current

	if totalDiff == 0 {
		return 0, nil
	}

	cpuUsage := (float64(totalDiff-idleDiff) / float64(totalDiff)) * 100.0
	return cpuUsage, nil
}

func readCpuStats() (cpuStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStats{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return cpuStats{}, fmt.Errorf("invalid cpu stats format")
			}

			return cpuStats{
				user:    parseUint64(fields[1]),
				nice:    parseUint64(fields[2]),
				system:  parseUint64(fields[3]),
				idle:    parseUint64(fields[4]),
				iowait:  parseUint64(fields[5]),
				irq:     parseUint64(fields[6]),
				softirq: parseUint64(fields[7]),
				steal:   parseUint64(fields[8]),
			}, nil
		}
	}

	return cpuStats{}, fmt.Errorf("cpu stats not found")
}

func (s *SystemMetrics) getMemoryUsage() (float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var memTotal, memAvailable uint64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotal = parseUint64(fields[1])
		case "MemAvailable:":
			memAvailable = parseUint64(fields[1])
		}
	}

	if memTotal == 0 {
		return 0, fmt.Errorf("failed to read memory info")
	}

	memUsed := memTotal - memAvailable
	memUsage := (float64(memUsed) / float64(memTotal)) * 100.0

	return memUsage, nil
}

func parseUint64(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}
