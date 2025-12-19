package monitor

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/joepadmiraal/obs-monitor/internal/metric"
	"github.com/joepadmiraal/obs-monitor/internal/writer"
)

type ObsConnectionInfo struct {
	Password       string
	Host           string
	CSVFile        string
	MetricInterval int
	WriterInterval int
}

type Monitor struct {
	client         *goobs.Client
	connectionInfo ObsConnectionInfo
	obsPinger      *metric.Pinger
	googlePinger   *metric.Pinger
	streamMetrics  *metric.StreamMetrics
	obsStats       *metric.ObsStats
	systemMetrics  *metric.SystemMetrics
	csvWriter      *writer.CSVWriter
	consoleWriter  *writer.ConsoleWriter
	metricInterval time.Duration
	writerInterval time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	shutdownDone   chan struct{}
}

// NewMonitor Connects to OBS and
func NewMonitor(connectionInfo ObsConnectionInfo) (*Monitor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		connectionInfo: connectionInfo,
		metricInterval: time.Duration(connectionInfo.MetricInterval) * time.Millisecond,
		writerInterval: time.Duration(connectionInfo.WriterInterval) * time.Millisecond,
		ctx:            ctx,
		cancel:         cancel,
		shutdownDone:   make(chan struct{}),
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

	// Get OBS stream server domain
	streamSettings, err := m.client.Config.GetStreamServiceSettings()
	if err != nil {
		return fmt.Errorf("failed to get stream settings: %w", err)
	}

	serverURL := streamSettings.StreamServiceSettings.Server
	if serverURL == "" {
		return fmt.Errorf("stream server URL not found in settings")
	}

	obsDomain, err := extractDomain(serverURL)
	if err != nil {
		return fmt.Errorf("failed to extract domain from URL: %w", err)
	}

	if err := m.initializePingers(obsDomain); err != nil {
		return err
	}

	// Initialize stream metrics
	m.streamMetrics, err = metric.NewStreamMetrics(m.client, m.metricInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize stream metrics: %w", err)
	}

	// Initialize OBS stats
	m.obsStats, err = metric.NewObsStats(m.client, m.metricInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize OBS stats: %w", err)
	}

	// Initialize system metrics
	m.systemMetrics, err = metric.NewSystemMetrics(m.metricInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize system metrics: %w", err)
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

	// Start stream metrics monitoring in a goroutine
	go func() {
		if err := m.streamMetrics.Start(); err != nil {
			fmt.Printf("Stream metrics error: %v\n", err)
		}
	}()

	// Start OBS stats monitoring in a goroutine
	go func() {
		if err := m.obsStats.Start(); err != nil {
			fmt.Printf("OBS stats error: %v\n", err)
		}
	}()

	// Start system metrics monitoring in a goroutine
	go func() {
		if err := m.systemMetrics.Start(); err != nil {
			fmt.Printf("System metrics error: %v\n", err)
		}
	}()

	// Start metrics collector
	go m.collectAndWriteMetrics()

	// Monitor for disconnection and OBS exit
	go m.monitorConnection()

	return nil
}

func (m *Monitor) initializePingers(obsDomain string) error {
	var err error

	m.obsPinger, err = metric.NewPinger(obsDomain, m.metricInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize OBS pinger: %w", err)
	}

	m.googlePinger, err = metric.NewPinger("google.com", m.metricInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize Google pinger: %w", err)
	}

	go func() {
		if err := m.obsPinger.Start(); err != nil {
			fmt.Printf("OBS pinger error: %v\n", err)
		}
	}()

	go func() {
		if err := m.googlePinger.Start(); err != nil {
			fmt.Printf("Google pinger error: %v\n", err)
		}
	}()

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
	fmt.Printf("Client library version: %s\n\n", goobs.LibraryVersion)
}

func (m *Monitor) Close() {
	if m.csvWriter != nil {
		if err := m.csvWriter.Close(); err != nil {
			fmt.Printf("Error closing CSV writer: %v\n", err)
		}
	}
	if m.client != nil {
		m.client.Disconnect()
	}
}

func (m *Monitor) Shutdown() {
	m.cancel()
}

func (m *Monitor) Done() <-chan struct{} {
	return m.shutdownDone
}

// collectAndWriteMetrics collects metrics from pingers, stream metrics, and system metrics and writes to CSV
func (m *Monitor) collectAndWriteMetrics() {
	ticker := time.NewTicker(m.writerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			obsRTT, obsErr := m.obsPinger.GetAndResetMaxRTT()
			googleRTT, googleErr := m.googlePinger.GetAndResetMaxRTT()
			streamData := m.streamMetrics.GetAndResetMaxValues()
			obsStatsData := m.obsStats.GetAndResetMaxValues()
			systemMetricsData := m.systemMetrics.GetAndResetMaxValues()

			m.writeMetrics(obsRTT, obsErr, googleRTT, googleErr, streamData, obsStatsData, systemMetricsData)
		}
	}
}

// writeMetrics writes a combined metrics row to CSV and console
func (m *Monitor) writeMetrics(obsRTT time.Duration, obsErr error, googleRTT time.Duration, googleErr error, streamData metric.StreamMetricsData, obsStatsData metric.ObsStatsData, systemMetricsData metric.SystemMetricsData) {
	data := writer.MetricsData{
		Timestamp:           streamData.Timestamp,
		ObsRTT:              obsRTT,
		ObsPingError:        obsErr,
		GoogleRTT:           googleRTT,
		GooglePingError:     googleErr,
		StreamActive:        streamData.Active,
		OutputBytes:         streamData.OutputBytes,
		OutputSkippedFrames: streamData.OutputSkippedFrames,
		StreamError:         streamData.Error,
		ObsCpuUsage:         obsStatsData.ObsCpuUsage,
		ObsMemoryUsage:      obsStatsData.ObsMemoryUsage,
		ObsStatsError:       obsStatsData.Error,
		SystemCpuUsage:      systemMetricsData.CpuUsage,
		SystemMemoryUsage:   systemMetricsData.MemoryUsage,
		SystemMetricsError:  systemMetricsData.Error,
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

func (m *Monitor) monitorConnection() {
	defer close(m.shutdownDone)

	listenDone := make(chan struct{})
	go func() {
		defer close(listenDone)
		m.client.Listen(func(event any) {
			switch event.(type) {
			case *events.ExitStarted:
				fmt.Println("\nOBS connection lost, closing obs-monitor...")
				m.cancel()
			}
		})
	}()

	select {
	case <-m.ctx.Done():
		m.client.Disconnect()
		<-listenDone
	case <-listenDone:
	}
}

func extractDomain(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "rtmp://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("no hostname found in URL")
	}

	return host, nil
}
