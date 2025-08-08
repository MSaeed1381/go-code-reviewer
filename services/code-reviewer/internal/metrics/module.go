package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go_code_reviewer/pkg/log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once     sync.Once
	instance *Metrics
)

func Get() *Metrics {
	once.Do(func() {
		instance = newMetrics()
	})
	return instance
}

type Metrics struct {
	kafkaPublishCounter     *prometheus.CounterVec
	eventProcessCounter     *prometheus.CounterVec
	processLatencyHistogram *prometheus.HistogramVec
}

func newMetrics() *Metrics {
	return &Metrics{
		kafkaPublishCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_publish_total",
				Help: "Total number of kafka publish",
			},
			[]string{"status"},
		),
		eventProcessCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "event_process_total",
				Help: "Total number of processed events",
			},
			[]string{"status"},
		),
		processLatencyHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "event_process_latency_seconds",
				Help:    "Latency of processing events",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
	}
}

type Status string

const (
	Success Status = "success"
	Failure Status = "failure"
)

func (m *Metrics) ObserveKafkaPublish(status string) {
	m.kafkaPublishCounter.WithLabelValues(status).Inc()
}

func (m *Metrics) ObserveEventProcessing(status Status) {
	m.eventProcessCounter.With(prometheus.Labels{"status": string(status)}).Inc()
}

func (m *Metrics) ObserveEventProcessingLatency(status Status, start time.Time) {
	m.processLatencyHistogram.With(prometheus.Labels{"status": string(status)}).Observe(time.Since(start).Seconds())
}

func Init(address string) {
	metrics := Get()
	prometheus.MustRegister(metrics.eventProcessCounter, metrics.processLatencyHistogram, metrics.kafkaPublishCounter)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(address, nil); err != nil {
			log.GetLogger().WithError(err).Fatal("failed to serve prometheus metrics")
		}
	}()
}
