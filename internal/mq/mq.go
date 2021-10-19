package mq

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type msgq struct {
	conn     *kafka.Conn
	kafkaUri string
	topic    string
}

type MQ interface {
	CreateReader(groupId string) *kafka.Reader
	Conn() *kafka.Conn
	KafkaUri() string
	Topic() string
	WriteMessage(ctx context.Context, msg interface{}) error
}

func (m *msgq) createTopic() error {
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             m.topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}
	return m.conn.CreateTopics(topicConfigs...)
}

func (m *msgq) CreateReader(groupId string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{m.kafkaUri},
		Topic:    m.topic,
		GroupID:  groupId,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
}

func (m *msgq) Conn() *kafka.Conn {
	return m.conn
}

func (m *msgq) KafkaUri() string {
	return m.kafkaUri
}

func (m *msgq) Topic() string {
	return m.topic
}

func (m *msgq) WriteMessage(ctx context.Context, msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	spanCtx, err := json.Marshal(ctx)
	if err != nil {
		return err
	}
	h := kafka.Header{Key: "span-context", Value: spanCtx}
	kmsg := kafka.Message{Value: b, Headers: []kafka.Header{h}}
	if _, err := m.conn.WriteMessages(kmsg); err != nil {
		return err
	}
	if err := m.conn.Close(); err != nil {
		return err
	}
	return nil
}

func NewMq(uri string, topic string) (res MQ, err error) {
	conn, err := kafka.DialLeader(context.Background(), "tcp", uri, topic, 0)
	if err != nil {
		return nil, err
	}
	msgQ := msgq{
		conn,
		uri,
		topic,
	}
	if err := msgQ.createTopic(); err != nil {
		return nil, err
	}
	return &msgQ, nil
}
