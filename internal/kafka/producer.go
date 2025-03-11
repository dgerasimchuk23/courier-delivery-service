package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		BatchTimeout: 5 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{writer: writer}, nil
}

func (p *Producer) Produce(topic string, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Value: value,
		},
	)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
