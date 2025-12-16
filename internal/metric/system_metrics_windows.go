//go:build windows

package metric

import (
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type SystemMetrics struct {
	maxCpuUsage    float64
	maxMemoryUsage float64
	lastError      error
	mu             sync.Mutex
	interval       time.Duration
	lastIdleTime   uint64
	lastKernelTime uint64
	lastUserTime   uint64
}

type SystemMetricsData struct {
	Timestamp   time.Time
	CpuUsage    float64
	MemoryUsage float64
	Error       error
}

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
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
	s.lastIdleTime, s.lastKernelTime, s.lastUserTime, _ = getSystemTimes()

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
	idle, kernel, user, err := getSystemTimes()
	if err != nil {
		return 0, err
	}

	idleDiff := idle - s.lastIdleTime
	kernelDiff := kernel - s.lastKernelTime
	userDiff := user - s.lastUserTime

	totalDiff := kernelDiff + userDiff

	s.lastIdleTime = idle
	s.lastKernelTime = kernel
	s.lastUserTime = user

	if totalDiff == 0 {
		return 0, nil
	}

	cpuUsage := (float64(totalDiff-idleDiff) / float64(totalDiff)) * 100.0
	return cpuUsage, nil
}

func getSystemTimes() (idle, kernel, user uint64, err error) {
	var idleTime, kernelTime, userTime syscall.Filetime

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getSystemTimes := kernel32.NewProc("GetSystemTimes")

	ret, _, err := getSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)

	if ret == 0 {
		return 0, 0, 0, err
	}

	idle = uint64(idleTime.HighDateTime)<<32 | uint64(idleTime.LowDateTime)
	kernel = uint64(kernelTime.HighDateTime)<<32 | uint64(kernelTime.LowDateTime)
	user = uint64(userTime.HighDateTime)<<32 | uint64(userTime.LowDateTime)

	return idle, kernel, user, nil
}

func (s *SystemMetrics) getMemoryUsage() (float64, error) {
	var memStatus memoryStatusEx
	memStatus.dwLength = uint32(unsafe.Sizeof(memStatus))

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		return 0, fmt.Errorf("GlobalMemoryStatusEx failed: %v", err)
	}

	memUsage := float64(memStatus.dwMemoryLoad)
	return memUsage, nil
}
