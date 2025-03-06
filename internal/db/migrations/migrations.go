package db

import (
	"database/sql"
	"log"
)

// MigrationResult содержит результаты выполнения миграции
type MigrationResult struct {
	TablesCreated    int
	IndicesCreated   int
	RowsAffected     int64
	ExecutionSuccess bool
}

// MigrateDB выполняет миграции базы данных
func MigrateDB(db *sql.DB) error {
	log.Println("Запуск миграций базы данных...")
	var result MigrationResult

	// Шаг 1: Создание схемы базы данных (таблиц)
	if err := InitSchema(db, "postgres"); err != nil {
		log.Printf("Ошибка при создании схемы базы данных: %v", err)
		return err
	}
	result.TablesCreated = 6 // Количество созданных таблиц (users, refresh_tokens, customer, courier, parcel, delivery)
	log.Printf("Создано или обновлено %d таблиц", result.TablesCreated)

	// Шаг 2: Создание индексов
	indexResult, err := InitIndexes(db)
	if err != nil {
		log.Printf("Ошибка при создании индексов: %v", err)
		return err
	}
	result.IndicesCreated = indexResult.IndicesCreated
	result.RowsAffected = indexResult.RowsAffected
	log.Printf("Создано %d индексов, затронуто %d строк", result.IndicesCreated, result.RowsAffected)

	result.ExecutionSuccess = true
	log.Println("Миграции успешно выполнены")
	return nil
}
