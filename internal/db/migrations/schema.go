package db

import (
	"database/sql"
	"log"
)

// Создание таблиц в базе данных, если их нет
func InitSchema(db *sql.DB, dbType string) error {
	schema := `
	CREATE TABLE IF NOT EXISTS customer (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS courier (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		phone TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		vehicle_id TEXT,
		status TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS parcel (
		id SERIAL PRIMARY KEY,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (client) REFERENCES customer(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS delivery (
		id SERIAL PRIMARY KEY,
		courier_id INTEGER NOT NULL,
		parcel_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		assigned_at TIMESTAMP NOT NULL,
		delivered_at TIMESTAMP DEFAULT NULL,
		FOREIGN KEY (courier_id) REFERENCES courier(id) ON DELETE CASCADE,
		FOREIGN KEY (parcel_id) REFERENCES parcel(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'client',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Printf("Ошибка создания таблиц: %v", err)
		return err
	}

	// Проверяем, существует ли колонка role в таблице users
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'role'
		)
	`).Scan(&exists)

	if err != nil {
		log.Printf("Ошибка при проверке существования колонки role: %v", err)
		return err
	}

	// Если колонка role не существует, добавляем ее
	if !exists {
		_, err := db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'client'`)
		if err != nil {
			log.Printf("Ошибка при добавлении колонки role: %v", err)
			return err
		}
		log.Println("Колонка role успешно добавлена в таблицу users")
	}

	return nil
}
