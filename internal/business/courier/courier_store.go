package courier

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
)

// Убедимся, что CourierStore реализует интерфейс CourierStorer
var _ CourierStorer = (*CourierStore)(nil)

type CourierStore struct {
	db        *sql.DB
	tableName string
}

func NewCourierStore(db *sql.DB) *CourierStore {
	return &CourierStore{
		db:        db,
		tableName: "courier",
	}
}

func (s *CourierStore) Add(c models.Courier) (int, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, phone, email, vehicle_id, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, s.tableName)

	var id int
	err := s.db.QueryRow(query, c.Name, c.Phone, c.Email, c.VehicleID, c.Status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка при добавлении курьера: %w", err)
	}

	return id, nil
}

func (s *CourierStore) Get(id int) (models.Courier, error) {
	query := fmt.Sprintf(`SELECT id, name, phone, email, vehicle_id, status FROM %s WHERE id = $1`, s.tableName)
	row := s.db.QueryRow(query, id)

	var courier models.Courier
	var vehicleID sql.NullString
	err := row.Scan(&courier.ID, &courier.Name, &courier.Phone, &courier.Email, &vehicleID, &courier.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return courier, fmt.Errorf("курьер с ID %d не найден", id)
		}
		return courier, fmt.Errorf("ошибка при получении курьера: %w", err)
	}

	if vehicleID.Valid {
		courier.VehicleID = vehicleID.String
	}

	return courier, nil
}

func (s *CourierStore) Update(courier models.Courier) error {
	query := fmt.Sprintf(`UPDATE %s SET name = $1, phone = $2, email = $3, vehicle_id = $4, status = $5 WHERE id = $6`, s.tableName)
	_, err := s.db.Exec(query, courier.Name, courier.Phone, courier.Email, courier.VehicleID, courier.Status, courier.ID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении курьера: %w", err)
	}
	return nil
}

func (s *CourierStore) Delete(id int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, s.tableName)
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении курьера: %w", err)
	}
	return nil
}

func (s *CourierStore) GetAll() ([]models.Courier, error) {
	query := fmt.Sprintf(`SELECT id, name, phone, email, vehicle_id, status FROM %s`, s.tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка курьеров: %w", err)
	}
	defer rows.Close()

	var couriers []models.Courier
	for rows.Next() {
		var courier models.Courier
		var vehicleID sql.NullString
		err := rows.Scan(&courier.ID, &courier.Name, &courier.Phone, &courier.Email, &vehicleID, &courier.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных курьера: %w", err)
		}

		if vehicleID.Valid {
			courier.VehicleID = vehicleID.String
		}

		couriers = append(couriers, courier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %w", err)
	}

	return couriers, nil
}

func (s *CourierStore) GetAvailableCouriers() ([]models.Courier, error) {
	query := fmt.Sprintf(`SELECT id, name, phone, email, vehicle_id, status FROM %s WHERE status = 'available'`, s.tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении доступных курьеров: %w", err)
	}
	defer rows.Close()

	var couriers []models.Courier
	for rows.Next() {
		var courier models.Courier
		var vehicleID sql.NullString
		err := rows.Scan(&courier.ID, &courier.Name, &courier.Phone, &courier.Email, &vehicleID, &courier.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных курьера: %w", err)
		}

		if vehicleID.Valid {
			courier.VehicleID = vehicleID.String
		}

		couriers = append(couriers, courier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %w", err)
	}

	return couriers, nil
}
