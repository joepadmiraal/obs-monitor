package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joepadmiraal/obs-monitor/internal/monitor"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	versionFlag := flag.Bool("version", false, "Show version information")
	password := flag.String("password", "", "OBS WebSocket password")
	host := flag.String("host", "localhost", "OBS WebSocket host")
	port := flag.String("port", "4455", "OBS WebSocket port")
	csvFile := flag.String("csv", "obs-monitor.csv", "Optional CSV file to write metrics to")
	metricIntervalMs := flag.Int("metric-interval", 1000, "Metric collection interval in milliseconds (default 1000ms)")
	writerIntervalMs := flag.Int("writer-interval", 1000, "Writer interval in milliseconds (default 1000ms)")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("obs-monitor %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	if *password == "" {
		fmt.Print("Enter OBS WebSocket password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}
		*password = string(passwordBytes)
	}

	if *metricIntervalMs > *writerIntervalMs {
		fmt.Printf("Error: metric interval (%dms) cannot be higher than writer interval (%dms)\n", *metricIntervalMs, *writerIntervalMs)
		os.Exit(1)
	}

	monitor, err := monitor.NewMonitor(monitor.ObsConnectionInfo{
		Host:           fmt.Sprintf("%s:%s", *host, *port),
		Password:       *password,
		CSVFile:        *csvFile,
		MetricInterval: *metricIntervalMs,
		WriterInterval: *writerIntervalMs,
	})
	if err != nil {
		panic(err)
	}
	defer monitor.Close()

	fmt.Println("\nPress Ctrl-C to exit")

	err = monitor.Start()
	if err != nil {
		fmt.Printf("Monitor error: %v\n", err)
		os.Exit(1)
	}

	waitForExit(monitor)
}

func waitForExit(mon *monitor.Monitor) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal, shutting down...")
		mon.Shutdown()
		<-mon.Done()
	case <-mon.Done():
	}
}
