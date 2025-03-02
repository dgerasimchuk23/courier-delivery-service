package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Database struct {
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
	} `json:"database"`
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
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

	// Проверяем, запущено ли приложение в контейнере
	inContainer := os.Getenv("IN_CONTAINER") == "true"

	// Если не в контейнере, используем localhost для всех сервисов
	if !inContainer {
		// Для базы данных
		config.Database.Host = "localhost"

		// Для Redis
		config.Redis.Host = "localhost"

		fmt.Println("Приложение запущено локально, используем localhost для всех сервисов")
	} else {
		// В контейнере используем имена сервисов из docker-compose
		config.Database.Host = "db"
		config.Redis.Host = "redis"

		fmt.Println("Приложение запущено в контейнере, используем имена сервисов")
	}

	// Переопределение значений из переменных окружения, если они заданы
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		config.Database.DBName = dbname
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		config.Database.SSLMode = sslmode
	}

	return &config, nil
}
