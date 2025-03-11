package delivery

import (
	"delivery/internal/api"
	"delivery/internal/business/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeliveryStore - мок для DeliveryStore
type MockDeliveryStore struct {
	mock.Mock
}

func (m *MockDeliveryStore) Add(delivery models.Delivery) (int, error) {
	args := m.Called(delivery)
	return args.Int(0), args.Error(1)
}

func (m *MockDeliveryStore) Get(id int) (models.Delivery, error) {
	args := m.Called(id)
	return args.Get(0).(models.Delivery), args.Error(1)
}

func (m *MockDeliveryStore) Update(delivery models.Delivery) error {
	args := m.Called(delivery)
	return args.Error(0)
}

func (m *MockDeliveryStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDeliveryStore) GetByCourierID(courierID int) ([]models.Delivery, error) {
	args := m.Called(courierID)
	return args.Get(0).([]models.Delivery), args.Error(1)
}

func (m *MockDeliveryStore) GetByParcelID(parcelID int) (models.Delivery, error) {
	args := m.Called(parcelID)
	return args.Get(0).(models.Delivery), args.Error(1)
}

// TestWebSocketIntegration - тест для демонстрации интеграции WebSocket
func TestWebSocketIntegration(t *testing.T) {
	// Создаем WebSocket сервер для тестирования
	wsManager := api.NewWebSocketManager()
	go wsManager.Run()

	// Создаем HTTP сервер для тестирования
	server := httptest.NewServer(http.HandlerFunc(wsManager.WebSocketHandler))
	defer server.Close()

	// Заменяем "http" на "ws" в URL сервера
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Подключаемся к WebSocket серверу
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Не удалось подключиться к WebSocket серверу: %v", err)
	}
	defer ws.Close()

	// Создаем мок для DeliveryStore
	mockStore := new(MockDeliveryStore)

	// Настраиваем мок для возврата тестовой доставки
	testDelivery := models.Delivery{
		ID:         1,
		ParcelID:   100,
		CourierID:  200,
		Status:     "assigned",
		AssignedAt: time.Now(),
	}

	// Настраиваем ожидаемое поведение мока
	mockStore.On("Get", 1).Return(testDelivery, nil)
	mockStore.On("Update", mock.Anything).Return(nil)

	// В этом тесте мы не используем DeliveryService напрямую,
	// а просто проверяем, что WebSocket менеджер работает корректно

	// Отправляем тестовое сообщение через WebSocket менеджер
	wsManager.BroadcastOrderStatusUpdate("1", "delivered")

	// Ожидаем получения сообщения от WebSocket сервера
	messageType, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Ошибка при чтении сообщения: %v", err)
	}

	// Проверяем, что сообщение получено и имеет правильный тип
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Contains(t, string(message), "ORDER_STATUS_UPDATE")
	assert.Contains(t, string(message), "1")
	assert.Contains(t, string(message), "delivered")
}

// Пример того, как можно было бы расширить DeliveryService для поддержки WebSocket
/*
// Добавляем поле wsManager в структуру DeliveryService
type DeliveryService struct {
	store       *DeliveryStore
	cacheClient *cache.RedisClient
	wsManager   *api.WebSocketManager
}

// Добавляем метод для установки WebSocket менеджера
func (s *DeliveryService) WithWebSocket(wsManager *api.WebSocketManager) *DeliveryService {
	s.wsManager = wsManager
	return s
}

// Модифицируем метод CompleteDelivery для отправки уведомлений через WebSocket
func (s *DeliveryService) CompleteDelivery(deliveryID int) error {
	delivery, err := s.store.Get(deliveryID)
	if err != nil {
		return fmt.Errorf("Ошибка при получении доставки: %w", err)
	}

	if delivery.Status != "assigned" && delivery.Status != "in progress" {
		return fmt.Errorf("Завершение доставки недоступно для статуса: %s", delivery.Status)
	}

	delivery.Status = "delivered"
	delivery.DeliveredAt = time.Now().UTC()

	// Обновляем кэш
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:%d", deliveryID)

		// Удаляем старые данные из кэша
		s.cacheClient.Delete(ctx, cacheKey)

		// Инвалидируем кэш списка доставок
		s.cacheClient.Delete(ctx, "deliveries:list")

		// Сохраняем обновленные данные в кэш
		s.cacheClient.SetJSON(ctx, cacheKey, delivery, 30*time.Minute)
	}

	// Отправляем уведомление через WebSocket, если менеджер инициализирован
	if s.wsManager != nil {
		s.wsManager.BroadcastOrderStatusUpdate(fmt.Sprintf("%d", deliveryID), "delivered")
	}

	return s.store.Update(delivery)
}
*/
