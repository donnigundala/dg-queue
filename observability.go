package dgqueue

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName = "github.com/donnigundala/dg-queue"
)

// RegisterMetrics registers queue metrics with OpenTelemetry.
// This initializes instruments and registers callbacks for observable metrics.
func (m *Manager) RegisterMetrics() error {
	meter := otel.GetMeterProvider().Meter(instrumentationName)

	var err error

	// 1. Observable Gauges (State)
	// Queue Depth
	m.metricQueueDepth, err = meter.Int64ObservableGauge(
		"queue.depth",
		metric.WithDescription("Current number of pending jobs in the queue"),
		metric.WithUnit("{job}"),
	)
	if err != nil {
		return err
	}

	// Active Workers
	m.metricActiveWorkers, err = meter.Int64ObservableGauge(
		"queue.workers",
		metric.WithDescription("Number of active workers processing jobs"),
		metric.WithUnit("{worker}"),
	)
	if err != nil {
		return err
	}

	// Register Callback for Gauges
	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		for name, pool := range m.workers {
			attrs := metric.WithAttributes(
				attribute.String("queue.name", name),
			)

			// Approximate depth: length of the channel
			o.ObserveInt64(m.metricQueueDepth, int64(len(pool.jobs)), attrs)

			// Active workers: concurrency (static for now, unless we track busy workers separately)
			// For better accuracy we might want to track 'busy' workers, but static concurrency is a good start
			o.ObserveInt64(m.metricActiveWorkers, int64(pool.concurrency), attrs)
		}
		return nil
	}, m.metricQueueDepth, m.metricActiveWorkers)
	if err != nil {
		return err
	}

	// 2. Sync Instruments (Events)
	// Job Processed Counter
	m.metricJobProcessed, err = meter.Int64Counter(
		"queue.job.processed",
		metric.WithDescription("Total number of jobs processed"),
		metric.WithUnit("{job}"),
	)
	if err != nil {
		return err
	}

	// Job Duration Histogram
	m.metricJobDuration, err = meter.Float64Histogram(
		"queue.job.duration",
		metric.WithDescription("Duration of job processing"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	return nil
}
