package delivery

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
)

type DeliveryStore struct {
	db        *sql.DB
	tableName string
}

func NewDeliveryStore(db *sql.DB) *DeliveryStore {
	return &DeliveryStore{
		db:        db,
		tableName: "delivery",
	}
}

// Методы для управления данными доставок в БД

func (s *DeliveryStore) Add(d models.Delivery) (int, error) {
	query := fmt.Sprintf(`INSERT INTO %s (courier_id, parcel_id, status, assigned_at, delivered_at) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`, s.tableName)

	var id int
	var deliveredAt sql.NullTime
	if !d.DeliveredAt.IsZero() {
		deliveredAt = sql.NullTime{Time: d.DeliveredAt, Valid: true}
	}

	err := s.db.QueryRow(query, d.CourierID, d.ParcelID, d.Status, d.AssignedAt, deliveredAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка при добавлении доставки: %w", err)
	}
	return id, nil
}

func (s *DeliveryStore) Get(id int) (models.Delivery, error) {
	query := fmt.Sprintf(`SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM %s WHERE id = $1`, s.tableName)
	row := s.db.QueryRow(query, id)
	d := models.Delivery{}

	var deliveredAt sql.NullTime
	err := row.Scan(&d.ID, &d.CourierID, &d.ParcelID, &d.Status, &d.AssignedAt, &deliveredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return d, fmt.Errorf("Доставка с ID %d не найдена", id)
		}
		return d, fmt.Errorf("Ошибка при получении доставки: %w", err)
	}
	if deliveredAt.Valid {
		d.DeliveredAt = deliveredAt.Time
	}
	return d, nil
}

func (s *DeliveryStore) Update(d models.Delivery) error {
	query := fmt.Sprintf(`UPDATE %s SET courier_id = $1, parcel_id = $2, status = $3, assigned_at = $4, delivered_at = $5 WHERE id = $6`, s.tableName)
	_, err := s.db.Exec(query, d.CourierID, d.ParcelID, d.Status, d.AssignedAt, sql.NullTime{Time: d.DeliveredAt, Valid: !d.DeliveredAt.IsZero()}, d.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении доставки: %w", err)
	}
	return nil
}

func (s *DeliveryStore) Delete(id int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, s.tableName)
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении доставки: %w", err)
	}
	return nil
}

func (s *DeliveryStore) GetByCourierID(courierID int) ([]models.Delivery, error) {
	query := fmt.Sprintf(`SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM %s WHERE courier_id = $1`, s.tableName)
	rows, err := s.db.Query(query, courierID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении доставок по ID курьера: %w", err)
	}
	defer rows.Close()

	var deliveries []models.Delivery
	for rows.Next() {
		d := models.Delivery{}
		var deliveredAt sql.NullTime
		err := rows.Scan(&d.ID, &d.CourierID, &d.ParcelID, &d.Status, &d.AssignedAt, &deliveredAt)
		if err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании данных доставки: %w", err)
		}
		if deliveredAt.Valid {
			d.DeliveredAt = deliveredAt.Time
		}
		deliveries = append(deliveries, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Ошибка при обработке результатов: %w", err)
	}

	return deliveries, nil
}

func (s *DeliveryStore) GetByParcelID(parcelID int) (models.Delivery, error) {
	query := fmt.Sprintf(`SELECT id, parcel_id, courier_id, status, assigned_at, delivered_at FROM %s WHERE parcel_id = $1`, s.tableName)
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
