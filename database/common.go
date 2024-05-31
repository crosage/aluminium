package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

var db *sql.DB

func InitDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to open database")
	}
	createTables()
}

func createTables() {
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS user (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		passhash TEXT,
		type INTEGER
	)`

	createFileTableSQL := `
	CREATE TABLE IF NOT EXISTS file (
		hash TEXT NOT NULL,
		path TEXT NOT NULL,
		uid INTEGER,
		FOREIGN KEY(uid) REFERENCES users(uid)
	)`

	createFilePermissionTableSQL := `
    CREATE TABLE IF NOT EXISTS file_permission (
        file_hash TEXT,
        uid INTEGER,
        FOREIGN KEY(file_hash) REFERENCES file(hash),
        FOREIGN KEY(uid) REFERENCES user(uid)
    )`
	_, err := db.Exec(createUserTableSQL)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to create user table")
	}

	_, err = db.Exec(createFileTableSQL)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to create file table")
	}

	_, err = db.Exec(createFilePermissionTableSQL)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to create file_permission table")
	}
}
