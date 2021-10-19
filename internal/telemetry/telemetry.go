package telemetry

import (
	"github.com/morzhanov/go-otel/internal/telemetry/meter"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TraceFn func(name string, opts ...trace.TracerOption) trace.Tracer

type telemetry struct {
	tp TraceFn
	mp meter.Meter
}

type Telemetry interface {
	Tracer() TraceFn
	Meter() meter.Meter
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

func (t *telemetry) Tracer() TraceFn    { return t.tp }
func (t *telemetry) Meter() meter.Meter { return t.mp }

func NewTelemetry(url string, service string, log *zap.Logger) (Telemetry, error) {
	tp, err := tracerProvider(url, service)
	if err != nil {
		return nil, err
	}
	mtr, err := meter.NewMeter(log)
	if err != nil {
		return nil, err
	}
	return &telemetry{tp: tp, mp: mtr}, nil
}
