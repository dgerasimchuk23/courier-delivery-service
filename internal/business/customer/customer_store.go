package customer

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
	"log"
	"regexp"
)

// Структура для работы с базой данных клиентов
type CustomerStore struct {
	db        *sql.DB
	tableName string
}

func NewCustomerStore(db *sql.DB) *CustomerStore {
	return &CustomerStore{
		db:        db,
		tableName: "customer",
	}
}

// Обрабатывает и логирует ошибки
func logAndReturnError(context string, err error) error {
	if err != nil {
		log.Printf("%s: %v", context, err)
	}
	return err
}

func (s CustomerStore) Add(c models.Customer) (int, error) {
	query := fmt.Sprintf("INSERT INTO %s (name, email, phone) VALUES ($1, $2, $3) RETURNING id", s.tableName)
	var id int
	err := s.db.QueryRow(query, c.Name, c.Email, c.Phone).Scan(&id)
	if err != nil {
		return 0, logAndReturnError("Ошибка добавления клиента", err)
	}
	return id, nil
}

func (s CustomerStore) Get(id int) (models.Customer, error) {
	query := fmt.Sprintf("SELECT id, name, email, phone FROM %s WHERE id = $1", s.tableName)
	row := s.db.QueryRow(query, id)
	c := models.Customer{}
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return c, fmt.Errorf("Клиент с ID %d не найден", id)
		}
		return c, logAndReturnError("Ошибка получения клиента", err)
	}
	return c, nil
}

func (s *CustomerStore) GetByClient(clientID int) ([]models.Customer, error) {
	query := fmt.Sprintf("SELECT id, name, email, phone FROM %s WHERE id = $1", s.tableName)
	rows, err := s.db.Query(query, clientID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при запросе клиентов: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone); err != nil {
			return nil, fmt.Errorf("Ошибка при чтении строки: %w", err)
		}
		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Ошибка при итерации строк: %w", err)
	}
	return customers, nil
}

func (s CustomerStore) Update(c models.Customer) error {
	query := fmt.Sprintf("UPDATE %s SET name = $1, email = $2, phone = $3 WHERE id = $4", s.tableName)
	_, err := s.db.Exec(query, c.Name, c.Email, c.Phone, c.ID)
	return logAndReturnError("Ошибка обновления клиента", err)
}

func (s CustomerStore) Delete(id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", s.tableName)
	_, err := s.db.Exec(query, id)
	return logAndReturnError("Ошибка удаления клиента", err)
}

func ValidateEmail(email string) error {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` // Регулярное выражение для email
	if !regexp.MustCompile(regex).MatchString(email) {
		return fmt.Errorf("Некорректный email: %s", email)
	}
	return nil
}

func ValidatePhone(phone string) error {
	regex := `^[0-9]{10,15}$`
	if !regexp.MustCompile(regex).MatchString(phone) {
		return fmt.Errorf("Некорректный номер телефона: %s", phone)
	}
	return nil
}

func (s *CustomerStore) GetAll() ([]models.Customer, error) {
	query := fmt.Sprintf("SELECT id, name, email, phone FROM %s", s.tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении списка клиентов: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone); err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании клиента: %w", err)
		}
		customers = append(customers, customer)
	}

	return customers, nil
}
