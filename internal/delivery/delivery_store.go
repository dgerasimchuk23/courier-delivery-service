package delivery

import (
	"database/sql"
	"fmt"
	"time"
)

type Delivery struct {
	ID          int
	CourierID   int
	ParcelID    int
	Status      string
	AssignedAt  time.Time
	DeliveredAt time.Time
}

type DeliveryStore struct {
	db *sql.DB
}

func SetupDeliveriesDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./internal/delivery/deliveries.db")
	if err != nil {
		fmt.Printf("Не удалось подключиться к базе данных deliveries: %v", err)
		return nil, err
	}

	createTable := `CREATE TABLE IF NOT EXISTS delivery(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		courier_id INTEGER,
		parcel_id INTEGER,
		status TEXT,
		assigned_at DATETIME,
		delivered_at DATETIME DEFAULT NULL);`
	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Printf("Ошибка при создании таблицы delivery: %v", err)
		return nil, err
	}

	fmt.Println("База данных deliveries.db и таблица delivery успешно инициализированы")
	return db, nil
}

func NewDeliveryStore(db *sql.DB) *DeliveryStore {
	return &DeliveryStore{db: db}
}

// Методы для управления данными доставок в БД

func (s DeliveryStore) Add(d Delivery) (int, error) {
	query := `INSERT INTO delivery (courier_id, parcel_id, status, assigned_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, d.CourierID, d.ParcelID, d.Status, d.AssignedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s DeliveryStore) Get(id int) (Delivery, error) {
	query := `SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM delivery WHERE id = ?`
	row := s.db.QueryRow(query, id)
	d := Delivery{}
	err := row.Scan(&d.ID, &d.CourierID, &d.ParcelID, &d.Status, &d.AssignedAt, &d.DeliveredAt)
	if err != nil {
		return d, err
	}
	return d, nil
}

func (s DeliveryStore) Update(d Delivery) error {
	query := `UPDATE delivery SET courier_id = ?, parcel_id = ?, status = ?, assigned_at = ?, delivered_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, d.CourierID, d.ParcelID, d.Status, d.AssignedAt, d.DeliveredAt, d.ID)
	return err
}

func (s DeliveryStore) Delete(id int) error {
	query := `DELETE FROM delivery WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s DeliveryStore) GetByCourierID(courierID int) ([]Delivery, error) {
	query := `SELECT id, courier_id, parcel_id, status, assigned_at, delivered_at FROM delivery WHERE courier_id = ?`
	rows, err := s.db.Query(query, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		d := Delivery{}
		err := rows.Scan(&d.ID, &d.CourierID, &d.ParcelID, &d.Status, &d.AssignedAt, &d.DeliveredAt)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return deliveries, nil
}
