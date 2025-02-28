package delivery

import (
	"delivery/internal/business/models"
	"fmt"
	"time"
)

type DeliveryService struct {
	store *DeliveryStore
}

func NewDeliveryService(store *DeliveryStore) *DeliveryService {
	return &DeliveryService{store: store}
}

func (s *DeliveryService) Create(delivery *models.Delivery) error {
	d := models.Delivery{
		ParcelID:   delivery.ParcelID,
		CourierID:  delivery.CourierID,
		Status:     delivery.Status,
		AssignedAt: time.Now().UTC(),
	}

	id, err := s.store.Add(d)
	if err != nil {
		return fmt.Errorf("Ошибка при создании доставки: %w", err)
	}

	delivery.ID = id
	delivery.AssignedAt = d.AssignedAt
	return nil
}

func (s *DeliveryService) Get(id int) (*models.Delivery, error) {
	delivery, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставки: %w", err)
	}

	return &models.Delivery{
		ID:          delivery.ID,
		ParcelID:    delivery.ParcelID,
		CourierID:   delivery.CourierID,
		Status:      delivery.Status,
		AssignedAt:  delivery.AssignedAt,
		DeliveredAt: delivery.DeliveredAt,
	}, nil
}

func (s *DeliveryService) Update(id int, delivery *models.Delivery) error {
	d := models.Delivery{
		ID:          id,
		ParcelID:    delivery.ParcelID,
		CourierID:   delivery.CourierID,
		Status:      delivery.Status,
		AssignedAt:  delivery.AssignedAt,
		DeliveredAt: delivery.DeliveredAt,
	}

	if err := s.store.Update(d); err != nil {
		return fmt.Errorf("Ошибка при обновлении доставки: %w", err)
	}
	return nil
}

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

	return s.store.Update(delivery)
}

func (s *DeliveryService) GetByParcelID(parcelID int) (*models.Delivery, error) {
	delivery, err := s.store.GetByParcelID(parcelID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставки по ID посылки: %w", err)
	}

	return &models.Delivery{
		ID:          delivery.ID,
		ParcelID:    delivery.ParcelID,
		CourierID:   delivery.CourierID,
		Status:      delivery.Status,
		AssignedAt:  delivery.AssignedAt,
		DeliveredAt: delivery.DeliveredAt,
	}, nil
}

func (s *DeliveryService) GetDeliveriesByCourier(courierID int) ([]models.Delivery, error) {
	deliveries, err := s.store.GetByCourierID(courierID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставок курьера: %w", err)
	}

	return deliveries, nil
}

func (s *DeliveryService) Delete(id int) error {
	err := s.store.Delete(id)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении доставки: %w", err)
	}
	return nil
}

func (s *DeliveryService) AssignDelivery(courierID, parcelID int) (models.Delivery, error) {
	delivery := models.Delivery{
		CourierID:  courierID,
		ParcelID:   parcelID,
		Status:     "assigned",
		AssignedAt: time.Now().UTC(),
	}

	id, err := s.store.Add(delivery)
	if err != nil {
		return models.Delivery{}, fmt.Errorf("Ошибка при создании доставки: %w", err)
	}

	delivery.ID = id
	return delivery, nil
}
