package delivery

import (
	"database/sql"
	"delivery/internal/business/models"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*sql.DB, string, func()) {
	// Подключаемся к базе данных PostgreSQL
	host := getEnv("DB_HOST", "db")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "delivery")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("ошибка при открытии тестовой БД: %v", err)
	}

	// Создаем уникальное имя таблицы для теста
	tableName := fmt.Sprintf("delivery_test_%d", time.Now().UnixNano())

	// Создаем таблицу delivery для тестов
	if _, err := testDB.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			courier_id INTEGER NOT NULL,
			parcel_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			assigned_at TIMESTAMP NOT NULL,
			delivered_at TIMESTAMP
		);
	`, tableName)); err != nil {
		testDB.Close()
		t.Fatalf("ошибка при создании таблицы delivery: %v", err)
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

// Вспомогательная функция для получения переменных окружения
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func TestDeliveryStore_Add(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &DeliveryStore{
		db:        db,
		tableName: tableName,
	}

	now := time.Now().UTC()
	delivery := models.Delivery{
		CourierID:  1,
		ParcelID:   2,
		Status:     "assigned",
		AssignedAt: now,
	}

	id, err := store.Add(delivery)
	assert.NoError(t, err)
	assert.Greater(t, id, 0)

	// Проверяем, что запись добавлена
	var count int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = $1", tableName), id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeliveryStore_Get(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &DeliveryStore{
		db:        db,
		tableName: tableName,
	}

	now := time.Now().UTC()
	delivery := models.Delivery{
		CourierID:  1,
		ParcelID:   2,
		Status:     "assigned",
		AssignedAt: now,
	}

	id, err := store.Add(delivery)
	assert.NoError(t, err)

	// Получаем доставку по ID
	retrieved, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, id, retrieved.ID)
	assert.Equal(t, delivery.CourierID, retrieved.CourierID)
	assert.Equal(t, delivery.ParcelID, retrieved.ParcelID)
	assert.Equal(t, delivery.Status, retrieved.Status)
	assert.WithinDuration(t, delivery.AssignedAt, retrieved.AssignedAt, time.Second)
	assert.True(t, retrieved.DeliveredAt.IsZero())
}

func TestDeliveryStore_Update(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &DeliveryStore{
		db:        db,
		tableName: tableName,
	}

	now := time.Now().UTC()
	delivery := models.Delivery{
		CourierID:  1,
		ParcelID:   2,
		Status:     "assigned",
		AssignedAt: now,
	}

	id, err := store.Add(delivery)
	assert.NoError(t, err)

	// Обновляем доставку
	delivery.ID = id
	delivery.Status = "delivered"
	delivery.DeliveredAt = time.Now().UTC()

	err = store.Update(delivery)
	assert.NoError(t, err)

	// Проверяем обновление
	retrieved, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, "delivered", retrieved.Status)
	assert.False(t, retrieved.DeliveredAt.IsZero())
}

func TestDeliveryStore_Delete(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &DeliveryStore{
		db:        db,
		tableName: tableName,
	}

	now := time.Now().UTC()
	delivery := models.Delivery{
		CourierID:  1,
		ParcelID:   2,
		Status:     "assigned",
		AssignedAt: now,
	}

	id, err := store.Add(delivery)
	assert.NoError(t, err)

	// Удаляем доставку
	err = store.Delete(id)
	assert.NoError(t, err)

	// Проверяем, что запись удалена
	var count int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = $1", tableName), id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeliveryStore_GetByCourierID(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &DeliveryStore{
		db:        db,
		tableName: tableName,
	}

	now := time.Now().UTC()
	delivery1 := models.Delivery{
		CourierID:  1,
		ParcelID:   2,
		Status:     "assigned",
		AssignedAt: now,
	}

	delivery2 := models.Delivery{
		CourierID:   1,
		ParcelID:    3,
		Status:      "delivered",
		AssignedAt:  now,
		DeliveredAt: now.Add(time.Hour),
	}

	_, err := store.Add(delivery1)
	assert.NoError(t, err)
	_, err = store.Add(delivery2)
	assert.NoError(t, err)

	// Получаем доставки по ID курьера
	deliveries, err := store.GetByCourierID(1)
	assert.NoError(t, err)
	assert.Len(t, deliveries, 2)
}

func TestDeliveryStore_GetByParcelID(t *testing.T) {
	db, tableName, cleanup := setupTestDB(t)
	defer cleanup()

	store := &DeliveryStore{
		db:        db,
		tableName: tableName,
	}

	now := time.Now().UTC()
	delivery := models.Delivery{
		CourierID:  1,
		ParcelID:   2,
		Status:     "assigned",
		AssignedAt: now,
	}

	_, err := store.Add(delivery)
	assert.NoError(t, err)

	// Получаем доставку по ID посылки
	retrieved, err := store.GetByParcelID(2)
	assert.NoError(t, err)
	assert.Equal(t, delivery.ParcelID, retrieved.ParcelID)
	assert.Equal(t, delivery.CourierID, retrieved.CourierID)
}
