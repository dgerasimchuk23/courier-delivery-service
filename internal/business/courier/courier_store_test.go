package courier

import (
	"database/sql"
	"delivery/internal/models"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Создаем временный файл базы данных для тестов
	tmpFile, err := os.CreateTemp("", "test_courier_*.db")
	if err != nil {
		t.Fatalf("ошибка при создании временного файла БД: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	// Подключаемся к базе данных
	testDB, err := sql.Open("sqlite3", tmpFilePath)
	if err != nil {
		t.Fatalf("ошибка при открытии тестовой БД: %v", err)
	}

	// Создаем таблицу courier для тестов
	if _, err := testDB.Exec(`
		CREATE TABLE courier (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			vehicle_id TEXT,
			status TEXT NOT NULL
		);
	`); err != nil {
		testDB.Close()
		os.Remove(tmpFilePath)
		t.Fatalf("ошибка при создании таблицы courier: %v", err)
	}

	// Возвращаем базу данных и функцию очистки
	cleanup := func() {
		testDB.Close()
		os.Remove(tmpFilePath)
	}

	return testDB, cleanup
}

func TestCourierStore_Add(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewCourierStore(db)

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
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewCourierStore(db)

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
