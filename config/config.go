package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Структура для хранения конфигурации
type Config struct {
	Database struct {
		Type     string `json:"type"`     // "sqlite" или "postgres"
		Host     string `json:"host"`     // для PostgreSQL
		Port     int    `json:"port"`     // для PostgreSQL
		User     string `json:"user"`     // для PostgreSQL
		Password string `json:"password"` // для PostgreSQL
		DBName   string `json:"dbname"`   // для PostgreSQL
		SSLMode  string `json:"sslmode"`  // для PostgreSQL
		DSN      string `json:"dsn"`      // для SQLite - путь к файлу БД
	} `json:"database"`
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
}

// Читает файл конфигурации и возвращает структуру Config
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла конфигурации: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("ошибка декодирования файла конфигурации: %v", err)
	}

	return &config, nil
}
