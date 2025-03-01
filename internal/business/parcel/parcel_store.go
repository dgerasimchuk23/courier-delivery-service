package parcel

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
	"time"
)

type ParcelStore struct {
	db        *sql.DB
	tableName string
}

func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{
		db:        db,
		tableName: "parcels",
	}
}

func (s *ParcelStore) Add(p models.Parcel) (int, error) {
	createdAt := p.CreatedAt.Format(time.RFC3339) // Преобразование в строку для хранения в базе данных
	query := fmt.Sprintf(`INSERT INTO %s (client_id, address, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id`, s.tableName)
	var id int
	err := s.db.QueryRow(query, p.ClientID, p.Address, p.Status, createdAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("Ошибка при добавлении посылки: %w", err)
	}
	return id, nil
}

func (s *ParcelStore) Get(id int) (*models.Parcel, error) {
	var parcel models.Parcel
	var createdAtStr string

	query := fmt.Sprintf(`SELECT id, client_id, address, status, created_at FROM %s WHERE id = $1`, s.tableName)
	err := s.db.QueryRow(query, id).Scan(&parcel.ID, &parcel.ClientID, &parcel.Address, &parcel.Status, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Посылка с ID %d не найдена", id)
		}
		return nil, fmt.Errorf("Ошибка при получении посылки: %w", err)
	}

	// Преобразование строки в time.Time
	parcel.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("Ошибка преобразования created_at: %w", err)
	}

	return &parcel, nil
}

func (s *ParcelStore) GetByClient(clientID int) ([]models.Parcel, error) {
	query := fmt.Sprintf(`SELECT id, client_id, address, status, created_at FROM %s WHERE client_id = $1`, s.tableName)
	rows, err := s.db.Query(query, clientID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении посылок клиента: %w", err)
	}
	defer rows.Close()

	var parcels []models.Parcel
	for rows.Next() {
		var parcel models.Parcel
		var createdAtStr string

		err := rows.Scan(&parcel.ID, &parcel.ClientID, &parcel.Address, &parcel.Status, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании посылки: %w", err)
		}

		// Преобразование created_at в time.Time
		parcel.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("Ошибка преобразования created_at: %w", err)
		}

		parcels = append(parcels, parcel)
	}

	return parcels, nil
}

func (s *ParcelStore) Update(p models.Parcel) error {
	createdAt := p.CreatedAt.Format(time.RFC3339) // Преобразование в строку
	query := fmt.Sprintf(`UPDATE %s SET client_id = $1, address = $2, status = $3, created_at = $4 WHERE id = $5`, s.tableName)
	_, err := s.db.Exec(query, p.ClientID, p.Address, p.Status, createdAt, p.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении посылки: %w", err)
	}
	return nil
}

func (s *ParcelStore) Delete(id int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, s.tableName)
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении посылки: %w", err)
	}
	return nil
}

func (s *ParcelStore) SetStatus(id int, status string) error {
	query := fmt.Sprintf(`UPDATE %s SET status = $1 WHERE id = $2`, s.tableName)
	_, err := s.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении статуса посылки: %w", err)
	}
	return nil
}

func (s *ParcelStore) SetAddress(id int, address string) error {
	query := fmt.Sprintf(`UPDATE %s SET address = $1 WHERE id = $2`, s.tableName)
	_, err := s.db.Exec(query, address, id)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении адреса посылки: %w", err)
	}
	return nil
}
