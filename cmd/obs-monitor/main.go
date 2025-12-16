package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joepadmiraal/obs-monitor/internal/monitor"
)

func main() {
	password := flag.String("password", "", "OBS WebSocket password")
	host := flag.String("host", "localhost", "OBS WebSocket host")
	port := flag.String("port", "4455", "OBS WebSocket port")
	csvFile := flag.String("csv", "", "Optional CSV file to write metrics to")
	flag.Parse()

	if *password == "" {
		fmt.Println("Usage: obs-monitor -password <password>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	monitor, err := monitor.NewMonitor(monitor.ObsConnectionInfo{
		Host:     fmt.Sprintf("%s:%s", *host, *port),
		Password: *password,
		CSVFile:  *csvFile,
	})
	if err != nil {
		panic(err)
	}
	defer monitor.Close()

	fmt.Println("\nPress Ctrl-C to exit")

	err = monitor.Start()
	if err != nil {
		fmt.Printf("Monitor error: %v\n", err)
	}

	waitForExit()
}

func waitForExit() {
	// Create a channel to wait for exit signal
	done := make(chan bool, 1)

	// Handle Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		done <- true
	}()

	// Wait for exit signal
	<-done
}
