package kafka

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProducer(t *testing.T) {
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

	topic := "test-producer-topic"
	testMessage := []byte("test message")

	producer, err := NewProducer(broker)
	require.NoError(t, err)
	defer producer.Close()

	err = producer.Produce(topic, testMessage)
	assert.NoError(t, err)
}
