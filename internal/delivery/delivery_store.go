package delivery

import (
	"database/sql"
	"delivery/internal/models"
	"fmt"
)

type DeliveryStore struct {
	db *sql.DB
}

func NewDeliveryStore(db *sql.DB) *DeliveryStore {
	return &DeliveryStore{db: db}
}

// Методы для управления данными доставок в БД

func (s DeliveryStore) Add(d models.Delivery) (int, error) {
	query := `INSERT INTO delivery (courier_id, parcel_id, status, assigned_at, delivered_at) 
              VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, d.CourierID, d.ParcelID, d.Status, d.AssignedAt, sql.NullTime{Valid: false})
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s DeliveryStore) Get(id int) (models.Delivery, error) {
	query := `SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM delivery WHERE id = ?`
	row := s.db.QueryRow(query, id)
	d := models.Delivery{}

	var deliveredAt sql.NullTime
	err := row.Scan(&d.ID, &d.CourierID, &d.ParcelID, &d.Status, &d.AssignedAt, &deliveredAt)
	if err != nil {
		return d, err
	}
	if deliveredAt.Valid {
		d.DeliveredAt = deliveredAt.Time
	}
	return d, nil
}

func (s DeliveryStore) Update(d models.Delivery) error {
	query := `UPDATE delivery SET courier_id = ?, parcel_id = ?, status = ?, assigned_at = ?, delivered_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, d.CourierID, d.ParcelID, d.Status, d.AssignedAt, sql.NullTime{Time: d.DeliveredAt, Valid: !d.DeliveredAt.IsZero()}, d.ID)
	return err
}

func (s DeliveryStore) Delete(id int) error {
	query := `DELETE FROM delivery WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s DeliveryStore) GetByCourierID(courierID int) ([]models.Delivery, error) {
	query := `SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM delivery WHERE courier_id = ?`
	rows, err := s.db.Query(query, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []models.Delivery
	for rows.Next() {
		d := models.Delivery{}
		var deliveredAt sql.NullTime
		err := rows.Scan(&d.ID, &d.CourierID, &d.ParcelID, &d.Status, &d.AssignedAt, &deliveredAt)
		if err != nil {
			return nil, err
		}
		if deliveredAt.Valid {
			d.DeliveredAt = deliveredAt.Time
		}
		deliveries = append(deliveries, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return deliveries, nil
}

func (s *DeliveryStore) GetByParcelID(parcelID int) (models.Delivery, error) {
	query := `SELECT id, parcel_id, courier_id, status, assigned_at, delivered_at FROM delivery WHERE parcel_id = ?`
	row := s.db.QueryRow(query, parcelID)

	var delivery models.Delivery
	var deliveredAt sql.NullTime
	err := row.Scan(&delivery.ID, &delivery.ParcelID, &delivery.CourierID, &delivery.Status, &delivery.AssignedAt, &deliveredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return delivery, fmt.Errorf("Доставка с ParcelID %d не найдена", parcelID)
		}
		return delivery, fmt.Errorf("Ошибка при получении доставки по ParcelID: %w", err)
	}
	if deliveredAt.Valid {
		delivery.DeliveredAt = deliveredAt.Time
	}

	return delivery, nil
}
