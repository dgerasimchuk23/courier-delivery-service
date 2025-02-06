package parcel

import (
	"testing"
	"time"

	"delivery/internal/models"

	"github.com/stretchr/testify/require"
)

func TestParcelStore(t *testing.T) {
	store := setupParcelTestDB()

	// Добавление посылки
	parcel := models.Parcel{
		ClientID:  123,
		Status:    models.ParcelStatusRegistered,
		Address:   "123 Main Street",
		CreatedAt: time.Now().UTC(),
	}
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Получение посылки
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.ClientID, storedParcel.ClientID)
	require.Equal(t, parcel.Address, storedParcel.Address)

	// Обновление статуса
	err = store.SetStatus(id, models.ParcelStatusSent)
	require.NoError(t, err)

	// Проверка обновленного статуса
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, models.ParcelStatusSent, updatedParcel.Status)

	// Удаление посылки
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверка, что посылка удалена
	_, err = store.Get(id)
	require.Error(t, err)
}
