package meter

import (
	"net/http"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.uber.org/zap"
)

type mtr struct {
	provider metric.MeterProvider
}

type Meter interface {
	Provider() metric.MeterProvider
}

func InitMeter(log *zap.Logger) metric.MeterProvider {
	conf := prometheus.Config{}
	c := controller.New(
		processor.New(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(conf.DefaultHistogramBoundaries),
			),
			export.CumulativeExportKindSelector(),
			processor.WithMemory(true),
		),
	)
	exporter, err := prometheus.New(conf, c)
	if err != nil {
		log.Error("failed to initialize prometheus exporter %v", zap.Error(err))
	}
	http.HandleFunc("/", exporter.ServeHTTP)
	go func() {
		_ = http.ListenAndServe(":2222", nil)
	}()
	log.Info("Prometheus server running on :2222")
	return exporter.MeterProvider()
}

func (m *mtr) Provider() metric.MeterProvider {
	return m.provider
}

func NewMeter(log *zap.Logger) Meter {
	provider := InitMeter(log)
	return &mtr{provider: provider}
}
