package payment

import "github.com/morzhanov/go-otel/internal/mq"

type payment struct {
	mq mq.MQ
}

type Payment interface {
}

func NewPaymentService(mq mq.MQ) Payment {
	return &payment{mq}
}
