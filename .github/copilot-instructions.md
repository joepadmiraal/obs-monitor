This is a small utility that monitors an OBS Studio instance via a WebSocket connection.
It's used for monitoring low latency streams so we use a measurement interval of 1 second.

`/internal/monitor/monitor.go` contains the orchestration of the monitoring.
`/internal/metric/` contains several metric collectors.
`/internal/writer/` contains several writers for outputting the monitoring results. 

## Code style
- Be very sparse with comments. Only add them when the code is not self-explanatory.
- Use idiomatic Go.

## Development
- Always run `go test -race ./...` and `go fmt ./...` after you made changes to the source code.

## Testing Best Practices
1. **Isolation**: Mock all external dependencies (OBS WebSocket, network, system calls)
2. **Determinism**: Use fixed timestamps, controlled randomness
3. **Concurrency**: Use `-race` flag to detect race conditions
4. **Clear naming**: `TestComponentName_Method_Scenario` pattern
5. **Table-driven tests**: For testing multiple scenarios with similar setup
6. **Cleanup**: Use `t.Cleanup()` for resource cleanup