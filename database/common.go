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
	    fid INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		uid INTEGER,
		share_code TEXT,
		FOREIGN KEY(uid) REFERENCES user(uid)
	)`
	createFileAccessTableSQL := `
	CREATE TABLE IF NOT EXISTS user_access (
		user_id INTEGER,
		file_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES user(uid),
		FOREIGN KEY(file_id) REFERENCES file(fid),
		PRIMARY KEY (user_id, file_id)
	)`
	_, err := db.Exec(createUserTableSQL)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to create user table")
	}

	_, err = db.Exec(createFileTableSQL)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to create file table")
	}

	_, err = db.Exec(createFileAccessTableSQL)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to create file_access table")
	}
}
