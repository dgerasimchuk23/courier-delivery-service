package customer

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// setupTestDB создает временную базу данных для тестов
func setupTestDB() *sql.DB {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	// Создание таблицы customer
	createTable := `CREATE TABLE customer (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		phone TEXT
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	return db
}

func TestCustomerStore(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	store := NewCustomerStore(db)

	// Тестирование Add
	customer := Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}
	id, err := store.Add(customer)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Тестирование Get
	storedCustomer, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, customer.Name, storedCustomer.Name)
	require.Equal(t, customer.Email, storedCustomer.Email)
	require.Equal(t, customer.Phone, storedCustomer.Phone)

	// Тестирование Update
	updatedCustomer := Customer{
		ID:    id,
		Name:  "Jane Doe",
		Email: "jane.doe@example.com",
		Phone: "0987654321",
	}
	err = store.Update(updatedCustomer)
	require.NoError(t, err)

	storedUpdatedCustomer, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, updatedCustomer.Name, storedUpdatedCustomer.Name)
	require.Equal(t, updatedCustomer.Email, storedUpdatedCustomer.Email)
	require.Equal(t, updatedCustomer.Phone, storedUpdatedCustomer.Phone)

	// Тестирование Delete
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestCustomerService(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	store := NewCustomerStore(db)
	service := NewCustomerService(store)

	// Тестирование RegisterCustomer
	customer, err := service.RegisterCustomer("John Doe", "john.doe@example.com", "1234567890")
	require.NoError(t, err)
	require.NotZero(t, customer.ID)
	require.Equal(t, "John Doe", customer.Name)
	require.Equal(t, "john.doe@example.com", customer.Email)
	require.Equal(t, "1234567890", customer.Phone)

	// Тестирование GetCustomer
	storedCustomer, err := service.GetCustomer(customer.ID)
	require.NoError(t, err)
	require.Equal(t, customer, storedCustomer)

	// Тестирование UpdateCustomer
	err = service.UpdateCustomer(customer.ID, "Jane Doe", "jane.doe@example.com", "0987654321")
	require.NoError(t, err)

	storedUpdatedCustomer, err := service.GetCustomer(customer.ID)
	require.NoError(t, err)
	require.Equal(t, "Jane Doe", storedUpdatedCustomer.Name)
	require.Equal(t, "jane.doe@example.com", storedUpdatedCustomer.Email)
	require.Equal(t, "0987654321", storedUpdatedCustomer.Phone)

	// Тестирование DeleteCustomer
	err = service.DeleteCustomer(customer.ID)
	require.NoError(t, err)

	_, err = service.GetCustomer(customer.ID)
	require.Error(t, err)
}
