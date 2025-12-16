# OBS Monitor

A utility that monitors an OBS Studio instance via a WebSocket connection. Measures stream metrics and network latency with a 1-second measurement interval for low latency stream monitoring.

## Usage

```bash
obs-monitor -password <password> [-host localhost] [-port 4455] [-csv output.csv]
```

### Flags

- `-password` (required): OBS WebSocket password
- `-host` (optional): OBS WebSocket host (default: localhost)
- `-port` (optional): OBS WebSocket port (default: 4455)
- `-csv` (optional): CSV file to write metrics to

## CSV Export

When the `-csv` flag is provided, the monitor will write one line per second to the CSV file containing:

- `timestamp`: ISO 8601 timestamp
- `rtt_ms`: Round-trip time to the streaming server in milliseconds
- `ping_error`: Any error that occurred during ping
- `stream_active`: Whether the stream is currently active
- `output_bytes`: Total bytes sent to the streaming server
- `output_skipped_frames`: Number of frames skipped
- `stream_error`: Any error that occurred while getting stream status

Example:
```bash
obs-monitor -password mypassword -csv metrics.csv
```

## Development
- https://github.com/andreykaipov/goobs
- https://github.com/golang-standards/project-layout

Tests can be run via `go test ./...`
