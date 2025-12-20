package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Config holds metrics configuration
type Config struct {
	Enabled bool
}

// Meter wraps OpenTelemetry meter
type Meter struct {
	meter metric.Meter
}

// New creates a new meter instance
func New(ctx context.Context, cfg Config, serviceName string) (*Meter, error) {
	if !cfg.Enabled {
		return &Meter{
			meter: otel.Meter("noop"),
		}, nil
	}

	// Get meter from global meter provider
	// In production, configure a proper meter provider with exporters
	meter := otel.Meter(serviceName)

	return &Meter{
		meter: meter,
	}, nil
}

// GetMeter returns the underlying meter
func (m *Meter) GetMeter() metric.Meter {
	return m.meter
}

// CreateCounter creates a new counter metric
func (m *Meter) CreateCounter(name, description string) (metric.Int64Counter, error) {
	counter, err := m.meter.Int64Counter(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter %s: %w", name, err)
	}
	return counter, nil
}

// CreateHistogram creates a new histogram metric
func (m *Meter) CreateHistogram(name, description, unit string) (metric.Float64Histogram, error) {
	histogram, err := m.meter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram %s: %w", name, err)
	}
	return histogram, nil
}

// CreateUpDownCounter creates a new up/down counter metric
func (m *Meter) CreateUpDownCounter(name, description string) (metric.Int64UpDownCounter, error) {
	counter, err := m.meter.Int64UpDownCounter(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create up/down counter %s: %w", name, err)
	}
	return counter, nil
}
