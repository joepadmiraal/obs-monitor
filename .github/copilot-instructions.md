This is a small utility that monitors an OBS Studio instance via a WebSocket connection.
It's used for monitoring low latency streams so we use a measurement interval of 1 second.

`/internal/monitor/monitor.go` contains the orchestration of the monitoring.
`/internal/writer/` contains several writers for outputting the monitoring results. 

## Code style
- Be very sparse with comments. Only add them when the code is not self-explanatory.

