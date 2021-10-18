package payment

import (
	"context"
	"encoding/json"

	gpayment "github.com/morzhanov/go-otel/api/grpc/payment"
	"github.com/morzhanov/go-otel/internal/config"
	"github.com/morzhanov/go-otel/internal/event"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type eventController struct {
	event.BaseController
	pay Payment
}

type Controller interface {
	Listen(ctx context.Context)
}

func (c *eventController) processPayment(in *kafka.Message) {
	span := c.CreateSpan(in)
	defer span.Finish()

	res := gpayment.ProcessPaymentMessage{}
	if err := json.Unmarshal(in.Value, &res); err != nil {
		c.Logger().Error("error during process payment event processing", zap.Error(err))
	}
	if err := c.pay.ProcessPayment(&res); err != nil {
		c.Logger().Error("error during process payment event processing", zap.Error(err))
	}
}

func (c *eventController) Listen(ctx context.Context) {
	c.BaseController.Listen(ctx, c.processPayment)
}

func NewPicturesEventsController(
	tracer opentracing.Tracer,
	logger *zap.Logger,
	pay Payment,
	c *config.Config,
) (Controller, error) {
	controller, err := event.NewController(tracer, logger, c)
	return &eventController{BaseController: controller, pay: pay}, err
}
