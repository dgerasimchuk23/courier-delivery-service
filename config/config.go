package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Структура для хранения конфигурации
type Config struct {
	Database struct {
		DSN string `json:"dsn"`
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
