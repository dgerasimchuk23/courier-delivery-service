package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// TokenCleanupStats содержит статистику очистки токенов
type TokenCleanupStats struct {
	DeletedExpired    int64
	DeletedDuplicates int64
	TotalRemaining    int64
	ExecutionTime     time.Duration
}

// CleanupExpiredTokens удаляет просроченные токены из базы данных
func CleanupExpiredTokens(db *sql.DB) (*TokenCleanupStats, error) {
	log.Println("Запуск очистки просроченных токенов...")
	startTime := time.Now()

	// Удаляем просроченные токены
	result, err := db.Exec("DELETE FROM refresh_tokens WHERE expires_at < $1", time.Now())
	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении просроченных токенов: %w", err)
	}

	deletedExpired, _ := result.RowsAffected()
	log.Printf("Удалено %d просроченных токенов", deletedExpired)

	// Получаем общее количество оставшихся токенов
	var totalRemaining int64
	err = db.QueryRow("SELECT COUNT(*) FROM refresh_tokens").Scan(&totalRemaining)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении количества токенов: %w", err)
	}

	executionTime := time.Since(startTime)

	return &TokenCleanupStats{
		DeletedExpired:    deletedExpired,
		DeletedDuplicates: 0,
		TotalRemaining:    totalRemaining,
		ExecutionTime:     executionTime,
	}, nil
}

// CleanupDuplicateTokens удаляет дубликаты токенов, оставляя только последние N для каждого пользователя
func CleanupDuplicateTokens(db *sql.DB, keepLatest int) (*TokenCleanupStats, error) {
	log.Printf("Запуск очистки дубликатов токенов (оставляем последние %d)...", keepLatest)
	startTime := time.Now()

	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("ошибка при начале транзакции: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Удаляем дубликаты токенов, оставляя только последние N для каждого пользователя
	result, err := tx.Exec(`
		DELETE FROM refresh_tokens 
		WHERE id IN (
			SELECT id FROM (
				SELECT id, 
				       ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) as row_num 
				FROM refresh_tokens
			) as ranked 
			WHERE row_num > $1
		)
	`, keepLatest)
	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении дубликатов токенов: %w", err)
	}

	deletedDuplicates, _ := result.RowsAffected()
	log.Printf("Удалено %d дубликатов токенов", deletedDuplicates)

	// Получаем общее количество оставшихся токенов
	var totalRemaining int64
	err = tx.QueryRow("SELECT COUNT(*) FROM refresh_tokens").Scan(&totalRemaining)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении количества токенов: %w", err)
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("ошибка при фиксации транзакции: %w", err)
	}

	executionTime := time.Since(startTime)

	return &TokenCleanupStats{
		DeletedExpired:    0,
		DeletedDuplicates: deletedDuplicates,
		TotalRemaining:    totalRemaining,
		ExecutionTime:     executionTime,
	}, nil
}

// OptimizeTokenStorage выполняет полную оптимизацию хранения токенов
func OptimizeTokenStorage(db *sql.DB, keepLatest int) (*TokenCleanupStats, error) {
	log.Println("Запуск полной оптимизации хранения токенов...")
	startTime := time.Now()

	// Удаляем просроченные токены
	expiredStats, err := CleanupExpiredTokens(db)
	if err != nil {
		return nil, err
	}

	// Удаляем дубликаты токенов
	duplicateStats, err := CleanupDuplicateTokens(db, keepLatest)
	if err != nil {
		return nil, err
	}

	// Выполняем VACUUM ANALYZE для таблицы refresh_tokens
	_, err = db.Exec("VACUUM ANALYZE refresh_tokens")
	if err != nil {
		log.Printf("Ошибка при выполнении VACUUM ANALYZE: %v", err)
	}

	executionTime := time.Since(startTime)

	return &TokenCleanupStats{
		DeletedExpired:    expiredStats.DeletedExpired,
		DeletedDuplicates: duplicateStats.DeletedDuplicates,
		TotalRemaining:    duplicateStats.TotalRemaining,
		ExecutionTime:     executionTime,
	}, nil
}

// ScheduleTokenCleanup запускает периодическую очистку токенов
func ScheduleTokenCleanup(db *sql.DB, interval time.Duration, keepLatest int) chan bool {
	done := make(chan bool)
	ticker := time.NewTicker(interval)

	go func() {
		// Сразу выполняем очистку при запуске
		_, err := OptimizeTokenStorage(db, keepLatest)
		if err != nil {
			log.Printf("Ошибка при оптимизации хранения токенов: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				_, err := OptimizeTokenStorage(db, keepLatest)
				if err != nil {
					log.Printf("Ошибка при оптимизации хранения токенов: %v", err)
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return done
}
