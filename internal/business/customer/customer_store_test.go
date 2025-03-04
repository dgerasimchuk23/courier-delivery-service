package customer

import (
	"fmt"
	"testing"

	"delivery/internal/business/models"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomerStore(t *testing.T) {
	store := setupCustomerTestDB()
	defer func() {
		_, _ = store.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", store.tableName))
	}()

	// Добавление клиента
	customer := models.Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}
	id, err := store.Add(customer)
	require.NoError(t, err)

	// Получение клиента
	storedCustomer, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, customer.Name, storedCustomer.Name)
	require.Equal(t, customer.Email, storedCustomer.Email)

	// Обновление клиента
	customer.ID = id
	customer.Name = "Jane Doe"
	customer.Email = "jane.doe@example.com"
	err = store.Update(customer)
	require.NoError(t, err)

	// Проверка обновления
	updatedCustomer, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, "Jane Doe", updatedCustomer.Name)
	require.Equal(t, "jane.doe@example.com", updatedCustomer.Email)

	// Удаление клиента
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}
