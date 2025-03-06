package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// SlowQueryThreshold определяет порог для медленных запросов (в миллисекундах)
const SlowQueryThreshold = 100 * time.Millisecond

// SlowQueryLogger структура для логирования медленных запросов
type SlowQueryLogger struct {
	logFile   *os.File
	threshold time.Duration
}

// NewSlowQueryLogger создает новый логгер медленных запросов
func NewSlowQueryLogger(logDir string, threshold time.Duration) (*SlowQueryLogger, error) {
	// Создаем директорию для логов, если она не существует
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания директории для логов: %w", err)
	}

	// Создаем файл лога с текущей датой
	logFileName := fmt.Sprintf("slow_queries_%s.log", time.Now().Format("20060102"))
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла лога: %w", err)
	}

	return &SlowQueryLogger{
		logFile:   logFile,
		threshold: threshold,
	}, nil
}

// Close закрывает файл лога
func (l *SlowQueryLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// LogQuery логирует запрос, если его время выполнения превышает порог
func (l *SlowQueryLogger) LogQuery(query string, duration time.Duration, rows int64) {
	if duration >= l.threshold {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("[%s] Медленный запрос (%.2f мс): %s\nЗатронуто строк: %d\n\n",
			timestamp, float64(duration.Microseconds())/1000.0, query, rows)

		if _, err := l.logFile.WriteString(logEntry); err != nil {
			log.Printf("Ошибка записи в лог медленных запросов: %v", err)
		}

		// Дублируем в обычный лог
		log.Printf("ВНИМАНИЕ: Медленный запрос (%.2f мс): %s", float64(duration.Microseconds())/1000.0, query)
	}
}

// PerformanceStats содержит статистику выполнения запроса
type PerformanceStats struct {
	Query        string
	Duration     time.Duration
	RowsAffected int64
	Explanation  string
}

// Глобальный логгер медленных запросов
var slowQueryLogger *SlowQueryLogger

// InitPerformanceMonitoring инициализирует мониторинг производительности
func InitPerformanceMonitoring(logDir string) error {
	var err error
	slowQueryLogger, err = NewSlowQueryLogger(logDir, SlowQueryThreshold)
	if err != nil {
		return fmt.Errorf("ошибка инициализации логгера медленных запросов: %w", err)
	}

	log.Printf("Мониторинг производительности запросов инициализирован. Порог для медленных запросов: %v", SlowQueryThreshold)
	return nil
}

// ClosePerformanceMonitoring закрывает мониторинг производительности
func ClosePerformanceMonitoring() {
	if slowQueryLogger != nil {
		if err := slowQueryLogger.Close(); err != nil {
			log.Printf("Ошибка закрытия логгера медленных запросов: %v", err)
		}
	}
}

// AnalyzeQuery выполняет анализ производительности запроса
func AnalyzeQuery(db *sql.DB, query string) (*PerformanceStats, error) {
	// Измеряем время выполнения запроса
	start := time.Now()

	// Выполняем запрос
	result, err := db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	// Получаем количество затронутых строк
	rowsAffected, _ := result.RowsAffected()

	// Вычисляем длительность выполнения
	duration := time.Since(start)

	// Логируем медленный запрос, если превышен порог
	if slowQueryLogger != nil {
		slowQueryLogger.LogQuery(query, duration, rowsAffected)
	}

	// Получаем план выполнения запроса
	explainQuery := fmt.Sprintf("EXPLAIN ANALYZE %s", query)
	rows, err := db.Query(explainQuery)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения плана запроса: %w", err)
	}
	defer rows.Close()

	// Собираем результаты EXPLAIN ANALYZE
	var explanation string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, fmt.Errorf("ошибка сканирования результата EXPLAIN: %w", err)
		}
		explanation += line + "\n"
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам EXPLAIN: %w", err)
	}

	return &PerformanceStats{
		Query:        query,
		Duration:     duration,
		RowsAffected: rowsAffected,
		Explanation:  explanation,
	}, nil
}

// CheckDatabasePerformanceDetailed проверяет производительность ключевых запросов с подробным логированием
func CheckDatabasePerformanceDetailed(db *sql.DB) error {
	log.Println("Проверка производительности базы данных с подробным логированием...")

	// Список ключевых запросов для проверки
	queries := []string{
		"SELECT * FROM users WHERE email = 'test@example.com'",
		"SELECT * FROM refresh_tokens WHERE user_id = 1",
		"SELECT * FROM refresh_tokens WHERE token = 'test-token'",
		"SELECT * FROM customer WHERE email = 'customer@example.com'",
		"SELECT * FROM courier WHERE status = 'active'",
		"SELECT * FROM parcel WHERE status = 'pending'",
		"SELECT * FROM delivery WHERE courier_id = 1",
		"SELECT d.* FROM delivery d JOIN parcel p ON d.parcel_id = p.id WHERE p.status = 'delivered'",
	}

	// Проверяем каждый запрос
	for _, query := range queries {
		stats, err := AnalyzeQuery(db, query)
		if err != nil {
			log.Printf("Ошибка анализа запроса '%s': %v", query, err)
			continue
		}

		log.Printf("Запрос: %s", stats.Query)
		log.Printf("Время выполнения: %v", stats.Duration)
		log.Printf("Затронуто строк: %d", stats.RowsAffected)
		log.Printf("План выполнения:\n%s", stats.Explanation)
		log.Println("-----------------------------------")
	}

	return nil
}

// AutoAddMissingIndices автоматически добавляет недостающие индексы на основе анализа запросов
func AutoAddMissingIndices(db *sql.DB) error {
	log.Println("Анализ и автоматическое добавление недостающих индексов...")

	// Получаем список запросов, которые используют Sequential Scan на больших таблицах
	rows, err := db.Query(`
		SELECT relname, seq_scan, seq_tup_read, idx_scan, idx_tup_fetch
		FROM pg_stat_user_tables
		WHERE seq_scan > 0
		ORDER BY seq_tup_read DESC
		LIMIT 10
	`)
	if err != nil {
		return fmt.Errorf("ошибка получения статистики таблиц: %w", err)
	}
	defer rows.Close()

	// Анализируем результаты и создаем индексы для таблиц с высоким seq_scan
	for rows.Next() {
		var tableName string
		var seqScan, seqTupRead, idxScan, idxTupFetch int64

		if err := rows.Scan(&tableName, &seqScan, &seqTupRead, &idxScan, &idxTupFetch); err != nil {
			log.Printf("Ошибка сканирования результата: %v", err)
			continue
		}

		// Если количество последовательных сканирований значительно превышает индексные
		// и количество прочитанных кортежей велико, рекомендуем создать индекс
		if seqScan > idxScan*2 && seqTupRead > 1000 {
			log.Printf("Таблица %s имеет высокое количество последовательных сканирований (%d) и прочитанных кортежей (%d)",
				tableName, seqScan, seqTupRead)

			// Получаем список часто используемых полей в WHERE
			fieldsRows, err := db.Query(`
				SELECT a.attname, s.most_common_vals
				FROM pg_stats s
				JOIN pg_attribute a ON s.attname = a.attname
				WHERE s.tablename = $1
				AND s.n_distinct > 0
				AND s.correlation < 0.5
				ORDER BY s.most_common_freq DESC
				LIMIT 3
			`, tableName)

			if err != nil {
				log.Printf("Ошибка получения статистики полей для таблицы %s: %v", tableName, err)
				continue
			}

			for fieldsRows.Next() {
				var fieldName string
				var mostCommonVals interface{}

				if err := fieldsRows.Scan(&fieldName, &mostCommonVals); err != nil {
					log.Printf("Ошибка сканирования результата: %v", err)
					continue
				}

				// Проверяем, существует ли уже индекс для этого поля
				var indexExists bool
				indexName := fmt.Sprintf("idx_%s_%s", tableName, fieldName)
				err := db.QueryRow(`
					SELECT EXISTS (
						SELECT 1 
						FROM pg_indexes 
						WHERE tablename = $1 AND indexname = $2
					)
				`, tableName, indexName).Scan(&indexExists)

				if err != nil {
					log.Printf("Ошибка проверки существования индекса: %v", err)
					continue
				}

				// Если индекс не существует, создаем его
				if !indexExists {
					log.Printf("Создание индекса для поля %s в таблице %s", fieldName, tableName)
					_, err := db.Exec(fmt.Sprintf("CREATE INDEX %s ON %s(%s)", indexName, tableName, fieldName))
					if err != nil {
						log.Printf("Ошибка создания индекса: %v", err)
					} else {
						log.Printf("Индекс %s успешно создан", indexName)
					}
				}
			}
			fieldsRows.Close()
		}
	}

	return nil
}
