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
	);`

	if _, err := db.Exec(schema); err != nil {
		log.Printf("Ошибка создания таблиц: %v", err)
		return err
	}
	return nil
}
