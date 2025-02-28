package parcel

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func setupParcelTestDB() *ParcelStore {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	_, _ = db.Exec("DROP TABLE IF EXISTS parcels")

	createTable := `
    CREATE TABLE parcels (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        client_id INTEGER, -- БЫЛО "client", теперь исправлено
        status TEXT,
        address TEXT,
        created_at TEXT
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	return NewParcelStore(db)
}
