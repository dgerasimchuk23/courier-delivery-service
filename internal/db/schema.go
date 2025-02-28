package db

import (
	"database/sql"
	"log"
)

// Создание таблиц в базе данных, если их нет
func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS customer (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS courier (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		phone TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		vehicle_id TEXT,
		status TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS parcel (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (client) REFERENCES customer(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS delivery (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		courier_id INTEGER NOT NULL,
		parcel_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		assigned_at DATETIME NOT NULL,
		delivered_at DATETIME DEFAULT NULL,
		FOREIGN KEY (courier_id) REFERENCES courier(id) ON DELETE CASCADE,
		FOREIGN KEY (parcel_id) REFERENCES parcel(id) ON DELETE CASCADE
	);`
	if _, err := db.Exec(schema); err != nil {
		log.Printf("Ошибка создания таблиц: %v", err)
		return err
	}
	return nil
}
