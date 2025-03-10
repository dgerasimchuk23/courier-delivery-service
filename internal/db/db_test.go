package db

import (
	"delivery/internal/config"
	"os"
	"strconv"
	"testing"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return intValue
}

func TestInitDB(t *testing.T) {
	// Тестовая конфигурация
	cfg := &config.Config{}
	cfg.Database.Type = "postgres"
	cfg.Database.Host = getEnv("DB_HOST", "db")
	cfg.Database.Port = getEnvInt("DB_PORT", 5432)
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.DBName = getEnv("DB_NAME", "delivery")
	cfg.Database.SSLMode = "disable"

	// Инициализация DB
	db := InitDB(cfg)
	if db == nil {
		t.Fatal("Не удалось инициализировать базу данных")
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Закрываем соединение
	if err := db.Close(); err != nil {
		t.Fatalf("Не удалось закрыть соединение с базой данных: %v", err)
	}
}

func TestCreatePostgresDBIfNotExists(t *testing.T) {
	cfg := &config.Config{}
	cfg.Database.Type = "postgres"
	cfg.Database.Host = getEnv("DB_HOST", "db")
	cfg.Database.Port = getEnvInt("DB_PORT", 5432)
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.DBName = getEnv("DB_NAME", "delivery")
	cfg.Database.SSLMode = "disable"

	// Проверка создания DB
	err := createPostgresDBIfNotExists(cfg)
	if err != nil {
		t.Fatalf("Не удалось создать базу данных: %v", err)
	}

	// Подключение к созданной DB
	db := InitDB(cfg)
	if db == nil {
		t.Fatal("Не удалось инициализировать базу данных")
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Закрываем соединение
	if err := db.Close(); err != nil {
		t.Fatalf("Не удалось закрыть соединение с базой данных: %v", err)
	}
}
