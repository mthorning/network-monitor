package config

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"strings"
)

const defaultIps = "192.168.68.1, 192.168.1.1, 1.1.1.1, 8.8.8.8, 79.79.79.79"

type Opts struct {
	PingIps      []string
	PingInterval uint // In seconds
	LogLevel     slog.Level
	ServerPort   string
}

func NewOpts() *Opts {
	pingIps, err := GetIps(defaultIps)
	if err != nil {
		slog.Error("Default IPs can't be parsed", "error", err, "ips", defaultIps)
		os.Exit(1)
	}

	opts := &Opts{
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

	pingIps, err := GetIps(*stringIps)
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

func GetIps(ipsString string) ([]string, error) {
	ips := strings.Split(ipsString, ",")
	var processedIps []string

	for i := range ips {
		ip := strings.TrimSpace(ips[i])

		if ip != "" {
			processedIps = append(processedIps, ip)
		}
	}

	if len(processedIps) == 0 {
		return nil, errors.New("No valid IP addresses supplied")
	}

	return processedIps, nil
}
