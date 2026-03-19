package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"network_monitor/internal/config"
	"network_monitor/internal/monitoring"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var BuildTime string = "not set"

func main() {
	opts := config.NewOpts()
	registry := prometheus.NewRegistry()
	metrics := config.NewMetrics(registry)

	var handler slog.Handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     opts.LogLevel,
		AddSource: true,
	})
	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", "network-monitor"),
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting Network Monitor", "ips", strings.Join(opts.PingIps, ","), "build", BuildTime)
	slog.Debug("Configuration options",
		"PingIps", opts.PingIps,
		"PingInterval", opts.PingInterval,
		"ServerPort", opts.ServerPort,
	)

	manager, err := monitoring.NewManager(opts, metrics)
	if err != nil {
		slog.Error("Failed to create new pinger", "error", err)
		os.Exit(1)
	}
	manager.Run()

	http.Handle("/metrics",
		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			}),
	)

	err = http.ListenAndServe(fmt.Sprintf(":%s", opts.ServerPort), nil)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	slog.Debug("Serving metrics at /metrics", "port", opts.ServerPort)
}
