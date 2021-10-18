package event

import (
	"context"

	"github.com/morzhanov/go-otel/internal/mq"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type baseController struct {
	//tracer          opentracing.Tracer
	mq      mq.MQ
	logger  *zap.Logger
	groupID string
}

type BaseController interface {
	//CreateSpan(in *kafka.Message) opentracing.Span
	Listen(ctx context.Context, processRequest func(*kafka.Message))
	Logger() *zap.Logger
	ConsumerGroupId() string
}

//func (c *baseController) CreateSpan(in *kafka.Message) opentracing.Span {
//	return tracing.StartSpanFromEventsRequest(c.tracer, in)
//}

func (c *baseController) Listen(
	ctx context.Context,
	processRequest func(*kafka.Message),
) {
	r := c.mq.CreateReader(c.groupID)
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
	return c.groupID
}

func NewController(
	//tracer opentracing.Tracer,
	logger *zap.Logger,
	kafkaUrl string,
	kafkaTopic string,
	kafkaGroupID string,
) (BaseController, error) {
	msgQ, err := mq.NewMq(kafkaUrl, kafkaTopic)
	if err != nil {
		return nil, err
	}
	c := &baseController{
		//tracer:          tracer,
		mq:      msgQ,
		logger:  logger,
		groupID: kafkaGroupID,
	}
	return c, err
}
