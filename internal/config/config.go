package config

import (
	"log/slog"
)

type Opts struct {
	PingIps      string // Comma separated list of IPs to ping
	PingInterval uint   // In seconds
	LogLevel     slog.Level
	ServerPort   string
}

func NewOpts() *Opts {
	config := &Opts{
		PingIps:    "192.168.68.1, 192.168.1.1, 1.1.1.1, 8.8.8.8, 79.79.79.79",
		LogLevel:   slog.LevelInfo,
		ServerPort: "8080",
	}

	return config
}
