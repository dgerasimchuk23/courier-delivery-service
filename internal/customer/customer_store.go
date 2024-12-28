package customer

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
)

func SetupCustomersDB() *sql.DB {

	db, err := sql.Open("sqlite", "./internal/customer/customers.db")
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных customers: %v", err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS customer (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		phone TEXT
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы customer: %v", err)
	}

	fmt.Println("База данных customers.db и таблица customer успешно инициализированы")
	return db
}

// Структура для хранения информации о клиенте
type Customer struct {
	ID    int
	Name  string
	Email string
	Phone string
}

// Структура для работы с базой данных клиентов
type CustomerStore struct {
	db *sql.DB
}

func NewCustomerStore(db *sql.DB) *CustomerStore {
	return &CustomerStore{db: db}
}

func (s CustomerStore) Add(c Customer) (int, error) {
	query := `INSERT INTO customer (name, email, phone) VALUES (?, ?, ?)`
	result, err := s.db.Exec(query, c.Name, c.Email, c.Phone)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s CustomerStore) Get(id int) (Customer, error) {
	query := `SELECT id, name, email, phone FROM customer WHERE id = ?`
	row := s.db.QueryRow(query, id)
	c := Customer{}
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (s CustomerStore) Update(c Customer) error {
	query := `UPDATE customer SET name = ?, email = ?, phone = ? WHERE id = ?`
	_, err := s.db.Exec(query, c.Name, c.Email, c.Phone, c.ID)
	return err
}

func (s CustomerStore) Delete(id int) error {
	query := `DELETE FROM customer WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func ValidateEmail(email string) error {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` //Регулярное выражение, описывает формат допустимого email
	if !regexp.MustCompile(regex).MatchString(email) {
		return fmt.Errorf("некорректный email: %s", email)
	}
	return nil
}

func ValidatePhone(phone string) error {
	regex := `^[0-9]{10,15}$`
	if !regexp.MustCompile(regex).MatchString(phone) {
		return fmt.Errorf("некорректный номер телефона: %s", phone)
	}
	return nil
}
