package parcel

import (
	"delivery/internal/business/models"
	"fmt"
	"time"
)

type ParcelService struct {
	store *ParcelStore
}

func NewParcelService(store *ParcelStore) *ParcelService {
	return &ParcelService{store: store}
}

func (s *ParcelService) Register(parcel *models.Parcel) error {
	p := models.Parcel{
		ClientID:  parcel.ClientID,
		Address:   parcel.Address,
		Status:    "registered",
		CreatedAt: time.Now().UTC(),
	}

	id, err := s.store.Add(p)
	if err != nil {
		return fmt.Errorf("Ошибка при регистрации посылки: %w", err)
	}

	parcel.ID = id
	return nil
}

func (s *ParcelService) Get(id int) (*models.Parcel, error) {
	parcel, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("parcel not found: %w", err)
	}
	return &models.Parcel{
		ID:       parcel.ID,
		ClientID: parcel.ClientID,
		Address:  parcel.Address,
		Status:   parcel.Status,
	}, nil
}

func (s *ParcelService) List(clientID int) ([]models.Parcel, error) {
	parcels, err := s.store.GetByClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении списка посылок: %w", err)
	}

	var result []models.Parcel
	for _, parcel := range parcels {
		result = append(result, models.Parcel{
			ID:       parcel.ID,
			ClientID: parcel.ClientID,
			Address:  parcel.Address,
			Status:   parcel.Status,
		})
	}
	return result, nil
}

func (s *ParcelService) Update(id int, parcel *models.Parcel) error {
	p := models.Parcel{
		ID:        id,
		ClientID:  parcel.ClientID,
		Address:   parcel.Address,
		Status:    parcel.Status,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.store.Update(p); err != nil {
		return fmt.Errorf("Ошибка при обновлении посылки: %w", err)
	}

	return nil
}

func (s *ParcelService) UpdateStatus(id int, status string) error {
	return s.store.SetStatus(id, status)
}

func (s *ParcelService) UpdateAddress(id int, address string) error {
	return s.store.SetAddress(id, address)
}

func (s *ParcelService) Delete(id int) error {
	return s.store.Delete(id)
}
