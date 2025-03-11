package db

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	// Пропускаем запросы, которые не являются SELECT, INSERT, UPDATE или DELETE
	query = strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(query), "SELECT") &&
		!strings.HasPrefix(strings.ToUpper(query), "INSERT") &&
		!strings.HasPrefix(strings.ToUpper(query), "UPDATE") &&
		!strings.HasPrefix(strings.ToUpper(query), "DELETE") {
		return nil, fmt.Errorf("запрос не является SELECT, INSERT, UPDATE или DELETE")
	}

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

	// Получаем план выполнения запроса для SELECT
	var explanation string
	if strings.HasPrefix(strings.ToUpper(query), "SELECT") {
		explainQuery := fmt.Sprintf("EXPLAIN ANALYZE %s", query)
		rows, err := db.Query(explainQuery)
		if err != nil {
			log.Printf("Ошибка получения плана запроса: %v", err)
		} else {
			defer rows.Close()

			// Собираем результаты EXPLAIN ANALYZE
			for rows.Next() {
				var line string
				if err := rows.Scan(&line); err != nil {
					log.Printf("Ошибка сканирования результата EXPLAIN: %v", err)
					continue
				}
				explanation += line + "\n"
			}

			if err := rows.Err(); err != nil {
				log.Printf("Ошибка при итерации по результатам EXPLAIN: %v", err)
			}
		}
	}

	return &PerformanceStats{
		Query:        query,
		Duration:     duration,
		RowsAffected: rowsAffected,
		Explanation:  explanation,
	}, nil
}

// ReadQueriesFromFile читает SQL-запросы из файла
func ReadQueriesFromFile(filePath string) ([]string, error) {
	// Получаем абсолютный путь к файлу
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	// Пытаемся прочитать файл с абсолютным путем
	log.Printf("Пытаемся прочитать файл: %s (абсолютный путь: %s)", filePath, absPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		// Если не удалось, пробуем с относительным путем
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения файла: %w", err)
		}
	}

	// Разделяем содержимое файла на отдельные запросы
	fileContent := string(content)
	queries := []string{}

	// Регулярное выражение для поиска SQL-запросов
	// Ищем запросы, которые начинаются с SELECT, INSERT, UPDATE, DELETE
	// и заканчиваются точкой с запятой
	re := regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE)[\s\S]*?;`)
	matches := re.FindAllString(fileContent, -1)

	for _, match := range matches {
		// Пропускаем комментарии и пустые строки
		if !strings.HasPrefix(strings.TrimSpace(match), "--") && strings.TrimSpace(match) != "" {
			queries = append(queries, strings.TrimSpace(match))
		}
	}

	return queries, nil
}

// ExtractQueriesFromLogs извлекает SQL-запросы из лог-файлов
func ExtractQueriesFromLogs(logDir string) ([]string, error) {
	// Получаем список файлов логов
	files, err := filepath.Glob(filepath.Join(logDir, "sql_queries_*.log"))
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска файлов логов: %w", err)
	}

	uniqueQueries := make(map[string]struct{})
	queryRegex := regexp.MustCompile(`SQL-запрос: (.+)`)

	// Обрабатываем каждый файл логов
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Printf("Ошибка открытия файла лога %s: %v", file, err)
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			matches := queryRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				query := strings.TrimSpace(matches[1])
				// Проверяем, что это SQL-запрос
				if strings.HasPrefix(strings.ToUpper(query), "SELECT") ||
					strings.HasPrefix(strings.ToUpper(query), "INSERT") ||
					strings.HasPrefix(strings.ToUpper(query), "UPDATE") ||
					strings.HasPrefix(strings.ToUpper(query), "DELETE") {
					// Добавляем в уникальные запросы
					uniqueQueries[query] = struct{}{}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Ошибка чтения файла лога %s: %v", file, err)
		}
	}

	// Преобразуем уникальные запросы в список
	queries := make([]string, 0, len(uniqueQueries))
	for query := range uniqueQueries {
		queries = append(queries, query)
	}

	return queries, nil
}

// CheckDatabasePerformanceDetailed проверяет производительность ключевых запросов с подробным логированием
func CheckDatabasePerformanceDetailed(db *sql.DB) error {
	log.Println("Проверка производительности базы данных с подробным логированием...")

	// Список запросов для проверки
	var queries []string

	// Читаем запросы из файла analyze_queries.sql
	// Используем только один правильный путь к файлу
	path := "/root/internal/db/analyze_queries.sql"
	fileQueries, fileReadErr := ReadQueriesFromFile(path)
	if fileReadErr == nil {
		log.Printf("Прочитано %d запросов из файла %s", len(fileQueries), path)
		queries = append(queries, fileQueries...)
	} else {
		log.Printf("Не удалось прочитать запросы из файла %s: %v", path, fileReadErr)
	}

	// 2. Пытаемся извлечь запросы из логов
	logQueries, err := ExtractQueriesFromLogs("logs/sql_queries")
	if err != nil {
		log.Printf("Не удалось извлечь запросы из логов: %v", err)
	} else {
		log.Printf("Извлечено %d уникальных запросов из логов", len(logQueries))
		queries = append(queries, logQueries...)
	}

	// 3. Если не удалось получить запросы из файла и логов, используем предопределенные запросы
	if len(queries) == 0 {
		log.Println("Используем предопределенные запросы для проверки производительности")
		queries = []string{
			"SELECT * FROM users WHERE email = 'test@example.com'",
			"SELECT * FROM refresh_tokens WHERE user_id = 1",
			"SELECT * FROM refresh_tokens WHERE token = 'test-token'",
			"SELECT * FROM customer WHERE email = 'customer@example.com'",
			"SELECT * FROM courier WHERE status = 'active'",
			"SELECT * FROM parcel WHERE status = 'pending'",
			"SELECT * FROM delivery WHERE courier_id = 1",
			"SELECT d.* FROM delivery d JOIN parcel p ON d.parcel_id = p.id WHERE p.status = 'delivered'",
		}
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
		if stats.Explanation != "" {
			log.Printf("План выполнения:\n%s", stats.Explanation)
		}
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
						WHERE tablename = $1 
						AND indexname = $2
					)
				`, tableName, indexName).Scan(&indexExists)

				if err != nil {
					log.Printf("Ошибка проверки существования индекса: %v", err)
					continue
				}

				// Если индекс не существует, создаем его
				if !indexExists {
					log.Printf("Создание индекса %s для поля %s в таблице %s", indexName, fieldName, tableName)
					_, err := db.Exec(fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, fieldName))
					if err != nil {
						log.Printf("Ошибка создания индекса: %v", err)
					} else {
						log.Printf("Индекс %s успешно создан", indexName)
					}
				} else {
					log.Printf("Индекс %s уже существует", indexName)
				}
			}
			fieldsRows.Close()
		}
	}

	return nil
}

// AnalyzeLoggedQueries анализирует запросы из логов и выявляет проблемные
func AnalyzeLoggedQueries(db *sql.DB) error {
	log.Println("Анализ запросов из логов...")

	// Извлекаем запросы из логов
	queries, err := ExtractQueriesFromLogs("logs/sql_queries")
	if err != nil {
		return fmt.Errorf("ошибка извлечения запросов из логов: %w", err)
	}

	log.Printf("Найдено %d уникальных запросов в логах", len(queries))

	// Анализируем каждый запрос
	slowQueries := 0
	for _, query := range queries {
		stats, err := AnalyzeQuery(db, query)
		if err != nil {
			log.Printf("Пропуск запроса '%s': %v", query, err)
			continue
		}

		// Если запрос выполняется медленно, логируем его
		if stats.Duration >= SlowQueryThreshold {
			slowQueries++
			log.Printf("МЕДЛЕННЫЙ ЗАПРОС: %s", stats.Query)
			log.Printf("Время выполнения: %v", stats.Duration)
			log.Printf("Затронуто строк: %d", stats.RowsAffected)
			if stats.Explanation != "" {
				log.Printf("План выполнения:\n%s", stats.Explanation)
			}
			log.Println("-----------------------------------")
		}
	}

	log.Printf("Анализ завершен. Найдено %d медленных запросов из %d проанализированных", slowQueries, len(queries))
	return nil
}

// SchedulePerformanceAnalysis запускает периодический анализ производительности
func SchedulePerformanceAnalysis(db *sql.DB, interval time.Duration) chan bool {
	done := make(chan bool)
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := AnalyzeLoggedQueries(db); err != nil {
					log.Printf("Ошибка анализа запросов: %v", err)
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return done
}
