package delivery

import (
	"context"
	"delivery/internal/business/models"
	"delivery/internal/cache"
	"fmt"
	"log"
	"time"
)

type DeliveryService struct {
	store       *DeliveryStore
	cacheClient *cache.RedisClient
}

func NewDeliveryService(store *DeliveryStore) *DeliveryService {
	return &DeliveryService{store: store}
}

// WithCache добавляет клиент кэширования к сервису
func (s *DeliveryService) WithCache(cacheClient *cache.RedisClient) *DeliveryService {
	s.cacheClient = cacheClient
	return s
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

	// Инвалидируем кэш списка доставок
	if s.cacheClient != nil {
		ctx := context.Background()
		if err := s.cacheClient.Delete(ctx, "deliveries:list"); err != nil {
			log.Printf("Ошибка при удалении кэша списка доставок: %v", err)
		}
	}

	return nil
}

func (s *DeliveryService) Get(id int) (*models.Delivery, error) {
	// Если кэширование включено, пытаемся получить из кэша
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:%d", id)

		var delivery models.Delivery
		err := s.cacheClient.GetJSON(ctx, cacheKey, &delivery)
		if err == nil {
			return &delivery, nil
		}
	}

	// Получаем из БД
	delivery, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставки: %w", err)
	}

	result := &models.Delivery{
		ID:          delivery.ID,
		ParcelID:    delivery.ParcelID,
		CourierID:   delivery.CourierID,
		Status:      delivery.Status,
		AssignedAt:  delivery.AssignedAt,
		DeliveredAt: delivery.DeliveredAt,
	}

	// Если кэширование включено, сохраняем в кэш
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:%d", id)
		if err := s.cacheClient.SetJSON(ctx, cacheKey, result, 30*time.Minute); err != nil {
			log.Printf("Ошибка при сохранении доставки в кэш: %v", err)
		}
	}

	return result, nil
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

	// Обновляем кэш
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:%d", id)

		// Удаляем старые данные из кэша
		if err := s.cacheClient.Delete(ctx, cacheKey); err != nil {
			log.Printf("Ошибка при удалении кэша доставки: %v", err)
		}

		// Инвалидируем кэш списка доставок
		if err := s.cacheClient.Delete(ctx, "deliveries:list"); err != nil {
			log.Printf("Ошибка при удалении кэша списка доставок: %v", err)
		}

		// Сохраняем обновленные данные в кэш
		if err := s.cacheClient.SetJSON(ctx, cacheKey, d, 30*time.Minute); err != nil {
			log.Printf("Ошибка при сохранении доставки в кэш: %v", err)
		}
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

	// Обновляем кэш
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:%d", deliveryID)

		// Удаляем старые данные из кэша
		if err := s.cacheClient.Delete(ctx, cacheKey); err != nil {
			log.Printf("Ошибка при удалении кэша доставки: %v", err)
		}

		// Инвалидируем кэш списка доставок
		if err := s.cacheClient.Delete(ctx, "deliveries:list"); err != nil {
			log.Printf("Ошибка при удалении кэша списка доставок: %v", err)
		}

		// Сохраняем обновленные данные в кэш
		if err := s.cacheClient.SetJSON(ctx, cacheKey, delivery, 30*time.Minute); err != nil {
			log.Printf("Ошибка при сохранении доставки в кэш: %v", err)
		}
	}

	return s.store.Update(delivery)
}

func (s *DeliveryService) GetByParcelID(parcelID int) (*models.Delivery, error) {
	// Если кэширование включено, пытаемся получить из кэша
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:parcel:%d", parcelID)

		var delivery models.Delivery
		err := s.cacheClient.GetJSON(ctx, cacheKey, &delivery)
		if err == nil {
			return &delivery, nil
		}
	}

	// Получаем из БД
	delivery, err := s.store.GetByParcelID(parcelID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставки по ID посылки: %w", err)
	}

	result := &models.Delivery{
		ID:          delivery.ID,
		ParcelID:    delivery.ParcelID,
		CourierID:   delivery.CourierID,
		Status:      delivery.Status,
		AssignedAt:  delivery.AssignedAt,
		DeliveredAt: delivery.DeliveredAt,
	}

	// Если кэширование включено, сохраняем в кэш
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:parcel:%d", parcelID)
		s.cacheClient.SetJSON(ctx, cacheKey, result, 30*time.Minute)
	}

	return result, nil
}

func (s *DeliveryService) GetDeliveriesByCourier(courierID int) ([]models.Delivery, error) {
	// Если кэширование включено, пытаемся получить из кэша
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("deliveries:courier:%d", courierID)

		var deliveries []models.Delivery
		err := s.cacheClient.GetJSON(ctx, cacheKey, &deliveries)
		if err == nil {
			return deliveries, nil
		}
	}

	// Получаем из БД
	deliveries, err := s.store.GetByCourierID(courierID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставок курьера: %w", err)
	}

	// Если кэширование включено, сохраняем в кэш
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("deliveries:courier:%d", courierID)
		s.cacheClient.SetJSON(ctx, cacheKey, deliveries, 15*time.Minute)
	}

	return deliveries, nil
}

func (s *DeliveryService) Delete(id int) error {
	err := s.store.Delete(id)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении доставки: %w", err)
	}

	// Удаляем из кэша
	if s.cacheClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("delivery:%d", id)

		// Удаляем данные из кэша
		s.cacheClient.Delete(ctx, cacheKey)

		// Инвалидируем кэш списка доставок
		s.cacheClient.Delete(ctx, "deliveries:list")
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

	// Инвалидируем кэш списка доставок
	if s.cacheClient != nil {
		ctx := context.Background()
		s.cacheClient.Delete(ctx, "deliveries:list")
	}

	return delivery, nil
}
