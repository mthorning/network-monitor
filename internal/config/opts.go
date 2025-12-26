package config

import (
	"flag"
	"log/slog"
	"network_monitor/internal/utils"
	"os"
)

const defaultIps = "192.168.68.1, 192.168.1.1, 1.1.1.1, 8.8.8.8, 79.79.79.79"

type Opts struct {
	PingIps      []string
	PingInterval uint // In seconds
	LogLevel     slog.Level
	ServerPort   string
}

func NewOpts() Opts {
	pingIps, err := utils.GetIps(defaultIps)
	if err != nil {
		slog.Error("Default IPs can't be parsed", "error", err, "ips", defaultIps)
		os.Exit(1)
	}

	opts := Opts{
		PingIps:      pingIps,
		LogLevel:     slog.LevelInfo,
		PingInterval: 15,
		ServerPort:   "8080",
	}

	opts.ParseFlags()

	return opts
}

func (o *Opts) ParseFlags() {
	stringIps := flag.String("ping-ips", defaultIps, "A comma-separated list of IPs to ping")
	pingInterval := flag.Uint("ping-interval", o.PingInterval, "Interval betweeen pings in seconds")
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
