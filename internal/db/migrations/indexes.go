package db

import (
	"database/sql"
	"log"
	"time"
)

// IndexResult содержит результаты создания индексов
type IndexResult struct {
	IndicesCreated int
	RowsAffected   int64
	ExecutionTime  time.Duration
}

// InitIndexes создает все необходимые индексы в базе данных
func InitIndexes(db *sql.DB) (*IndexResult, error) {
	log.Println("Создание индексов в базе данных...")
	startTime := time.Now()
	result := &IndexResult{}

	// Индексы для таблицы users
	userResult, err := createUsersIndexes(db)
	if err != nil {
		return result, err
	}
	result.IndicesCreated += userResult

	// Индексы для таблицы refresh_tokens
	refreshResult, err := createRefreshTokensIndexes(db)
	if err != nil {
		return result, err
	}
	result.IndicesCreated += refreshResult

	// Индексы для таблицы customer
	customerResult, err := createCustomerIndexes(db)
	if err != nil {
		return result, err
	}
	result.IndicesCreated += customerResult

	// Индексы для таблицы courier
	courierResult, err := createCourierIndexes(db)
	if err != nil {
		return result, err
	}
	result.IndicesCreated += courierResult

	// Индексы для таблицы parcel
	parcelResult, err := createParcelIndexes(db)
	if err != nil {
		return result, err
	}
	result.IndicesCreated += parcelResult

	// Индексы для таблицы delivery
	deliveryResult, err := createDeliveryIndexes(db)
	if err != nil {
		return result, err
	}
	result.IndicesCreated += deliveryResult

	result.ExecutionTime = time.Since(startTime)
	log.Printf("Индексы успешно созданы: %d индексов, время выполнения: %v", result.IndicesCreated, result.ExecutionTime)
	return result, nil
}

// createUsersIndexes создает индексы для таблицы users
func createUsersIndexes(db *sql.DB) (int, error) {
	indicesCreated := 0

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
		return indicesCreated, err
	}

	// Если индекс не существует, создаем его
	if !exists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на email: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на email успешно создан")
		indicesCreated++
	} else {
		log.Println("Индекс на email уже существует")
	}

	return indicesCreated, nil
}

// createRefreshTokensIndexes создает индексы для таблицы refresh_tokens
func createRefreshTokensIndexes(db *sql.DB) (int, error) {
	indicesCreated := 0
	var rowsAffected int64

	// 1. Добавляем индекс на expires_at для быстрого удаления просроченных токенов
	var expiresExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'refresh_tokens' AND indexname = 'idx_refresh_tokens_expires_at'
		)
	`).Scan(&expiresExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на expires_at: %v", err)
		return indicesCreated, err
	}

	if !expiresExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на expires_at: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на expires_at успешно создан")
		indicesCreated++
	}

	// 2. Добавляем индекс на token для быстрого поиска по токену
	var tokenExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'refresh_tokens' AND indexname = 'idx_refresh_tokens_token'
		)
	`).Scan(&tokenExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на token: %v", err)
		return indicesCreated, err
	}

	if !tokenExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на token: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на token успешно создан")
		indicesCreated++
	}

	// 3. Добавляем индекс на user_id для быстрого поиска токенов пользователя
	var userIdExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'refresh_tokens' AND indexname = 'idx_refresh_tokens_user_id'
		)
	`).Scan(&userIdExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на user_id: %v", err)
		return indicesCreated, err
	}

	if !userIdExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на user_id: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на user_id успешно создан")
		indicesCreated++
	}

	// 4. Удаляем просроченные токены
	result, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < $1`, time.Now())
	if err != nil {
		log.Printf("Ошибка при удалении просроченных токенов: %v", err)
		return indicesCreated, err
	}

	rowsAffected, _ = result.RowsAffected()
	log.Printf("Удалено %d просроченных токенов", rowsAffected)

	return indicesCreated, nil
}

// createCustomerIndexes создает индексы для таблицы customer
func createCustomerIndexes(db *sql.DB) (int, error) {
	indicesCreated := 0

	var emailExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'customer' AND indexname = 'idx_customer_email'
		)
	`).Scan(&emailExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на customer.email: %v", err)
		return indicesCreated, err
	}

	if !emailExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_customer_email ON customer(email)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на customer.email: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на customer.email успешно создан")
		indicesCreated++
	}

	return indicesCreated, nil
}

// createCourierIndexes создает индексы для таблицы courier
func createCourierIndexes(db *sql.DB) (int, error) {
	indicesCreated := 0

	// Индекс на email
	var emailExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'courier' AND indexname = 'idx_courier_email'
		)
	`).Scan(&emailExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на courier.email: %v", err)
		return indicesCreated, err
	}

	if !emailExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_courier_email ON courier(email)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на courier.email: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на courier.email успешно создан")
		indicesCreated++
	}

	// Индекс на status
	var statusExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'courier' AND indexname = 'idx_courier_status'
		)
	`).Scan(&statusExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на courier.status: %v", err)
		return indicesCreated, err
	}

	if !statusExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_courier_status ON courier(status)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на courier.status: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на courier.status успешно создан")
		indicesCreated++
	}

	return indicesCreated, nil
}

// createParcelIndexes создает индексы для таблицы parcel
func createParcelIndexes(db *sql.DB) (int, error) {
	indicesCreated := 0

	// Индекс на status
	var statusExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'parcel' AND indexname = 'idx_parcel_status'
		)
	`).Scan(&statusExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на parcel.status: %v", err)
		return indicesCreated, err
	}

	if !statusExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_parcel_status ON parcel(status)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на parcel.status: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на parcel.status успешно создан")
		indicesCreated++
	}

	// Индекс на client
	var clientExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'parcel' AND indexname = 'idx_parcel_client'
		)
	`).Scan(&clientExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на parcel.client: %v", err)
		return indicesCreated, err
	}

	if !clientExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_parcel_client ON parcel(client)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на parcel.client: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на parcel.client успешно создан")
		indicesCreated++
	}

	return indicesCreated, nil
}

// createDeliveryIndexes создает индексы для таблицы delivery
func createDeliveryIndexes(db *sql.DB) (int, error) {
	indicesCreated := 0

	// Индекс на status
	var statusExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'delivery' AND indexname = 'idx_delivery_status'
		)
	`).Scan(&statusExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на delivery.status: %v", err)
		return indicesCreated, err
	}

	if !statusExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_delivery_status ON delivery(status)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на delivery.status: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на delivery.status успешно создан")
		indicesCreated++
	}

	// Индекс на courier_id
	var courierIdExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'delivery' AND indexname = 'idx_delivery_courier_id'
		)
	`).Scan(&courierIdExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на delivery.courier_id: %v", err)
		return indicesCreated, err
	}

	if !courierIdExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_delivery_courier_id ON delivery(courier_id)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на delivery.courier_id: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на delivery.courier_id успешно создан")
		indicesCreated++
	}

	// Индекс на parcel_id
	var parcelIdExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_indexes 
			WHERE tablename = 'delivery' AND indexname = 'idx_delivery_parcel_id'
		)
	`).Scan(&parcelIdExists)

	if err != nil {
		log.Printf("Ошибка при проверке существования индекса на delivery.parcel_id: %v", err)
		return indicesCreated, err
	}

	if !parcelIdExists {
		_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_delivery_parcel_id ON delivery(parcel_id)`)
		if err != nil {
			log.Printf("Ошибка при создании индекса на delivery.parcel_id: %v", err)
			return indicesCreated, err
		}
		log.Println("Индекс на delivery.parcel_id успешно создан")
		indicesCreated++
	}

	return indicesCreated, nil
}
