# Watchdog Monitor

A Windows system metrics monitoring agent written in Go. Runs as a Windows Service, collects system metrics every 5 seconds, and publishes each reading to a NATS server in Protocol Buffers format.

## What it does

Watchdog continuously samples CPU, memory, GPU, network, and process data from your machine and streams it over NATS for downstream consumers (dashboards, time-series databases, alerting, etc.).

**Motivating question:** "Did the game stutter because the GPU was maxed out?" — Watchdog gives you the telemetry to answer it.

## Metrics collected

| Source | Data |
|--------|------|
| **CPU** | Total usage %, per-core usage % |
| **Memory** | Total bytes, available bytes, used % |
| **GPU** | VRAM total, temperature °C (Windows WMI) |
| **Network** | Bytes sent/received per interface (delta since last cycle) |
| **Processes** | Top 5 by CPU — name, PID, CPU %, memory (RSS) |

Each payload is timestamped in Unix milliseconds (UTC) and published to:
```
watchdog.<hostname>.metrics
```

> **Note:** GPU utilization % is always 0 due to a WMI limitation. VRAM is capped at ~4 GB (WMI uint32 property). GPU temperature is optional and may be absent on some systems.

## Architecture

```
cmd/watchdog/main.go          # Entry point (Windows only)
internal/
  collector/                  # CPU, memory, GPU, network, process collectors
  config/                     # TOML config loading and validation
  logger/                     # Structured logging with file rotation
  publisher/                  # NATS connection and protobuf publishing
  service/                    # Windows SCM integration + 5s collection cycle
proto/metrics.proto           # Protobuf schema
gen/metrics/                  # Generated Go code from proto
```

## Requirements

- **Go 1.25+**
- **Windows 10/11 x64** (runtime)
- **NATS server** accessible at the configured URL
- **protoc** (only if regenerating the protobuf schema)

## Building

```bash
# Build Windows binary (cross-compile from WSL2 or Windows)
make build

# Run tests
make test

# Regenerate protobuf code
make proto

# See all available commands
make help
```

The binary is output to `bin/watchdog.exe`.

## Configuration

Place a `watchdog.toml` next to `watchdog.exe`:

```toml
nats_url      = "nats://localhost:4222"  # required
log_path      = ""                        # optional: defaults to watchdog.log in exe dir
nats_user     = ""                        # optional: for future NATS auth
nats_password = ""                        # optional: for future NATS auth
```

Restrict read access to this file to the service account only — it will contain credentials when NATS auth is enabled.

## Running

### Interactively (dev/testing)

```bash
make run VERSION=1.0.0
```

Stop with `Ctrl+C`. In another terminal, watch the log:

```bash
make logs
```

### As a Windows Service

1. Build the binary and copy to the target machine.
2. Run `make install-service` to print the `sc create` command.
3. Paste the command into an elevated Command Prompt.
4. Start the service: `sc start WatchdogMonitor`

The service runs under a dedicated account. Required permissions:
- Read: `watchdog.exe`, `watchdog.toml`
- Write: log directory
- WMI namespaces: `root\cimv2` (GPU, processes) and `root\wmi` (temperature)

### Local NATS server (development)

```bash
docker compose up -d
```

Starts NATS 2.10 on port `4222` (client) and `8222` (HTTP monitoring).

## Integration

Downstream consumers subscribe to `watchdog.*.metrics` and deserialize the Protocol Buffers payload.

Full integration contract (payload schema, behavioral guarantees, example subscriber code): [`docs/publisher-contract.md`](docs/publisher-contract.md)

Quick example:

```go
nc.Subscribe("watchdog.*.metrics", func(msg *nats.Msg) {
    var snapshot metrics.SystemSnapshot
    if err := proto.Unmarshal(msg.Data, &snapshot); err != nil {
        log.Printf("decode error: %v", err)
        return
    }
    // snapshot.CpuTotalPercent, snapshot.Memory, snapshot.Gpu, etc.
})
```

## Design decisions

- **5-second absolute ticker** — drift-free; collection duration does not accumulate.
- **At-most-once delivery** — NATS core (no JetStream). Failed publishes are logged and dropped; the next cycle retries normally.
- **Graceful degradation** — each collector fails independently. A GPU error does not stop CPU or memory collection.
- **Error deduplication** — repeated identical errors are suppressed in the log after the first occurrence; recovery is also logged.
- **Windows-only build tag** — platform-agnostic code is isolated so the rest of the codebase compiles anywhere.

## License

MIT
