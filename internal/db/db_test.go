package db

import (
	"delivery/config"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Тестовая конфигурация
	cfg := &config.Config{}
	cfg.Database.Type = "postgres"
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "postgres"
	cfg.Database.Password = "postgres"
	cfg.Database.DBName = "delivery_test"
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
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "postgres"
	cfg.Database.Password = "postgres"
	cfg.Database.DBName = "delivery_test"
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
