package courier

import (
	"delivery/internal/business/models"
	"fmt"
)

// Определяет интерфейс для хранилища курьеров
type CourierStorer interface {
	Add(courier models.Courier) (int, error)
	Get(id int) (models.Courier, error)
	Update(courier models.Courier) error
	Delete(id int) error
	GetAll() ([]models.Courier, error)
	GetAvailableCouriers() ([]models.Courier, error)
}

type CourierService struct {
	store CourierStorer
}

func NewCourierService(store CourierStorer) *CourierService {
	return &CourierService{store: store}
}

func (s *CourierService) Create(courier *models.Courier) error {
	if courier.Status == "" {
		courier.Status = "available" // По умолчанию курьер доступен
	}

	id, err := s.store.Add(*courier)
	if err != nil {
		return err
	}
	courier.ID = id
	return nil
}

func (s *CourierService) Get(id int) (*models.Courier, error) {
	courier, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("курьер не найден: %w", err)
	}
	return &models.Courier{
		ID:        courier.ID,
		Name:      courier.Name,
		Phone:     courier.Phone,
		Email:     courier.Email,
		VehicleID: courier.VehicleID,
		Status:    courier.Status,
	}, nil
}

func (s *CourierService) Update(id int, courier *models.Courier) error {
	cour := models.Courier{
		ID:        id,
		Name:      courier.Name,
		Phone:     courier.Phone,
		Email:     courier.Email,
		VehicleID: courier.VehicleID,
		Status:    courier.Status,
	}

	return s.store.Update(cour)
}

func (s *CourierService) Delete(id int) error {
	return s.store.Delete(id)
}

func (s *CourierService) List() ([]models.Courier, error) {
	couriers, err := s.store.GetAll()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка курьеров: %w", err)
	}

	return couriers, nil
}

func (s *CourierService) GetAvailableCouriers() ([]models.Courier, error) {
	couriers, err := s.store.GetAvailableCouriers()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении доступных курьеров: %w", err)
	}

	return couriers, nil
}

func (s *CourierService) UpdateCourierStatus(id int, status string) error {
	courier, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("курьер не найден: %w", err)
	}

	// Проверка допустимых статусов
	if status != "available" && status != "busy" && status != "offline" {
		return fmt.Errorf("недопустимый статус курьера: %s", status)
	}

	courier.Status = status
	return s.store.Update(courier)
}
