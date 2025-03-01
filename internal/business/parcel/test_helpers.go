package parcel

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func setupParcelTestDB() *ParcelStore {
	// Подключение к PostgreSQL
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=delivery_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	// Создаем уникальную таблицу для теста
	tableName := fmt.Sprintf("parcel_test_%d", time.Now().UnixNano())

	_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))

	createTable := fmt.Sprintf(`
    CREATE TABLE %s (
        id SERIAL PRIMARY KEY,
        client_id INTEGER,
        status TEXT,
        address TEXT,
        created_at TIMESTAMP
    );`, tableName)

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	// Переопределяем SQL-запросы для работы с тестовой таблицей
	store := NewParcelStore(db)
	store.tableName = tableName
	return store
}
