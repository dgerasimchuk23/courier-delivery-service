package db

import (
	"database/sql"
	"log"
	"time"
)

// MigrateDB выполняет миграции базы данных
func MigrateDB(db *sql.DB) error {
	log.Println("Запуск миграций базы данных...")

	// Миграция 1: Добавление индекса на email в таблице users
	if err := addEmailIndex(db); err != nil {
		return err
	}

	// Миграция 2: Оптимизация таблицы refresh_tokens
	if err := optimizeRefreshTokens(db); err != nil {
		return err
	}

	// Миграция 3: Добавление индексов для ускорения запросов
	if err := addPerformanceIndices(db); err != nil {
		return err
	}

	log.Println("Миграции успешно выполнены")
	return nil
}

// addEmailIndex добавляет индекс на поле email в таблице users
func addEmailIndex(db *sql.DB) error {
	log.Println("Добавление индекса на email в таблице users...")

	// Проверяем, существует ли индекс
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'users' AND indexname = 'idx_users_email'
		)
	`).Scan(&exists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса: %v", err)
		return err
	}

	// Если индекс не существует, создаем его
	if !exists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на email: %v", err)
			return err
		}
		log.Println("Индекс на email успешно создан")
	} else {
		log.Println("Индекс на email уже существует")
	}

	return nil
}

// optimizeRefreshTokens оптимизирует таблицу refresh_tokens
func optimizeRefreshTokens(db *sql.DB) error {
	log.Println("Оптимизация таблицы refresh_tokens...")

	// 1. Добавляем индекс на expires_at для быстрого удаления просроченных токенов
	_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на expires_at: %v", err)
		return err
	}

	// 2. Добавляем индекс на token для быстрого поиска по токену
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на token: %v", err)
		return err
	}

	// 3. Добавляем индекс на user_id для быстрого поиска токенов пользователя
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на user_id: %v", err)
		return err
	}

	// 4. Удаляем просроченные токены
	result, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < $1`, time.Now())
	if err != nil {
		log.Printf("Ошибка при удалении просроченных токенов: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Удалено %d просроченных токенов", rowsAffected)

	return nil
}

// addPerformanceIndices добавляет индексы для ускорения запросов
func addPerformanceIndices(db *sql.DB) error {
	log.Println("Добавление индексов для ускорения запросов...")

	// Индексы для таблицы customer
	_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_customer_email ON customer(email)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на customer.email: %v", err)
		return err
	}

	// Индексы для таблицы courier
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_courier_email ON courier(email)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на courier.email: %v", err)
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_courier_status ON courier(status)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на courier.status: %v", err)
		return err
	}

	// Индексы для таблицы parcel
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_parcel_status ON parcel(status)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на parcel.status: %v", err)
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_parcel_client ON parcel(client)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на parcel.client: %v", err)
		return err
	}

	// Индексы для таблицы delivery
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_delivery_status ON delivery(status)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на delivery.status: %v", err)
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_delivery_courier_id ON delivery(courier_id)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на delivery.courier_id: %v", err)
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_delivery_parcel_id ON delivery(parcel_id)`)
	if err != nil {
		log.Printf("Ошибка при создании индекса на delivery.parcel_id: %v", err)
		return err
	}

	return nil
}
