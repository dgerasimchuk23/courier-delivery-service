package parcel

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	// Источник псевдослучайных чисел.
	// Используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// Генерация случайных чисел
	randRange = rand.New(randSource)
)

// Возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339), // time.RFC3339 - форматирование времени в стандартный формат
	}
}

// Временная база данных в памяти
func setupTestDB() *sql.DB {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	// Создание таблицы parcel
	createTable := `CREATE TABLE parcel (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	return db
}

func TestAddGetDelete(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id

	// Получение посылки по идентификатору
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, storedParcel)

	// delete
	// Удаление добавленной посылки, убедитесь в отсутствии ошибки
	err = store.Delete(id)
	require.NoError(t, err) // Ожидание ошибки, так как посылка удалена (err != nil) - посылку больше нельзя получить из БД.

	// Проверка, что посылку больше нельзя получить из базы данных
	_, err = store.Get(id)
	require.Error(t, err)
}

// Проверка обновления адреса
func TestSetAddress(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Проверка добавления посылки
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление адреса
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Проверка обновлённого адреса
	parcelStored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, parcelStored.Address)
}

// Проверка обновлённого статуса посылки
func TestSetStatus(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление статуса
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// Проверка обновлённого статуса
	parcelStored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, parcelStored.Status)
}

// Проверка получения посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// Идентификатор клиента для посылок
	clientID := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = clientID
	}

	// Добавление посылок
	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id

		// Сохранение добавленных посылок в map, чтобы можно было достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// Получение посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	// Проверка, что данные полученных посылок корректны
	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]
		require.True(t, ok, "Посылка с идентификатором %d не найдена в map", parcel.Number)
		require.Equal(t, expectedParcel.Client, parcel.Client)
		require.Equal(t, expectedParcel.Status, parcel.Status)
		require.Equal(t, expectedParcel.Address, parcel.Address)
		require.Equal(t, expectedParcel.CreatedAt, parcel.CreatedAt)
	}
}
