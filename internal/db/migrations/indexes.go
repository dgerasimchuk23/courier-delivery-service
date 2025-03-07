package db

import (
	"database/sql"
	"log"
	"time"
)

// IndexResult хранит результаты создания индексов
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
	query := `CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`
	if _, err := db.Exec(query); err != nil {
		return 0, err
	}
	return 1, nil
}

// createRefreshTokensIndexes создает индексы для таблицы refresh_tokens
func createRefreshTokensIndexes(db *sql.DB) (int, error) {
	query := `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);`
	if _, err := db.Exec(query); err != nil {
		return 0, err
	}
	return 1, nil
}

// createCustomerIndexes создает индексы для таблицы customer
func createCustomerIndexes(db *sql.DB) (int, error) {
	query := `CREATE INDEX IF NOT EXISTS idx_customer_email ON customer(email);`
	if _, err := db.Exec(query); err != nil {
		return 0, err
	}
	return 1, nil
}

// createCourierIndexes создает индексы для таблицы courier
func createCourierIndexes(db *sql.DB) (int, error) {
	query := `CREATE INDEX IF NOT EXISTS idx_courier_email ON courier(email);`
	if _, err := db.Exec(query); err != nil {
		return 0, err
	}
	return 1, nil
}

// createParcelIndexes создает индексы для таблицы parcel
func createParcelIndexes(db *sql.DB) (int, error) {
	query := `CREATE INDEX IF NOT EXISTS idx_parcel_client ON parcel(client);`
	if _, err := db.Exec(query); err != nil {
		return 0, err
	}
	return 1, nil
}

// createDeliveryIndexes создает индексы для таблицы delivery
func createDeliveryIndexes(db *sql.DB) (int, error) {
	query := `CREATE INDEX IF NOT EXISTS idx_delivery_courier_id ON delivery(courier_id);`
	if _, err := db.Exec(query); err != nil {
		return 0, err
	}
	return 1, nil
}
