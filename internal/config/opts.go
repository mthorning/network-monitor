package config

import (
	"flag"
	"log/slog"
	"network_monitor/internal/utils"
	"os"
)

const defaultIps = "8.8.8.8"

type Opts struct {
	PingIps               []string
	PingInterval          int // In seconds
	TraceFrequency        int // In iterations
	TraceTimeoutThreshold int
	LogLevel              slog.Level
	ServerPort            string
}

func NewOpts() Opts {
	pingIps, err := utils.GetIps(defaultIps)
	if err != nil {
		slog.Error("Default IPs can't be parsed", "error", err, "ips", defaultIps)
		os.Exit(1)
	}

	opts := Opts{
		PingIps:               pingIps,
		LogLevel:              slog.LevelInfo,
		PingInterval:          15,
		TraceFrequency:        20,
		TraceTimeoutThreshold: 5,
		ServerPort:            "8080",
	}

	opts.ParseFlags()

	return opts
}

func (o *Opts) ParseFlags() {
	stringIps := flag.String("ping-ips", defaultIps, "A comma-separated list of IPs to ping")
	pingInterval := flag.Int("ping-interval", o.PingInterval, "Interval betweeen pings in seconds")
	traceFrequency := flag.Int("trace-frequency", o.TraceFrequency, "Will run a trace every x iterations of the loop")
	traceTimeoutThreshold := flag.Int("trace-timeout-threshold", o.TraceTimeoutThreshold, "Will run a trace after x timeouts")
	logLevel := flag.String("log-level", o.LogLevel.String(), "One of ERROR, WARN, INFO, or DEBUG")
	serverPort := flag.String("server-port", o.ServerPort, "Port to serve metrics on")

	flag.Parse()

	pingIps, err := utils.GetIps(*stringIps)
	if err != nil {
		slog.Error("Default IPs can't be parsed", "error", err, "ips", stringIps)
		os.Exit(1)
	}

	o.PingIps = pingIps
	o.PingInterval = *pingInterval
	o.TraceFrequency = *traceFrequency
	o.TraceTimeoutThreshold = *traceTimeoutThreshold
	o.ServerPort = *serverPort

	switch *logLevel {
	case "DEBUG":
		o.LogLevel = slog.LevelDebug
	case "INFO":
		o.LogLevel = slog.LevelInfo
	case "WARN":
		o.LogLevel = slog.LevelWarn
	case "ERROR":
		o.LogLevel = slog.LevelError
	}
}
