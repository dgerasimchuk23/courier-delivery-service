package customer

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"delivery/internal/business/models"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func setupCustomerTestDB() *CustomerStore {
	// Подключение к PostgreSQL
	host := getEnv("DB_HOST", "db")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "delivery")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	// Создаем уникальную таблицу для теста
	tableName := fmt.Sprintf("customer_test_%d", time.Now().UnixNano())

	_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))

	createTable := fmt.Sprintf(`
	CREATE TABLE %s (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		phone TEXT
	);`, tableName)

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	// Переопределяем SQL-запросы для работы с тестовой таблицей
	store := NewCustomerStore(db)
	store.tableName = tableName
	return store
}

func TestAddCustomer(t *testing.T) {
	store := setupCustomerTestDB()
	defer func() {
		_, _ = store.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", store.tableName))
	}()

	customer := models.Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}
	id, err := store.Add(customer)
	require.NoError(t, err)
	require.NotZero(t, id)
}

func TestGetCustomer(t *testing.T) {
	store := setupCustomerTestDB()
	defer func() {
		_, _ = store.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", store.tableName))
	}()

	customer := models.Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}
	id, err := store.Add(customer)
	require.NoError(t, err)

	fetchedCustomer, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, customer.Name, fetchedCustomer.Name)
	require.Equal(t, customer.Email, fetchedCustomer.Email)
	require.Equal(t, customer.Phone, fetchedCustomer.Phone)
}

func TestUpdateCustomer(t *testing.T) {
	store := setupCustomerTestDB()
	defer func() {
		_, _ = store.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", store.tableName))
	}()

	customer := models.Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}
	id, err := store.Add(customer)
	require.NoError(t, err)

	updatedCustomer := models.Customer{
		ID:    id,
		Name:  "Jane Doe",
		Email: "jane.doe@example.com",
		Phone: "0987654321",
	}
	err = store.Update(updatedCustomer)
	require.NoError(t, err)

	fetchedCustomer, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, "Jane Doe", fetchedCustomer.Name)
	require.Equal(t, "jane.doe@example.com", fetchedCustomer.Email)
	require.Equal(t, "0987654321", fetchedCustomer.Phone)
}

func TestDeleteCustomer(t *testing.T) {
	store := setupCustomerTestDB()
	defer func() {
		_, _ = store.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", store.tableName))
	}()

	customer := models.Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}
	id, err := store.Add(customer)
	require.NoError(t, err)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

// Вспомогательная функция для получения переменных окружения
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
