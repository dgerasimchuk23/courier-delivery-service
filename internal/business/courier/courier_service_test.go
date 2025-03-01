package courier

import (
	"delivery/internal/business/models"
	"errors"
	"testing"
)

// Мок-объект хранилища для тестирования
type MockCourierStore struct {
	couriers    map[int]models.Courier
	nextID      int
	shouldError bool
}

func NewMockCourierStore() *MockCourierStore {
	return &MockCourierStore{
		couriers: make(map[int]models.Courier),
		nextID:   1,
	}
}

func (m *MockCourierStore) Add(courier models.Courier) (int, error) {
	if m.shouldError {
		return 0, errors.New("ошибка при добавлении")
	}
	courier.ID = m.nextID
	m.couriers[m.nextID] = courier
	m.nextID++
	return courier.ID, nil
}

func (m *MockCourierStore) Get(id int) (models.Courier, error) {
	if m.shouldError {
		return models.Courier{}, errors.New("ошибка при получении")
	}
	courier, exists := m.couriers[id]
	if !exists {
		return models.Courier{}, errors.New("курьер не найден")
	}
	return courier, nil
}

func (m *MockCourierStore) Update(courier models.Courier) error {
	if m.shouldError {
		return errors.New("ошибка при обновлении")
	}
	_, exists := m.couriers[courier.ID]
	if !exists {
		return errors.New("курьер не найден")
	}
	m.couriers[courier.ID] = courier
	return nil
}

func (m *MockCourierStore) Delete(id int) error {
	if m.shouldError {
		return errors.New("ошибка при удалении")
	}
	_, exists := m.couriers[id]
	if !exists {
		return errors.New("курьер не найден")
	}
	delete(m.couriers, id)
	return nil
}

func (m *MockCourierStore) GetAll() ([]models.Courier, error) {
	if m.shouldError {
		return nil, errors.New("ошибка при получении списка")
	}
	var couriers []models.Courier
	for _, c := range m.couriers {
		couriers = append(couriers, c)
	}
	return couriers, nil
}

func (m *MockCourierStore) GetAvailableCouriers() ([]models.Courier, error) {
	if m.shouldError {
		return nil, errors.New("ошибка при получении списка")
	}
	var couriers []models.Courier
	for _, c := range m.couriers {
		if c.Status == "available" {
			couriers = append(couriers, c)
		}
	}
	return couriers, nil
}

func TestCourierService_Create(t *testing.T) {
	mockStore := NewMockCourierStore()
	service := NewCourierService(mockStore)

	courier := &models.Courier{
		Name:      "Тестовый Курьер",
		Phone:     "+79001234567",
		Email:     "test.courier@example.com",
		VehicleID: "A123BC",
	}

	// Проверяем успешное создание
	err := service.Create(courier)
	if err != nil {
		t.Fatalf("Ошибка при создании курьера: %v", err)
	}

	if courier.ID != 1 {
		t.Errorf("ожидался ID=1, получено ID=%d", courier.ID)
	}

	// Проверяем, что статус по умолчанию установлен
	if courier.Status != "available" {
		t.Errorf("ожидался статус 'available', получено '%s'", courier.Status)
	}

	// Проверяем ошибку
	mockStore.shouldError = true
	err = service.Create(courier)
	if err == nil {
		t.Errorf("ожидалась ошибка, но её не было")
	}
}
