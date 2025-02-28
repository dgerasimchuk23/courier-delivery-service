package courier

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
)

// Убедимся, что CourierStore реализует интерфейс CourierStorer
var _ CourierStorer = (*CourierStore)(nil)

type CourierStore struct {
	db *sql.DB
}

func NewCourierStore(db *sql.DB) *CourierStore {
	return &CourierStore{db: db}
}

func (s *CourierStore) Add(c models.Courier) (int, error) {
	query := `INSERT INTO courier (name, phone, email, vehicle_id, status) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, c.Name, c.Phone, c.Email, c.VehicleID, c.Status)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления курьера: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID последнего добавленного курьера: %w", err)
	}
	return int(id), nil
}

func (s *CourierStore) Get(id int) (models.Courier, error) {
	query := `SELECT id, name, phone, email, vehicle_id, status FROM courier WHERE id = ?`
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
	query := `UPDATE courier SET name = ?, phone = ?, email = ?, vehicle_id = ?, status = ? WHERE id = ?`
	_, err := s.db.Exec(query, courier.Name, courier.Phone, courier.Email, courier.VehicleID, courier.Status, courier.ID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении курьера: %w", err)
	}
	return nil
}

func (s *CourierStore) Delete(id int) error {
	query := `DELETE FROM courier WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении курьера: %w", err)
	}
	return nil
}

func (s *CourierStore) GetAll() ([]models.Courier, error) {
	query := `SELECT id, name, phone, email, vehicle_id, status FROM courier`
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
	query := `SELECT id, name, phone, email, vehicle_id, status FROM courier WHERE status = 'available'`
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
