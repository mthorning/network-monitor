package config

import (
	"flag"
	"log/slog"
)

type Opts struct {
	PingIps      string // Comma separated list of IPs to ping
	PingInterval uint   // In seconds
	LogLevel     slog.Level
	ServerPort   string
}

func NewOpts() *Opts {
	opts := &Opts{
		PingIps:      "192.168.68.1, 192.168.1.1, 1.1.1.1, 8.8.8.8, 79.79.79.79",
		LogLevel:     slog.LevelInfo,
		PingInterval: 15,
		ServerPort:   "8080",
	}

	opts.ParseFlags()

	return opts
}

func (o *Opts) ParseFlags() {
	pingIps := flag.String("ping-ips", o.PingIps, "A comma-separated list of IPs to ping")
	pingInterval := flag.Uint("ping-interval", o.PingInterval, "Interval betweeen pings in seconds")
	logLevel := flag.String("log-level", o.LogLevel.String(), "One of ERROR, WARN, INFO, or DEBUG")
	serverPort := flag.String("server-port", o.ServerPort, "Port to serve metrics on")

	flag.Parse()

	o.PingIps = *pingIps
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
