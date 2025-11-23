network-monitor/README.md
# Network Monitor

Network Monitor is a Go-based tool for monitoring network connectivity and latency to a configurable list of IP addresses. It periodically pings each IP and exposes detailed metrics about response times via a Prometheus-compatible `/metrics` endpoint.

## Features

- **IP Address Monitoring:** Periodically pings a list of IP addresses.
- **Prometheus Metrics:** Exposes ping duration metrics for scraping.
- **Configurable Logging:** Structured logs with adjustable log levels.
- **Docker Support:** Includes a Dockerfile for containerized deployment.

## How It Works

- Reads a comma-separated list of IPs to monitor.
- Pings each IP using ICMP at a configurable interval.
- Records response times in a Prometheus histogram.
- Runs an HTTP server exposing `/metrics` for Prometheus scraping.

## Getting Started

### Prerequisites

- Go 1.20+
- (Optional) Docker

### Build & Run

```sh
go build -o network-monitor ./cmd/network-monitor.go
./network-monitor
```

Or with Docker:

```sh
docker build -t network-monitor .
docker run -p 8080:8080 network-monitor
```

### Configuration

Configuration options are currently set in `internal/config/config.go`:

- `PingIps`: Comma-separated IPs to ping (default: `192.168.68.1, 192.168.1.1, 1.1.1.1, 8.8.8.8, 198.41.0.4`)
- `PingInterval`: Ping interval in seconds (default: not set, add as needed)
- `LogLevel`: Logging level (default: Info)
- `ServerPort`: HTTP server port (default: 8080)

## Metrics

Prometheus scrapes metrics from `/metrics`. Example metric:

- `ping_request_duration_seconds{ip="1.1.1.1"}`

# Check in Prometheus

Use this query in Prometheus:
```
histogram_quantile(0.99, sum by(le, ip) (rate(ping_request_duration_seconds_bucket{}[5m]))) * 1000
```
Or view [here](http://pi.local:9090/query?g0.expr=histogram_quantile%280.99%2C+sum+by%28le%2C+ip%29+%28rate%28ping_request_duration_seconds_bucket%7B%7D%5B5m%5D%29%29%29+*+1000&g0.show_tree=0&g0.tab=graph&g0.range_input=1h&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0) if you're me.
