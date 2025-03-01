package courier

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*sql.DB, string, func()) {
	// Подключаемся к базе данных PostgreSQL
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=delivery sslmode=disable"
	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("ошибка при открытии тестовой БД: %v", err)
	}

	// Создаем уникальное имя таблицы для теста
	tableName := fmt.Sprintf("courier_test_%d", time.Now().UnixNano())

	// Создаем таблицу courier для тестов
	if _, err := testDB.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			vehicle_id TEXT,
			status TEXT NOT NULL
		);
	`, tableName)); err != nil {
		testDB.Close()
		t.Fatalf("ошибка при создании таблицы courier: %v", err)
	}

	// Возвращаем базу данных и функцию очистки
	cleanup := func() {
		_, err := testDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
		if err != nil {
			t.Logf("ошибка при удалении таблицы %s: %v", tableName, err)
		}
		testDB.Close()
	}

	return testDB, tableName, cleanup
}

func TestCourierStore_Add(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &CourierStore{
		db:        db,
		tableName: tableName,
	}

	courier := models.Courier{
		Name:      "Тестовый Курьер",
		Phone:     "+79001234567",
		Email:     "test.courier@example.com",
		VehicleID: "A123BC",
		Status:    "available",
	}

	id, err := store.Add(courier)
	if err != nil {
		t.Fatalf("ошибка при добавлении курьера: %v", err)
	}

	if id <= 0 {
		t.Errorf("ожидался положительный ID, получено %d", id)
	}
}

func TestCourierStore_Get(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &CourierStore{
		db:        db,
		tableName: tableName,
	}

	// Добавляем тестового курьера
	courier := models.Courier{
		Name:      "Тестовый Курьер",
		Phone:     "+79001234567",
		Email:     "test.courier@example.com",
		VehicleID: "A123BC",
		Status:    "available",
	}

	id, err := store.Add(courier)
	if err != nil {
		t.Fatalf("ошибка при добавлении курьера: %v", err)
	}

	// Получаем курьера по ID
	retrieved, err := store.Get(id)
	if err != nil {
		t.Fatalf("ошибка при получении курьера: %v", err)
	}

	// Проверяем данные
	if retrieved.ID != id {
		t.Errorf("ожидался ID %d, получено %d", id, retrieved.ID)
	}
	if retrieved.Name != courier.Name {
		t.Errorf("ожидалось имя %s, получено %s", courier.Name, retrieved.Name)
	}
	if retrieved.Phone != courier.Phone {
		t.Errorf("ожидался телефон %s, получено %s", courier.Phone, retrieved.Phone)
	}
	if retrieved.Email != courier.Email {
		t.Errorf("ожидался email %s, получено %s", courier.Email, retrieved.Email)
	}
	if retrieved.VehicleID != courier.VehicleID {
		t.Errorf("ожидался VehicleID %s, получено %s", courier.VehicleID, retrieved.VehicleID)
	}
	if retrieved.Status != courier.Status {
		t.Errorf("ожидался статус %s, получено %s", courier.Status, retrieved.Status)
	}
}
