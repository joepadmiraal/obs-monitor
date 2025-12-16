package metric

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andreykaipov/goobs"
	probing "github.com/prometheus-community/pro-bing"
)

type Pinger struct {
	client      *goobs.Client
	domain      string
	metricsChan chan PingMetrics
}

type PingMetrics struct {
	Timestamp time.Time
	RTT       time.Duration
	Error     error
}

func NewPinger(client *goobs.Client) (*Pinger, error) {
	return &Pinger{
		client:      client,
		metricsChan: make(chan PingMetrics, 10),
	}, nil
}

func (p *Pinger) GetMetricsChan() <-chan PingMetrics {
	return p.metricsChan
}

func (p *Pinger) Start() error {
	streamSettings, err := p.client.Config.GetStreamServiceSettings()
	if err != nil {
		return fmt.Errorf("failed to get stream settings: %w", err)
	}

	serverURL := streamSettings.StreamServiceSettings.Server
	if serverURL == "" {
		return fmt.Errorf("stream server URL not found in settings")
	}

	domain, err := extractDomain(serverURL)
	if err != nil {
		return fmt.Errorf("failed to extract domain from URL: %w", err)
	}

	p.domain = domain
	fmt.Printf("Pinging stream server: %s\n\n", domain)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		go func() {
			timestamp := time.Now()
			rtt, err := p.ping(domain)

			select {
			case p.metricsChan <- PingMetrics{
				Timestamp: timestamp,
				RTT:       rtt,
				Error:     err,
			}:
			default:
			}
		}()
	}

	return nil
}

func (p *Pinger) ping(domain string) (time.Duration, error) {
	pinger, err := probing.NewPinger(domain)
	if err != nil {
		return 0, err
	}

	pinger.Count = 1
	pinger.Timeout = 1 * time.Second
	pinger.SetPrivileged(false)

	err = pinger.Run()
	if err != nil {
		return 0, err
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return 0, fmt.Errorf("no response received")
	}

	return stats.AvgRtt, nil
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
