package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

// PerformanceStats содержит статистику выполнения запроса
type PerformanceStats struct {
	Query        string
	Duration     time.Duration
	RowsAffected int64
	Explanation  string
}

// LoadQueries загружает SQL-запросы из файла
func LoadQueries(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %w", filename, err)
	}
	queries := strings.Split(string(data), ";") // Разделяем по `;`
	return queries, nil
}

// AnalyzeQuery выполняет анализ производительности запроса
func AnalyzeQuery(db *sql.DB, query string) (*PerformanceStats, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil // Пропускаем пустые запросы
	}

	start := time.Now()
	result, err := db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	duration := time.Since(start)

	// Получаем план выполнения запроса
	explainQuery := fmt.Sprintf("EXPLAIN ANALYZE %s", query)
	rows, err := db.Query(explainQuery)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения плана запроса: %w", err)
	}
	defer rows.Close()

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

// CheckDatabasePerformance проверяет производительность запросов из файла
func CheckDatabasePerformance(db *sql.DB, queriesFile string) error {
	log.Println("Проверка производительности базы данных...")

	queries, err := LoadQueries(queriesFile)
	if err != nil {
		return err
	}

	for _, query := range queries {
		stats, err := AnalyzeQuery(db, query)
		if err != nil {
			log.Printf("Ошибка анализа запроса '%s': %v", query, err)
			continue
		}
		if stats == nil {
			continue
		}

		log.Printf("Запрос: %s", stats.Query)
		log.Printf("Время выполнения: %v", stats.Duration)
		log.Printf("Затронуто строк: %d", stats.RowsAffected)
		log.Printf("План выполнения:\n%s", stats.Explanation)
		log.Println("-----------------------------------")

		if stats.Duration > 100*time.Millisecond {
			log.Printf("ВНИМАНИЕ: Запрос выполняется слишком долго: %v", stats.Duration)
		}
	}

	return nil
}

// OptimizeDatabase выполняет оптимизацию базы данных
func OptimizeDatabase(db *sql.DB) error {
	log.Println("Выполнение оптимизации базы данных...")

	_, err := db.Exec("VACUUM ANALYZE")
	if err != nil {
		return fmt.Errorf("ошибка выполнения VACUUM: %w", err)
	}

	log.Println("Оптимизация базы данных завершена")
	return nil
}
