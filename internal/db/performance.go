package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// PerformanceStats содержит статистику выполнения запроса
type PerformanceStats struct {
	Query        string
	Duration     time.Duration
	RowsAffected int64
	Explanation  string
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

// CheckDatabasePerformance проверяет производительность ключевых запросов
func CheckDatabasePerformance(db *sql.DB) error {
	log.Println("Проверка производительности базы данных...")

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

		// Если запрос выполняется слишком долго, выводим предупреждение
		if stats.Duration > 100*time.Millisecond {
			log.Printf("ВНИМАНИЕ: Запрос выполняется слишком долго: %v", stats.Duration)
		}
	}

	return nil
}

// OptimizeDatabase выполняет оптимизацию базы данных
func OptimizeDatabase(db *sql.DB) error {
	log.Println("Выполнение оптимизации базы данных...")

	// Выполняем VACUUM для очистки и анализа таблиц
	_, err := db.Exec("VACUUM ANALYZE")
	if err != nil {
		return fmt.Errorf("ошибка выполнения VACUUM: %w", err)
	}

	// Анализируем таблицы для обновления статистики
	tables := []string{"users", "refresh_tokens", "customer", "courier", "parcel", "delivery"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("ANALYZE %s", table))
		if err != nil {
			log.Printf("Ошибка выполнения ANALYZE для таблицы %s: %v", table, err)
		}
	}

	log.Println("Оптимизация базы данных завершена")
	return nil
}
