package delivery

import (
	"database/sql"
	"fmt"
)

// Инициальзация БД курьеров
func SetupCouriersDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./internal/delivery/couriers.db")
	if err != nil {
		fmt.Printf("Не удалось подключиться к базе данных couriers: %v", err)
		return nil, err
	}

	createTable := `CREATE TABLE IF NOT EXISTS courier (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		phone TEXT,
		capacity INTEGER
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Printf("Ошибка при создании таблицы courier: %v", err)
		return nil, err
	}
	fmt.Println("База данных couriers.db и таблица courier успешно инициализированы")
	return db, nil
}

type Courier struct {
	ID       int
	Name     string
	Phone    string
	Capacity int // Кол-во посылок, которые курьер может доставить за 1 маршрут
}

type CourierStore struct {
	db *sql.DB
}

func NewCourierStore(db *sql.DB) *CourierStore {
	return &CourierStore{db: db}
}

// Методы для управления данными курьеров в БД

func (s CourierStore) Add(c Courier) (int, error) {
	query := `INSERT INTO courier (name, phone, capacity) VALUES (?, ?, ?)`
	result, err := s.db.Exec(query, c.Name, c.Phone, c.Capacity)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s CourierStore) Get(id int) (Courier, error) {
	query := `SELECT id, name, phone, capacity FROM courier WHERE id = ?`
	row := s.db.QueryRow(query, id)
	c := Courier{}
	err := row.Scan(&c.ID, &c.Name, &c.Phone, &c.Capacity)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (s CourierStore) Update(c Courier) error {
	query := `UPDATE courier SET name = ?, phone = ?, capacity = ? WHERE id = ?`
	_, err := s.db.Exec(query, c.Name, c.Phone, c.Capacity, c.ID)
	return err
}

func (s CourierStore) Delete(id int) error {
	query := `DELETE FROM courier WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}
