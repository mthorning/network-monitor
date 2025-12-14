package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"network-monitor/internal/config"
	"network-monitor/internal/utils"
	"os"

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

	ips, err := utils.GetIps(opts.PingIps)
	if err != nil {
		slog.Error(err.Error(), "config.PingIps", opts.PingIps)
		os.Exit(1)
	}

	slog.Info("Starting Network Monitor", "IPs", ips)
	slog.Debug("Configuration options",
		"PingIps", opts.PingIps,
		"PingInterval", opts.PingInterval,
		"ServerPort", opts.ServerPort,
	)

	utils.PingAndReport(ips, opts, metrics)

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
}
