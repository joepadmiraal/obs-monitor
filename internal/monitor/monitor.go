package monitor

import (
	"fmt"
	"time"

	"github.com/andreykaipov/goobs"
)

type ObsConnectionInfo struct {
	Password string
	Host     string
}

type Monitor struct {
	client         *goobs.Client
	connectionInfo ObsConnectionInfo
}

// NewMonitor Connects to OBS and
func NewMonitor(connectionInfo ObsConnectionInfo) (*Monitor, error) {

	return &Monitor{
		connectionInfo: connectionInfo,
	}, nil
}

// Monitor continuously monitors and prints stream status
func (m *Monitor) Start() error {

	var err error
	m.client, err = goobs.New(m.connectionInfo.Host, goobs.WithPassword(m.connectionInfo.Password))
	if err != nil {
		return err
	}

	m.PrintInfo()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status, err := m.client.Stream.GetStreamStatus()
		if err != nil {
			fmt.Printf("Error getting stream status: %v\n", err)
			continue
		}
		fmt.Printf("Active: %t, bytes: %.0f, skipped: %.0f\n", status.OutputActive, status.OutputBytes, status.OutputSkippedFrames)
	}

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
	m.client.Disconnect()
}
