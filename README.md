# OBS Monitor

A utility that monitors an OBS Studio instance via a WebSocket connection. Measures stream metrics, network latency and generic performance metrics with a 1-second measurement interval for low latency stream monitoring.

It connects to OBS via a WebSocket connection.
All generic metrics are collected from the machine that runs OBS Monitor so it makes sense to run this on the same machine as OBS itself.

## Usage

```bash
obs-monitor -password <password>
```

### Flags

- `-password` (required): OBS WebSocket password
- `-host` (optional): OBS WebSocket host (default: localhost)
- `-port` (optional): OBS WebSocket port (default: 4455)
- `-csv` (optional): CSV file to write metrics to
- `-metric-interval` (optional): Metric collection interval in milliseconds (default: 1000ms)
- `-writer-interval` (optional): Writer interval in milliseconds (default: 1000ms)

## CSV Export

When the `-csv` flag is provided, the monitor will write one line per second to the CSV file containing:

- `timestamp`: ISO 8601 timestamp
- `obs_rtt_ms`: Round-trip time to the streaming server in milliseconds
- `google_rtt_ms`: Round-trip time to Google in milliseconds
- `stream_active`: Whether the stream is currently active
- `output_bytes`: Total bytes sent to the streaming server
- `output_skipped_frames`: Number of frames skipped
- `stream_error`: Any error that occurred while getting stream status

Example:
```bash
obs-monitor -password mypassword -csv metrics.csv
```

## OBS

The WebSocket password can be set and read from `Tools->WebSocket Server Settings`.

![OBS WebSocket Server Settings 1](docs/obs-1.png)
![OBS WebSocket Server Settings 2](docs/obs-2.png)

## Development

- OBS WebSocket client: https://github.com/andreykaipov/goobs
- Golang project setup: https://github.com/golang-standards/project-layout

Tests can be run via `go test ./...`
