package kafka

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumer(t *testing.T) {
	// Используем адрес из переменной окружения или значение по умолчанию
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "kafka:9092"
	}

	// Проверяем доступность Kafka
	conn, err := net.DialTimeout("tcp", broker, 1*time.Second)
	if err != nil {
		t.Skip("Skipping test because Kafka is not available:", err)
	}
	conn.Close()

	topic := "test-consumer-topic"
	groupID := "test-group"
	testMessage := []byte("test message")

	// Создаем producer для отправки тестового сообщения
	producer, err := NewProducer(broker)
	require.NoError(t, err)
	defer producer.Close()

	// Отправляем тестовое сообщение
	err = producer.Produce(topic, testMessage)
	require.NoError(t, err)

	// Создаем consumer
	consumer, err := NewConsumer(broker, topic, groupID)
	require.NoError(t, err)
	defer consumer.Close()

	// Читаем сообщение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message, err := consumer.Consume(ctx)
	require.NoError(t, err)
	assert.Equal(t, testMessage, message)
}
