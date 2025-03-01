package db

import (
	"database/sql"
	"delivery/config"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Создание базы данных PostgreSQL, если она не существует
func createPostgresDBIfNotExists(config *config.Config) error {
	// Подключаемся к postgres (системная база данных)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Database.Host, config.Database.Port, config.Database.User,
		config.Database.Password, config.Database.SSLMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	// Проверка существования базы данных
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", config.Database.DBName)
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования базы данных: %v", err)
	}

	// Если DB не существует, создать
	if !exists {
		log.Printf("База данных '%s' не существует. Создание...", config.Database.DBName)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.Database.DBName))
		if err != nil {
			return fmt.Errorf("ошибка создания базы данных: %v", err)
		}
		log.Printf("База данных '%s' успешно создана", config.Database.DBName)
	}

	return nil
}

func InitDB(config *config.Config) *sql.DB {
	var db *sql.DB
	var err error

	// Если DB не существует, создать
	if err := createPostgresDBIfNotExists(config); err != nil {
		log.Fatalf("Ошибка при создании базы данных PostgreSQL: %v", err)
	}

	// Подключение к PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host, config.Database.Port, config.Database.User,
		config.Database.Password, config.Database.DBName, config.Database.SSLMode)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка проверки соединения с базой данных: %v", err)
	}

	// Инициализация схемы БД
	if err := InitSchema(db, "postgres"); err != nil {
		log.Fatalf("Ошибка инициализации схемы: %v", err)
	}

	return db
}
