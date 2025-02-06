package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func InitDB(dsn string) *sql.DB {
	// Получить директорию из пути
	dir := "internal/db"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Ошибка создания директории для базы данных: %v", err)
	}

	// Проверка существования файла базы данных
	if _, err := os.Stat(dsn); os.IsNotExist(err) {
		log.Printf("Файл базы данных не найден. Будет создан: %s", dsn)
		file, err := os.Create(dsn)
		if err != nil {
			log.Fatalf("Ошибка создания файла базы данных: %v", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	// Инициализация схемы БД
	if err := InitSchema(db); err != nil {
		log.Fatalf("Ошибка инициализации схемы: %v", err)
	}

	return db
}
