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

func main() {
	opts := config.NewOpts()
	registry := prometheus.NewRegistry()
	metrics := config.NewMetrics(registry)

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     opts.LogLevel,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting Network Monitor", "ips", strings.Join(opts.PingIps, ", "))
	slog.Debug("Configuration options",
		"PingIps", opts.PingIps,
		"PingInterval", opts.PingInterval,
		"ServerPort", opts.ServerPort,
	)

	pinger := monitoring.NewPinger(opts, metrics)
	pinger.Run()

	http.Handle("/metrics",
		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			}),
	)

	err := http.ListenAndServe(fmt.Sprintf(":%s", opts.ServerPort), nil)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
