package db

import (
	"context"
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
	TimedOut          bool
}

// CleanupExpiredTokens удаляет просроченные токены из базы данных
func CleanupExpiredTokens(db *sql.DB) (*TokenCleanupStats, error) {
	return CleanupExpiredTokensWithTimeout(db, 30*time.Second)
}

// CleanupExpiredTokensWithTimeout удаляет просроченные токены из базы данных с ограничением по времени
func CleanupExpiredTokensWithTimeout(db *sql.DB, timeout time.Duration) (*TokenCleanupStats, error) {
	log.Println("Запуск очистки просроченных токенов...")
	startTime := time.Now()

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Удаляем просроченные токены с использованием контекста
	result, err := db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE expires_at < $1", time.Now())

	// Проверяем, не истек ли таймаут
	timedOut := ctx.Err() == context.DeadlineExceeded
	if timedOut {
		log.Printf("Превышено максимальное время выполнения очистки токенов (%s)", timeout)
		return &TokenCleanupStats{
			ExecutionTime: time.Since(startTime),
			TimedOut:      true,
		}, fmt.Errorf("превышено максимальное время выполнения: %w", ctx.Err())
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении просроченных токенов: %w", err)
	}

	deletedExpired, _ := result.RowsAffected()
	log.Printf("Удалено %d просроченных токенов", deletedExpired)

	// Получаем общее количество оставшихся токенов
	var totalRemaining int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM refresh_tokens").Scan(&totalRemaining)

	// Снова проверяем таймаут
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Превышено максимальное время выполнения при подсчете оставшихся токенов (%s)", timeout)
		return &TokenCleanupStats{
			DeletedExpired: deletedExpired,
			ExecutionTime:  time.Since(startTime),
			TimedOut:       true,
		}, fmt.Errorf("превышено максимальное время выполнения: %w", ctx.Err())
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении количества токенов: %w", err)
	}

	executionTime := time.Since(startTime)

	return &TokenCleanupStats{
		DeletedExpired:    deletedExpired,
		DeletedDuplicates: 0,
		TotalRemaining:    totalRemaining,
		ExecutionTime:     executionTime,
		TimedOut:          false,
	}, nil
}

// CleanupDuplicateTokens удаляет дубликаты токенов, оставляя только последние N для каждого пользователя
func CleanupDuplicateTokens(db *sql.DB, keepLatest int) (*TokenCleanupStats, error) {
	return CleanupDuplicateTokensWithTimeout(db, keepLatest, 30*time.Second)
}

// CleanupDuplicateTokensWithTimeout удаляет дубликаты токенов с ограничением по времени
func CleanupDuplicateTokensWithTimeout(db *sql.DB, keepLatest int, timeout time.Duration) (*TokenCleanupStats, error) {
	log.Printf("Запуск очистки дубликатов токенов (оставляем последние %d)...", keepLatest)
	startTime := time.Now()

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Начинаем транзакцию
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при начале транзакции: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Ошибка при откате транзакции: %v", rollbackErr)
			}
		}
	}()

	// Проверяем, не истек ли таймаут
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Превышено максимальное время выполнения при начале транзакции (%s)", timeout)
		return &TokenCleanupStats{
			ExecutionTime: time.Since(startTime),
			TimedOut:      true,
		}, fmt.Errorf("превышено максимальное время выполнения: %w", ctx.Err())
	}

	// Удаляем дубликаты токенов, оставляя только последние N для каждого пользователя
	result, err := tx.ExecContext(ctx, `
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

	// Проверяем таймаут после выполнения запроса
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Превышено максимальное время выполнения при удалении дубликатов (%s)", timeout)
		return &TokenCleanupStats{
			ExecutionTime: time.Since(startTime),
			TimedOut:      true,
		}, fmt.Errorf("превышено максимальное время выполнения: %w", ctx.Err())
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка при удалении дубликатов токенов: %w", err)
	}

	deletedDuplicates, _ := result.RowsAffected()
	log.Printf("Удалено %d дубликатов токенов", deletedDuplicates)

	// Получаем общее количество оставшихся токенов
	var totalRemaining int64
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM refresh_tokens").Scan(&totalRemaining)

	// Проверяем таймаут
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Превышено максимальное время выполнения при подсчете оставшихся токенов (%s)", timeout)
		return &TokenCleanupStats{
			DeletedDuplicates: deletedDuplicates,
			ExecutionTime:     time.Since(startTime),
			TimedOut:          true,
		}, fmt.Errorf("превышено максимальное время выполнения: %w", ctx.Err())
	}

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
		TimedOut:          false,
	}, nil
}

// OptimizeTokenStorage выполняет полную оптимизацию хранения токенов
func OptimizeTokenStorage(db *sql.DB, keepLatest int) (*TokenCleanupStats, error) {
	return OptimizeTokenStorageWithTimeout(db, keepLatest, 60*time.Second)
}

// OptimizeTokenStorageWithTimeout выполняет полную оптимизацию хранения токенов с ограничением по времени
func OptimizeTokenStorageWithTimeout(db *sql.DB, keepLatest int, timeout time.Duration) (*TokenCleanupStats, error) {
	log.Println("Запуск полной оптимизации хранения токенов...")
	startTime := time.Now()

	// Распределяем таймаут между операциями
	expiredTimeout := timeout / 3
	duplicateTimeout := timeout / 3
	vacuumTimeout := timeout / 3

	// Удаляем просроченные токены
	expiredStats, err := CleanupExpiredTokensWithTimeout(db, expiredTimeout)
	if err != nil {
		if expiredStats != nil && expiredStats.TimedOut {
			log.Printf("Очистка просроченных токенов прервана по таймауту (%s)", expiredTimeout)
			return &TokenCleanupStats{
				DeletedExpired: expiredStats.DeletedExpired,
				ExecutionTime:  time.Since(startTime),
				TimedOut:       true,
			}, err
		}
		return nil, err
	}

	// Удаляем дубликаты токенов
	duplicateStats, err := CleanupDuplicateTokensWithTimeout(db, keepLatest, duplicateTimeout)
	if err != nil {
		if duplicateStats != nil && duplicateStats.TimedOut {
			log.Printf("Очистка дубликатов токенов прервана по таймауту (%s)", duplicateTimeout)
			return &TokenCleanupStats{
				DeletedExpired:    expiredStats.DeletedExpired,
				DeletedDuplicates: 0,
				ExecutionTime:     time.Since(startTime),
				TimedOut:          true,
			}, err
		}
		return nil, err
	}

	// Создаем контекст с таймаутом для VACUUM
	// Примечание: PostgreSQL не поддерживает отмену VACUUM, поэтому контекст здесь не используется
	// Но мы все равно ограничиваем время выполнения для логирования
	vacuumStartTime := time.Now()
	_, err = db.Exec("VACUUM ANALYZE refresh_tokens")
	if time.Since(vacuumStartTime) > vacuumTimeout {
		log.Printf("VACUUM ANALYZE выполнялся дольше установленного таймаута (%s)", vacuumTimeout)
	}
	if err != nil {
		log.Printf("Ошибка при выполнении VACUUM ANALYZE: %v", err)
	}

	executionTime := time.Since(startTime)

	return &TokenCleanupStats{
		DeletedExpired:    expiredStats.DeletedExpired,
		DeletedDuplicates: duplicateStats.DeletedDuplicates,
		TotalRemaining:    duplicateStats.TotalRemaining,
		ExecutionTime:     executionTime,
		TimedOut:          false,
	}, nil
}

// ScheduleTokenCleanup запускает периодическую очистку токенов
func ScheduleTokenCleanup(db *sql.DB, interval time.Duration, keepLatest int) chan bool {
	done := make(chan bool)
	ticker := time.NewTicker(interval)

	go func() {
		// Сразу выполняем очистку при запуске
		stats, err := OptimizeTokenStorageWithTimeout(db, keepLatest, 2*time.Minute)
		if err != nil {
			if stats != nil && stats.TimedOut {
				log.Printf("Первоначальная оптимизация хранения токенов прервана по таймауту")
			} else {
				log.Printf("Ошибка при оптимизации хранения токенов: %v", err)
			}
		} else {
			log.Printf("Оптимизация токенов: удалено %d просроченных, %d дубликатов, осталось %d, время выполнения: %v",
				stats.DeletedExpired, stats.DeletedDuplicates, stats.TotalRemaining, stats.ExecutionTime)
		}

		for {
			select {
			case <-ticker.C:
				stats, err := OptimizeTokenStorageWithTimeout(db, keepLatest, 2*time.Minute)
				if err != nil {
					if stats != nil && stats.TimedOut {
						log.Printf("Периодическая оптимизация хранения токенов прервана по таймауту")
					} else {
						log.Printf("Ошибка при оптимизации хранения токенов: %v", err)
					}
				} else {
					log.Printf("Оптимизация токенов: удалено %d просроченных, %d дубликатов, осталось %d, время выполнения: %v",
						stats.DeletedExpired, stats.DeletedDuplicates, stats.TotalRemaining, stats.ExecutionTime)
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return done
}
