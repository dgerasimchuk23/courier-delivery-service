package kafka

import (
	"context"
	"log"
	"os"
)

// Client представляет собой клиент Kafka, объединяющий Producer и Consumer
type Client struct {
	Producer  *Producer
	Consumers map[string]*Consumer
}

// NewClient создает нового клиента Kafka
func NewClient() (*Client, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092" // Значение по умолчанию
	}

	producer, err := NewProducer(broker)
	if err != nil {
		return nil, err
	}

	return &Client{
		Producer:  producer,
		Consumers: make(map[string]*Consumer),
	}, nil
}

// RegisterConsumer регистрирует нового Consumer для указанного топика и группы
func (c *Client) RegisterConsumer(topic, groupID string) (*Consumer, error) {
	key := topic + "-" + groupID
	if consumer, exists := c.Consumers[key]; exists {
		return consumer, nil
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092" // Значение по умолчанию
	}

	consumer, err := NewConsumer(broker, topic, groupID)
	if err != nil {
		return nil, err
	}

	c.Consumers[key] = consumer
	return consumer, nil
}

// StartTestConsumer запускает тестового потребителя для указанного топика
func (c *Client) StartTestConsumer(topic, groupID string) error {
	consumer, err := c.RegisterConsumer(topic, groupID)
	if err != nil {
		return err
	}

	// Запускаем чтение сообщений в отдельной горутине
	go func() {
		ctx := context.Background()
		for {
			msg, err := consumer.Consume(ctx)
			if err != nil {
				log.Printf("Error consuming message: %v", err)
				return
			}
			log.Printf("Processing message: %s", string(msg))
		}
	}()

	return nil
}

// SendTestMessage отправляет тестовое сообщение в указанный топик
func (c *Client) SendTestMessage(topic string, message string) error {
	return c.Producer.Produce(topic, []byte(message))
}

// Close закрывает клиент Kafka и все связанные ресурсы
func (c *Client) Close() error {
	var errs []error

	// Закрываем producer
	if err := c.Producer.Close(); err != nil {
		errs = append(errs, err)
	}

	// Закрываем все consumers
	for _, consumer := range c.Consumers {
		if err := consumer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0] // Возвращаем первую ошибку для простоты
	}
	return nil
}
