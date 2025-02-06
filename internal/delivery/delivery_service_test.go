package delivery

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"delivery/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestServiceGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := NewDeliveryStore(db)
	service := NewDeliveryService(store)

	// Указание конкретных колонок
	rows := sqlmock.NewRows([]string{"id", "courier_id", "parcel_id", "status", "assigned_at", "delivered_at"}).
		AddRow(1, 1, 2, "assigned", time.Now().UTC(), sql.NullTime{Valid: false})

	mock.ExpectQuery("SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM delivery WHERE id = ?").
		WithArgs(1).
		WillReturnRows(rows)

	delivery, err := service.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, delivery.ID)
	assert.Equal(t, "assigned", delivery.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := NewDeliveryStore(db)
	service := NewDeliveryService(store)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE delivery SET courier_id = ?, parcel_id = ?, status = ?, assigned_at = ?, delivered_at = ? WHERE id = ?")).
		WithArgs(1, 2, "delivered", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	delivery := &models.Delivery{
		ID:          1,
		CourierID:   1,
		ParcelID:    2,
		Status:      "delivered",
		AssignedAt:  time.Now().UTC(),
		DeliveredAt: time.Now().UTC(),
	}

	err = service.Update(1, delivery)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
