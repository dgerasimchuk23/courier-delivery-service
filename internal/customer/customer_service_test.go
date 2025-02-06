package customer

import (
	"database/sql"
	"testing"

	"delivery/internal/models"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupCustomerTestDB() *CustomerStore {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	_, _ = db.Exec("DROP TABLE IF EXISTS customer")

	createTable := `
	CREATE TABLE customer (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		phone TEXT
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	return NewCustomerStore(db)
}

func TestAddCustomer(t *testing.T) {
	store := setupCustomerTestDB()

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
