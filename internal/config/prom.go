package config

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	TotalPingsCounter  prometheus.Counter
	TotalTimoutCounter *prometheus.CounterVec
	DurationHist       *prometheus.HistogramVec
}

func NewMetrics(reg *prometheus.Registry) *Metrics {
	m := &Metrics{
		prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "ping_total",
				Help: "Total number of pings made",
			},
		),
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ping_total_timeouts",
				Help: "Total number of requests which timed out",
			},
			[]string{"ip"},
		),
		prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ping_request_duration_seconds",
				Help:    "Duration of the ping request in seconds",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 16.0},
			},
			[]string{"ip"},
		),
		// rate(ping_total_timeouts_total[5m]) / rate(ping_total[5m])
		// histogram_quantile(0.99, sum(rate(ping_request_duration_seconds_bucket[5m])) by (le))
	}
	reg.MustRegister(m.TotalPingsCounter)
	reg.MustRegister(m.TotalTimoutCounter)
	reg.MustRegister(m.DurationHist)
	return m
}
