package delivery

import (
	"fmt"
	"time"
)

type DeliveryService struct {
	store *DeliveryStore
}

func NewDeliveryService(store *DeliveryStore) *DeliveryService {
	return &DeliveryService{store: store}
}

type DeliveryStatus struct {
	Assigned   string
	InProgress string
	Delivered  string
	Failed     string
}

var Status = DeliveryStatus{
	Assigned:   "assigned",
	InProgress: "in progress",
	Delivered:  "delivered",
	Failed:     "failed",
}

// Назначить доставку курьеру
func (s *DeliveryService) AssignDelivery(courierID, parcelID int) (Delivery, error) {
	delivery := Delivery{
		CourierID:  courierID,
		ParcelID:   parcelID,
		Status:     Status.Assigned,
		AssignedAt: time.Now().UTC(),
	}
	id, err := s.store.Add(delivery)
	if err != nil {
		return Delivery{}, fmt.Errorf("ошибка при создании доставки: %w", err)
	}
	delivery.ID = id
	return delivery, nil
}

// Завершить доставку
func (s *DeliveryService) CompleteDelivery(deliveryID int) error {
	delivery, err := s.store.Get(deliveryID)
	if err != nil {
		return fmt.Errorf("Ошибка при получении доставки: %w", err)
	}

	if delivery.Status != Status.Assigned && delivery.Status != Status.InProgress {
		return fmt.Errorf("Завершение доставки недоступно для данного статуса: %s", delivery.Status)
	}

	delivery.Status = Status.Delivered
	delivery.DeliveredAt = time.Now().UTC()

	err = s.store.Update(delivery)
	if err != nil {
		return fmt.Errorf("Ошибка при завершении доставки: %w", err)
	}
	return nil
}

// Получить доставку по идентификатору
func (s *DeliveryService) GetDelivery(id int) (Delivery, error) {
	delivery, err := s.store.Get(id)
	if err != nil {
		return Delivery{}, fmt.Errorf("ошибка при получении доставки: %w", err)
	}
	return delivery, nil
}

// Получить все доставки курьера
func (s *DeliveryService) GetDeliveriesByCourier(courierID int) ([]Delivery, error) {
	deliveries, err := s.store.GetByCourierID(courierID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставок курьера: %w", err)
	}
	return deliveries, nil
}

func (s *DeliveryService) String() string {
	return "DeliveryService {работает с доставками}"
}
