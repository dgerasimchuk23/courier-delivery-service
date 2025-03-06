package db

import (
	"database/sql"
	"log"
	"time"
)

// OptimizeDatabase выполняет оптимизацию базы данных
func OptimizeDatabase(db *sql.DB) error {
	log.Println("Запуск оптимизации базы данных...")

	// Оптимизация таблицы refresh_tokens
	if err := optimizeRefreshTokens(db); err != nil {
		return err
	}

	// Выполняем VACUUM ANALYZE для всех таблиц
	tables := []string{"users", "refresh_tokens", "customer", "courier", "parcel", "delivery"}
	for _, table := range tables {
		_, err := db.Exec("VACUUM ANALYZE " + table)
		if err != nil {
			log.Printf("Ошибка при выполнении VACUUM ANALYZE для таблицы %s: %v", table, err)
		}
	}

	log.Println("Оптимизация базы данных успешно завершена")
	return nil
}

// optimizeRefreshTokens оптимизирует таблицу refresh_tokens
func optimizeRefreshTokens(db *sql.DB) error {
	log.Println("Оптимизация таблицы refresh_tokens...")

	// Удаляем просроченные токены
	result, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < $1`, time.Now())
	if err != nil {
		log.Printf("Ошибка при удалении просроченных токенов: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Удалено %d просроченных токенов", rowsAffected)

	return nil
}

// CheckDatabasePerformance проверяет производительность ключевых запросов
func CheckDatabasePerformance(db *sql.DB) error {
	log.Println("Проверка производительности базы данных...")

	// Проверка производительности запросов к таблице users
	start := time.Now()
	_, err := db.Exec(`EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com'`)
	if err != nil {
		log.Printf("Ошибка при проверке запроса к таблице users: %v", err)
		return err
	}
	log.Printf("Запрос к таблице users выполнен за %v", time.Since(start))

	// Проверка производительности запросов к таблице refresh_tokens
	start = time.Now()
	_, err = db.Exec(`EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE token = 'test_token'`)
	if err != nil {
		log.Printf("Ошибка при проверке запроса к таблице refresh_tokens: %v", err)
		return err
	}
	log.Printf("Запрос к таблице refresh_tokens выполнен за %v", time.Since(start))

	// Проверка производительности запросов к таблице courier
	start = time.Now()
	_, err = db.Exec(`EXPLAIN ANALYZE SELECT * FROM courier WHERE status = 'active'`)
	if err != nil {
		log.Printf("Ошибка при проверке запроса к таблице courier: %v", err)
		return err
	}
	log.Printf("Запрос к таблице courier выполнен за %v", time.Since(start))

	// Проверка производительности запросов к таблице delivery
	start = time.Now()
	_, err = db.Exec(`EXPLAIN ANALYZE SELECT * FROM delivery WHERE courier_id = 1`)
	if err != nil {
		log.Printf("Ошибка при проверке запроса к таблице delivery: %v", err)
		return err
	}
	log.Printf("Запрос к таблице delivery выполнен за %v", time.Since(start))

	log.Println("Проверка производительности базы данных успешно завершена")
	return nil
}
