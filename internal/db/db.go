package db

import (
	"context"
	"database/sql"
	"delivery/internal/config"
	migrations "delivery/internal/db/migrations"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var tokenCleanupDone chan bool
var queryLogMutex sync.Mutex
var queryLogFile *os.File
var performanceAnalysisDone chan bool

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

// LogQuery логирует SQL-запрос в файл и консоль
func LogQuery(query string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] SQL-запрос: %s\nПараметры: %v\n\n", timestamp, query, args)

	// Логируем в файл
	queryLogMutex.Lock()
	defer queryLogMutex.Unlock()

	if queryLogFile != nil {
		if _, err := queryLogFile.WriteString(logEntry); err != nil {
			log.Printf("Ошибка записи SQL-запроса в лог: %v", err)
		}
	}

	// Логируем в консоль в режиме отладки
	if os.Getenv("DEBUG_SQL") == "true" {
		log.Printf("SQL: %s, args: %v", query, args)
	}
}

// InitQueryLogging инициализирует логирование SQL-запросов
func InitQueryLogging(logDir string) error {
	// Создаем директорию для логов, если она не существует
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории для логов запросов: %w", err)
	}

	// Создаем файл лога с текущей датой
	logFileName := fmt.Sprintf("sql_queries_%s.log", time.Now().Format("20060102"))
	logFilePath := filepath.Join(logDir, logFileName)

	var err error
	queryLogFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла лога запросов: %w", err)
	}

	log.Printf("Логирование SQL-запросов инициализировано в файл: %s", logFilePath)
	return nil
}

// CloseQueryLogging закрывает файл лога запросов
func CloseQueryLogging() {
	queryLogMutex.Lock()
	defer queryLogMutex.Unlock()

	if queryLogFile != nil {
		if err := queryLogFile.Close(); err != nil {
			log.Printf("Ошибка закрытия файла лога запросов: %v", err)
		}
		queryLogFile = nil
	}
}

// DB представляет обертку над sql.DB с логированием запросов
type DB struct {
	*sql.DB
}

// NewDB создает новую обертку над sql.DB
func NewDB(db *sql.DB) *DB {
	return &DB{DB: db}
}

// Query выполняет запрос с логированием
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	LogQuery(query, args...)
	return db.DB.Query(query, args...)
}

// QueryRow выполняет запрос, возвращающий одну строку, с логированием
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	LogQuery(query, args...)
	return db.DB.QueryRow(query, args...)
}

// Exec выполняет запрос без возврата строк с логированием
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	LogQuery(query, args...)
	return db.DB.Exec(query, args...)
}

// QueryContext выполняет запрос с контекстом и логированием
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	LogQuery(query, args...)
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRowContext выполняет запрос, возвращающий одну строку, с контекстом и логированием
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	LogQuery(query, args...)
	return db.DB.QueryRowContext(ctx, query, args...)
}

// ExecContext выполняет запрос без возврата строк с контекстом и логированием
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	LogQuery(query, args...)
	return db.DB.ExecContext(ctx, query, args...)
}

// CleanupTokens выполняет очистку устаревших токенов
func CleanupTokens(db *sql.DB, maxTokensPerUser int, done chan bool) {
	log.Printf("Запущена задача очистки токенов. Максимальное количество токенов на пользователя: %d", maxTokensPerUser)

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	// Функция очистки токенов
	cleanup := func() {
		// Устанавливаем таймаут на выполнение операции
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		startTime := time.Now()
		log.Println("Начало очистки токенов...")

		// Удаляем просроченные токены
		expiredResult, err := db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE expires_at < NOW()")
		if err != nil {
			log.Printf("Ошибка при удалении просроченных токенов: %v", err)
		} else {
			expiredCount, _ := expiredResult.RowsAffected()
			log.Printf("Удалено %d просроченных токенов", expiredCount)
		}

		// Удаляем лишние токены, оставляя только последние N для каждого пользователя
		excessResult, err := db.ExecContext(ctx, `
			DELETE FROM refresh_tokens 
			WHERE id IN (
				SELECT id FROM (
					SELECT id, 
						ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) as row_num 
					FROM refresh_tokens
				) as ranked 
				WHERE row_num > $1
			)`, maxTokensPerUser)

		if err != nil {
			log.Printf("Ошибка при удалении лишних токенов: %v", err)
		} else {
			excessCount, _ := excessResult.RowsAffected()
			log.Printf("Удалено %d лишних токенов", excessCount)
		}

		duration := time.Since(startTime)
		log.Printf("Очистка токенов завершена за %v", duration)
	}

	// Выполняем очистку сразу при запуске
	cleanup()

	// Затем выполняем по расписанию
	for {
		select {
		case <-ticker.C:
			cleanup()
		case <-done:
			log.Println("Задача очистки токенов остановлена")
			return
		}
	}
}

func InitDB(config *config.Config) *DB {
	var sqlDB *sql.DB
	var err error
	maxRetries := 5
	retryInterval := 3 * time.Second

	// Если DB не существует, создать
	if err := createPostgresDBIfNotExists(config); err != nil {
		log.Fatalf("Ошибка при создании базы данных PostgreSQL: %v", err)
	}

	// Инициализируем логирование запросов
	if err := InitQueryLogging("logs/sql_queries"); err != nil {
		log.Printf("Предупреждение: не удалось инициализировать логирование запросов: %v", err)
	}

	// Инициализируем мониторинг производительности
	if err := InitPerformanceMonitoring("logs/db_performance"); err != nil {
		log.Printf("Предупреждение: не удалось инициализировать мониторинг производительности: %v", err)
	}

	// Подключение к PostgreSQL с ретраями
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host, config.Database.Port, config.Database.User,
		config.Database.Password, config.Database.DBName, config.Database.SSLMode)

	// Пытаемся подключиться с ретраями
	for i := 0; i < maxRetries; i++ {
		sqlDB, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Printf("Попытка %d: Ошибка подключения к PostgreSQL: %v", i+1, err)
			time.Sleep(retryInterval)
			continue
		}

		// Проверка соединения
		if err := sqlDB.Ping(); err != nil {
			log.Printf("Попытка %d: Ошибка проверки соединения с базой данных: %v", i+1, err)
			sqlDB.Close()
			time.Sleep(retryInterval)
			continue
		}

		// Если дошли сюда, значит подключение успешно
		log.Printf("Успешное подключение к базе данных после %d попыток", i+1)
		break
	}

	// Если после всех попыток не удалось подключиться
	if sqlDB == nil || err != nil {
		log.Fatalf("Не удалось подключиться к базе данных после %d попыток: %v", maxRetries, err)
	}

	// Создаем обертку для логирования запросов
	db := NewDB(sqlDB)

	// Инициализация схемы БД
	if err := migrations.InitSchema(sqlDB, "postgres"); err != nil {
		log.Fatalf("Ошибка инициализации схемы: %v", err)
	}

	// Выполнение миграций
	if err := migrations.MigrateDB(sqlDB); err != nil {
		log.Fatalf("Ошибка выполнения миграций: %v", err)
	}

	// Проверка производительности базы данных
	if err := CheckDatabasePerformanceDetailed(sqlDB); err != nil {
		log.Printf("Предупреждение: ошибка проверки производительности базы данных: %v", err)
	}

	// Автоматическое добавление недостающих индексов
	if err := AutoAddMissingIndices(sqlDB); err != nil {
		log.Printf("Предупреждение: ошибка автоматического добавления индексов: %v", err)
	}

	// Запуск очистки токенов в фоне
	tokenCleanupDone = make(chan bool)
	maxTokens := 5 // Значение по умолчанию
	if config.Database.MaxRefreshTokens > 0 {
		maxTokens = config.Database.MaxRefreshTokens
	}
	go CleanupTokens(sqlDB, maxTokens, tokenCleanupDone)

	// Запуск периодического анализа производительности
	performanceAnalysisDone = SchedulePerformanceAnalysis(sqlDB, 24*time.Hour)
	log.Println("Запущен периодический анализ производительности (интервал: 24 часа)")

	return db
}

func CloseDB(db *DB) {
	if db != nil {
		// Останавливаем очистку токенов
		if tokenCleanupDone != nil {
			tokenCleanupDone <- true
			close(tokenCleanupDone)
		}

		// Останавливаем анализ производительности
		if performanceAnalysisDone != nil {
			performanceAnalysisDone <- true
			close(performanceAnalysisDone)
		}

		// Закрываем мониторинг производительности
		ClosePerformanceMonitoring()

		// Закрываем логирование запросов
		CloseQueryLogging()

		// Закрываем соединение с БД
		if err := db.Close(); err != nil {
			log.Printf("Ошибка закрытия соединения с базой данных: %v", err)
		}
	}
}
