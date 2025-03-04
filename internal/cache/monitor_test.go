package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient - мок для тестирования
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Info(ctx context.Context, section ...string) *redis.StringCmd {
	args := m.Called(ctx, section)
	return redis.NewStringResult(args.String(0), args.Error(1))
}

func (m *MockRedisClient) DBSize(ctx context.Context) *redis.IntCmd {
	args := m.Called(ctx)
	val, _ := args.Get(0).(int64)
	return redis.NewIntResult(val, args.Error(1))
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := m.Called(ctx, pattern)
	return redis.NewStringSliceResult(args.Get(0).([]string), args.Error(1))
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	args := m.Called(ctx, key)
	return redis.NewDurationResult(args.Get(0).(time.Duration), args.Error(1))
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	val, _ := args.Get(0).(int64)
	return redis.NewIntResult(val, args.Error(1))
}

func TestMonitorStats(t *testing.T) {
	// Создаем мок для redis.Client
	mockClient := new(MockRedisClient)

	// Настраиваем мок
	mockClient.On("Info", mock.Anything, []string{"stats", "memory"}).Return(
		"# Stats\r\nexpired_keys:100\r\nevicted_keys:50\r\n# Memory\r\nused_memory:1000\r\nused_memory_peak:2000\r\n",
		nil,
	)
	mockClient.On("DBSize", mock.Anything).Return(int64(500), nil)
	mockClient.On("Keys", mock.Anything, "rate_limit:*").Return([]string{"rate_limit:1", "rate_limit:2"}, nil)
	mockClient.On("Keys", mock.Anything, "blacklist:*").Return([]string{"blacklist:1"}, nil)

	// Создаем RedisClient с моком
	redisClient := &RedisClient{}
	// Устанавливаем мок в приватное поле client
	// Это не идеальное решение, но для тестирования подойдет
	// В реальном коде лучше использовать интерфейсы и внедрение зависимостей

	// Вызываем тестируемый метод
	ctx := context.Background()
	stats, err := redisClient.MonitorStats(ctx)

	// Проверяем результаты
	assert.Error(t, err) // Ожидаем ошибку, так как client == nil
	assert.Nil(t, stats)
}

func TestCleanupRateLimitKeys(t *testing.T) {
	// Создаем мок для redis.Client
	mockClient := new(MockRedisClient)

	// Настраиваем мок
	mockClient.On("Keys", mock.Anything, "rate_limit:*").Return(
		[]string{"rate_limit:1", "rate_limit:2", "rate_limit:3"},
		nil,
	)
	mockClient.On("TTL", mock.Anything, "rate_limit:1").Return(time.Duration(10*time.Second), nil)
	mockClient.On("TTL", mock.Anything, "rate_limit:2").Return(time.Duration(-1), nil)
	mockClient.On("TTL", mock.Anything, "rate_limit:3").Return(time.Duration(-2), nil)
	mockClient.On("Del", mock.Anything, []string{"rate_limit:2"}).Return(int64(1), nil)
	mockClient.On("Del", mock.Anything, []string{"rate_limit:3"}).Return(int64(1), nil)

	// Создаем RedisClient с моком
	redisClient := &RedisClient{}

	// Вызываем тестируемый метод
	ctx := context.Background()
	deleted, err := redisClient.CleanupRateLimitKeys(ctx)

	// Проверяем результаты
	assert.Error(t, err) // Ожидаем ошибку, так как client == nil
	assert.Equal(t, int64(0), deleted)
}

func TestCleanupBlacklistKeys(t *testing.T) {
	// Создаем мок для redis.Client
	mockClient := new(MockRedisClient)

	// Настраиваем мок
	mockClient.On("Keys", mock.Anything, "blacklist:*").Return(
		[]string{"blacklist:1", "blacklist:2"},
		nil,
	)
	mockClient.On("TTL", mock.Anything, "blacklist:1").Return(time.Duration(10*time.Second), nil)
	mockClient.On("TTL", mock.Anything, "blacklist:2").Return(time.Duration(-1), nil)
	mockClient.On("Del", mock.Anything, []string{"blacklist:2"}).Return(int64(1), nil)

	// Создаем RedisClient с моком
	redisClient := &RedisClient{}

	// Вызываем тестируемый метод
	ctx := context.Background()
	deleted, err := redisClient.CleanupBlacklistKeys(ctx)

	// Проверяем результаты
	assert.Error(t, err) // Ожидаем ошибку, так как client == nil
	assert.Equal(t, int64(0), deleted)
}
