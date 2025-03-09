package parcel

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func setupParcelTestDB() *ParcelStore {
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

// Вспомогательная функция для получения переменных окружения
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
