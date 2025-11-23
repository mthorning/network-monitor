package config

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	DurationHist *prometheus.HistogramVec
}

func NewMetrics(reg *prometheus.Registry) *Metrics {
	m := &Metrics{
		prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ping_request_duration_seconds",
				Help:    "Duration of the ping request in seconds",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0},
			},
			[]string{"ip"},
		),
		// histogram_quantile(0.99, sum(rate(ping_request_duration_seconds_bucket[5m])) by (le))
	}
	reg.MustRegister(m.DurationHist)
	return m
}
