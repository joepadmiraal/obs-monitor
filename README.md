# OBS Monitor

A utility that monitors an OBS Studio instance via a WebSocket connection. Measures stream metrics, network latency and generic performance metrics.

It connects to OBS via a WebSocket connection.
All generic metrics are collected from the machine that runs OBS Monitor so it makes sense to run this on the same machine as OBS itself.

It's possible that within one write window, multiple measurements are collected. In which case it will take the max value of the measurements.

## Usage

```bash
obs-monitor -password <password>
```

Example output

```bash
Press Ctrl-C to exit
Pinging google.com every 1s
Pinging a.rtmp.youtube.com every every 1s
OBS Studio version: 32.0.4
Server protocol version: 5.6.3
Client protocol version: 5.5.6
Client library version: 1.5.6

timestamp                 | obs_rtt_ms | google_rtt_ms | stream_active | output_bytes | output_skipped_frames | errors
--------------------------|------------|---------------|---------------|--------------|-----------------------|--------
2025-12-16T15:07:37+01:00 |       5.80 |          4.59 |          true |       915861 |                     0 | 
2025-12-16T15:07:38+01:00 |       9.29 |          5.72 |          true |      1009403 |                     0 | 
2025-12-16T15:07:39+01:00 |       6.43 |          7.03 |          true |      1053680 |                     0 | 
2025-12-16T15:07:40+01:00 |       9.25 |          6.65 |          true |      1071794 |                     0 | 
2025-12-16T15:07:41+01:00 |       8.78 |          3.80 |          true |       980014 |                     0 | 
2025-12-16T15:07:42+01:00 |       7.17 |          3.86 |          true |       823529 |                     0 | 
```

### Flags

- `-password` (optional): OBS WebSocket password, the program will ask for it if it's not provided
- `-host` (optional): OBS WebSocket host (default: localhost)
- `-port` (optional): OBS WebSocket port (default: 4455)
- `-csv` (optional): CSV file to write metrics to, set to empty to prevent csv file generation (default: obs-monitor.csv)
- `-metric-interval` (optional): Metric collection interval in milliseconds (default: 1000ms)
- `-writer-interval` (optional): Writer interval in milliseconds (default: 1000ms)

## CSV Export

The monitor will write one line per second to the CSV file containing:

- `timestamp`: ISO 8601 timestamp
- `obs_rtt_ms`: Round-trip time to the streaming server in milliseconds
- `google_rtt_ms`: Round-trip time to Google in milliseconds
- `stream_active`: Whether the stream is currently active
- `output_bytes`: Total bytes sent to the streaming server during the writer-interval
- `output_skipped_frames`: Number of frames skipped during the writer-interval
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
