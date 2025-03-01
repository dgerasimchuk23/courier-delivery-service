package parcel

import (
	"testing"
	"time"

	"delivery/internal/business/models"

	"github.com/stretchr/testify/require"
)

func TestParcelService(t *testing.T) {
	store := setupParcelTestDB()
	service := NewParcelService(store)

	// Добавление посылки
	parcel := models.Parcel{
		ClientID:  123,
		Status:    models.ParcelStatusRegistered,
		Address:   "123 Main Street",
		CreatedAt: time.Now().UTC(),
	}
	err := service.Register(&parcel)
	require.NoError(t, err)
	require.NotZero(t, parcel.ID)

	// Получение посылки
	fetchedParcel, err := service.Get(parcel.ID)
	require.NoError(t, err)
	require.Equal(t, parcel.ClientID, fetchedParcel.ClientID)
	require.Equal(t, parcel.Address, fetchedParcel.Address)

	// Обновление статуса
	err = service.UpdateStatus(parcel.ID, models.ParcelStatusSent)
	require.NoError(t, err)

	// Проверка обновленного статуса
	updatedParcel, err := service.Get(parcel.ID)
	require.NoError(t, err)
	require.Equal(t, models.ParcelStatusSent, updatedParcel.Status)

	// Удаление посылки
	err = service.Delete(parcel.ID)
	require.NoError(t, err)

	// Проверка, что посылка удалена
	_, err = service.Get(parcel.ID)
	require.Error(t, err)
}
