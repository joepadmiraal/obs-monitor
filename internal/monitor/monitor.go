package monitor

import (
	"fmt"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/joepadmiraal/obs-monitor/internal/metric"
	"github.com/joepadmiraal/obs-monitor/internal/writer"
)

type ObsConnectionInfo struct {
	Password string
	Host     string
	CSVFile  string
}

type Monitor struct {
	client         *goobs.Client
	connectionInfo ObsConnectionInfo
	pinger         *metric.Pinger
	streamMetrics  *metric.StreamMetrics
	csvWriter      *writer.CSVWriter
	consoleWriter  *writer.ConsoleWriter
}

// NewMonitor Connects to OBS and
func NewMonitor(connectionInfo ObsConnectionInfo) (*Monitor, error) {

	return &Monitor{
		connectionInfo: connectionInfo,
	}, nil
}

// connect establishes a connection to OBS (internal use only)
func (m *Monitor) connect() error {
	var err error
	m.client, err = goobs.New(m.connectionInfo.Host, goobs.WithPassword(m.connectionInfo.Password))
	if err != nil {
		return err
	}
	return nil
}

// Start connects to OBS and starts all monitoring components
func (m *Monitor) Start() error {
	// Connect to OBS
	if err := m.connect(); err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	// Initialize pinger
	var err error
	m.pinger, err = metric.NewPinger(m.client)
	if err != nil {
		return fmt.Errorf("failed to initialize pinger: %w", err)
	}

	// Initialize stream metrics
	m.streamMetrics, err = metric.NewStreamMetrics(m.client)
	if err != nil {
		return fmt.Errorf("failed to initialize stream metrics: %w", err)
	}

	// Initialize CSV writer if filename is provided
	if m.connectionInfo.CSVFile != "" {
		m.csvWriter, err = writer.NewCSVWriter(m.connectionInfo.CSVFile)
		if err != nil {
			return fmt.Errorf("failed to initialize CSV writer: %w", err)
		}
		fmt.Printf("Writing metrics to CSV file: %s\n", m.connectionInfo.CSVFile)
	}

	// Initialize console writer
	m.consoleWriter = writer.NewConsoleWriter()

	m.PrintInfo()

	// Start pinger in a goroutine
	go func() {
		if err := m.pinger.Start(); err != nil {
			fmt.Printf("Pinger error: %v\n", err)
		}
	}()

	// Start stream metrics monitoring in a goroutine
	go func() {
		if err := m.streamMetrics.Start(); err != nil {
			fmt.Printf("Stream metrics error: %v\n", err)
		}
	}()

	// Start metrics collector
	go m.collectAndWriteMetrics()

	return nil
}

func (m *Monitor) PrintInfo() {
	version, err := m.client.General.GetVersion()
	if err != nil {
		panic(err)
	}

	fmt.Printf("OBS Studio version: %s\n", version.ObsVersion)
	fmt.Printf("Server protocol version: %s\n", version.ObsWebSocketVersion)
	fmt.Printf("Client protocol version: %s\n", goobs.ProtocolVersion)
	fmt.Printf("Client library version: %s\n", goobs.LibraryVersion)
}

func (m *Monitor) Close() {
	if m.csvWriter != nil {
		if err := m.csvWriter.Close(); err != nil {
			fmt.Printf("Error closing CSV writer: %v\n", err)
		}
	}
	m.client.Disconnect()
}

// collectAndWriteMetrics collects metrics from both pinger and stream metrics and writes to CSV
func (m *Monitor) collectAndWriteMetrics() {
	var lastPingMetrics metric.PingMetrics
	var lastStreamMetrics metric.StreamMetricsData
	var havePing, haveStream bool

	pingChan := m.pinger.GetMetricsChan()
	streamChan := m.streamMetrics.GetMetricsChan()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case pingMetrics := <-pingChan:
			lastPingMetrics = pingMetrics
			havePing = true

		case streamMetrics := <-streamChan:
			lastStreamMetrics = streamMetrics
			haveStream = true

		case <-ticker.C:
			// Write metrics once per second with the latest data from both sources
			if havePing || haveStream {
				m.writeMetrics(lastPingMetrics, lastStreamMetrics)
			}
		}
	}
}

// writeMetrics writes a combined metrics row to CSV and console
func (m *Monitor) writeMetrics(pingMetrics metric.PingMetrics, streamMetrics metric.StreamMetricsData) {
	// Use the more recent timestamp
	timestamp := pingMetrics.Timestamp
	if streamMetrics.Timestamp.After(pingMetrics.Timestamp) {
		timestamp = streamMetrics.Timestamp
	}

	data := writer.MetricsData{
		Timestamp:           timestamp,
		RTT:                 pingMetrics.RTT,
		PingError:           pingMetrics.Error,
		StreamActive:        streamMetrics.Active,
		OutputBytes:         streamMetrics.OutputBytes,
		OutputSkippedFrames: streamMetrics.OutputSkippedFrames,
		StreamError:         streamMetrics.Error,
	}

	// Write to CSV if enabled
	if m.csvWriter != nil {
		if err := m.csvWriter.WriteMetrics(data); err != nil {
			fmt.Printf("Error writing to CSV: %v\n", err)
		}
	}

	// Write to console
	if err := m.consoleWriter.WriteMetrics(data); err != nil {
		fmt.Printf("Error writing to console: %v\n", err)
	}
}
