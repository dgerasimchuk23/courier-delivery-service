package parcel

import (
	"database/sql"
	"fmt"
	"log"
)

// Подключение к БД - parcels.db
func SetupParcelsDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./internal/parcel/parcels.db")
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных parcels: %v", err)
		return nil, err
	}

	createTable := `CREATE TABLE IF NOT EXISTS parcel (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Printf("Ошибка при создании таблицы parcel: %v", err)
		return nil, err
	}

	fmt.Println("База данных parcels.db и таблица parcel успешно инициализированы")
	return db, nil
}

// Структура посылки
type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

// Структура хранилища данных для работы с БД посылок
type ParcelStore struct {
	db *sql.DB
}

// Создание объекта хранилища данных
func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{db: db}
}

// Добавление новой посылки в базу данных
func (s ParcelStore) Add(p Parcel) (int, error) {
	query := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	// Получение идентификатора добавленной записи
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Получение данных о посылке по её идентификатору
func (s ParcelStore) Get(number int) (Parcel, error) {
	query := `SELECT id, client, status, address, created_at FROM parcel WHERE id = ?`
	row := s.db.QueryRow(query, number)
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	return p, nil
}

// Получение списка посылок клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := `SELECT id, client, status, address, created_at FROM parcel WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parceles []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		parceles = append(parceles, p)
	}
	return parceles, nil
}

// Обновление статуса посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	query := `UPDATE parcel SET status = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, number)
	return err
}

// Обновление адреса посылки (доступно только для статуса "registered")
func (s ParcelStore) SetAddress(number int, address string) error {
	query := `UPDATE parcel SET address = ? WHERE id = ? AND status = ?`
	result, err := s.db.Exec(query, address, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	// Проверка, было ли обновлено хотя бы одно значение
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Изменение адреса доступно только для посылок со статусом %s", ParcelStatusRegistered)
	}
	return nil
}

// Удаление посылки (доступно только для статуса "registered")
func (s ParcelStore) Delete(number int) error {
	query := `DELETE FROM parcel WHERE id = ? AND status = ?`
	result, err := s.db.Exec(query, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	// Проверка, было ли удалено хотя бы одно значение
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Удаление доступно только для посылок со статусом %s", ParcelStatusRegistered)
	}
	return nil
}
