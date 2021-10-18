package controller

import (
	"context"

	"github.com/morzhanov/go-realworld/internal/common/config"
	"github.com/morzhanov/go-realworld/internal/common/mq"
	"github.com/morzhanov/go-realworld/internal/common/sender"
	"github.com/morzhanov/go-realworld/internal/common/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type baseController struct {
	tracer          opentracing.Tracer
	sender          sender.Sender
	mq              mq.MQ
	logger          *zap.Logger
	consumerGroupId string
}

type BaseController interface {
	CreateSpan(in *kafka.Message) opentracing.Span
	Listen(ctx context.Context, processRequest func(*kafka.Message))
	Logger() *zap.Logger
	ConsumerGroupId() string
}

func (c *baseController) CreateSpan(in *kafka.Message) opentracing.Span {
	return tracing.StartSpanFromEventsRequest(c.tracer, in)
}

func (c *baseController) Listen(
	ctx context.Context,
	processRequest func(*kafka.Message),
) {
	r := c.mq.CreateReader(c.consumerGroupId)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			c.logger.Error(err.Error())
			continue
		}
		go processRequest(&m)
		select {
		case <-ctx.Done():
			break
		default:
			continue
		}
	}
}

func (c *baseController) Logger() *zap.Logger {
	return c.logger
}

func (c *baseController) ConsumerGroupId() string {
	return c.consumerGroupId
}

func NewController(
	s sender.Sender,
	tracer opentracing.Tracer,
	logger *zap.Logger,
	conf *config.Config,
) (BaseController, error) {
	msgQ, err := mq.NewMq(conf, conf.KafkaTopic)
	if err != nil {
		return nil, err
	}
	c := &baseController{
		sender:          s,
		tracer:          tracer,
		mq:              msgQ,
		logger:          logger,
		consumerGroupId: conf.KafkaConsumerGroupId,
	}
	return c, err
}
