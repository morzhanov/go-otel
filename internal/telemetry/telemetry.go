package telemetry

import (
	"github.com/morzhanov/go-otel/internal/telemetry/meter"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TraceFn func(name string, opts ...trace.TracerOption) trace.Tracer

type telemetry struct {
	tp TraceFn
	mp metric.Meter
}

type Telemetry interface {
	Tracer() TraceFn
	Meter() metric.Meter
}

func tracerProvider(url string, service string) (TraceFn, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
		)),
	)
	return tp.Tracer, nil
}

func (t *telemetry) Tracer() TraceFn     { return t.tp }
func (t *telemetry) Meter() metric.Meter { return t.mp }

func NewTelemetry(url string, service string, log *zap.Logger) (Telemetry, error) {
	tp, err := tracerProvider(url, service)
	if err != nil {
		return nil, err
	}
	mtr := meter.NewMeter(log)
	return &telemetry{tp: tp, mp: mtr.Provider().Meter("prometheus")}, nil
}
