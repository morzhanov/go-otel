package mq_test

import (
	"github.com/morzhanov/go-otel/internal/mq"
	"github.com/segmentio/kafka-go"
)

type MqMock struct {
	createReaderMock func(groupId string) *kafka.Reader
	connMock         func() *kafka.Conn
	kafkaUriMock     func() string
	topicMock        func() string
	writeMock        func() error
}

func (m *MqMock) CreateReader(groupId string) *kafka.Reader {
	return m.createReaderMock(groupId)
}
func (m *MqMock) Conn() *kafka.Conn {
	return m.connMock()
}
func (m *MqMock) KafkaUri() string {
	return m.kafkaUriMock()
}
func (m *MqMock) Topic() string {
	return m.topicMock()
}
func (m *MqMock) WriteMessage(_ interface{}) error {
	return m.writeMock()
}

func NewMqMock(
	createReader func(groupId string) *kafka.Reader,
	conn func() *kafka.Conn,
	kafkaUri func() string,
	topic func() string,
	write func() error,
) mq.MQ {
	m := MqMock{}
	if createReader != nil {
		m.createReaderMock = createReader
	} else {
		m.createReaderMock = func(groupId string) *kafka.Reader { return nil }
	}
	if conn != nil {
		m.connMock = conn
	} else {
		m.connMock = func() *kafka.Conn { return nil }
	}
	if kafkaUri != nil {
		m.kafkaUriMock = kafkaUri
	} else {
		m.kafkaUriMock = func() string { return "uri" }
	}
	if topic != nil {
		m.topicMock = topic
	} else {
		m.topicMock = func() string { return "topic" }
	}
	if write != nil {
		m.writeMock = write
	} else {
		m.writeMock = func() error { return nil }
	}
	return &m
}
