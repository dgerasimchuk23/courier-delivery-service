package db

import (
	"database/sql"
	"delivery/config"
	migrations "delivery/internal/db/migrations"
	"fmt"
	"log"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

var tokenCleanupDone chan bool

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
	maxRetries := 5
	retryInterval := 3 * time.Second

	// Если DB не существует, создать
	if err := createPostgresDBIfNotExists(config); err != nil {
		log.Fatalf("Ошибка при создании базы данных PostgreSQL: %v", err)
	}

	// Подключение к PostgreSQL с ретраями
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host, config.Database.Port, config.Database.User,
		config.Database.Password, config.Database.DBName, config.Database.SSLMode)

	// Пытаемся подключиться с ретраями
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Printf("Попытка %d: Ошибка подключения к PostgreSQL: %v", i+1, err)
			time.Sleep(retryInterval)
			continue
		}

		// Проверка соединения
		if err := db.Ping(); err != nil {
			log.Printf("Попытка %d: Ошибка проверки соединения с базой данных: %v", i+1, err)
			db.Close()
			time.Sleep(retryInterval)
			continue
		}

		// Если дошли сюда, значит подключение успешно
		log.Printf("Успешное подключение к базе данных после %d попыток", i+1)
		break
	}

	// Если после всех попыток не удалось подключиться
	if db == nil || err != nil {
		log.Fatalf("Не удалось подключиться к базе данных после %d попыток: %v", maxRetries, err)
	}

	// Инициализация схемы БД
	if err := migrations.InitSchema(db, "postgres"); err != nil {
		log.Fatalf("Ошибка инициализации схемы: %v", err)
	}

	// Выполнение миграций
	if err := migrations.MigrateDB(db); err != nil {
		log.Fatalf("Ошибка выполнения миграций: %v", err)
	}

	// Оптимизация базы данных
	if err := OptimizeDatabase(db); err != nil {
		log.Printf("Ошибка оптимизации базы данных: %v", err)
	}

	// Запускаем периодическую очистку токенов
	// Используем значение из конфигурации вместо жестко заданного 5
	maxRefreshTokens := config.Database.MaxRefreshTokens
	log.Printf("Настройка очистки токенов: хранение последних %d токенов для каждого пользователя", maxRefreshTokens)
	tokenCleanupDone = ScheduleTokenCleanup(db, 12*time.Hour, maxRefreshTokens)

	// Инициализация мониторинга производительности
	logDir := filepath.Join("logs", "db_performance")
	if err := InitPerformanceMonitoring(logDir); err != nil {
		log.Printf("Ошибка инициализации мониторинга производительности: %v", err)
	}

	// Проверка производительности ключевых запросов
	if config.Database.CheckPerformance {
		if err := CheckDatabasePerformance(db); err != nil {
			log.Printf("Ошибка проверки производительности базы данных: %v", err)
		}

		// Дополнительно запускаем подробную проверку с логированием
		if err := CheckDatabasePerformanceDetailed(db); err != nil {
			log.Printf("Ошибка подробной проверки производительности базы данных: %v", err)
		}

		// Автоматическое добавление недостающих индексов
		if err := AutoAddMissingIndices(db); err != nil {
			log.Printf("Ошибка при автоматическом добавлении индексов: %v", err)
		}
	}

	return db
}

// CloseDB закрывает соединение с базой данных и останавливает фоновые задачи
func CloseDB(db *sql.DB) {
	// Останавливаем периодическую очистку токенов
	if tokenCleanupDone != nil {
		tokenCleanupDone <- true
	}

	// Закрываем мониторинг производительности
	ClosePerformanceMonitoring()

	// Закрываем соединение с базой данных
	if db != nil {
		db.Close()
	}
}
