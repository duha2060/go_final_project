package db

import (
	"database/sql"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title VARCHAR(256) NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX idx_date ON scheduler (date);
`

var db *sql.DB

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := errors.Is(err, os.ErrNotExist)
	if err != nil && !install {
		return err
	}

	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if install {
		if _, err := conn.Exec(schema); err != nil {
			conn.Close()
			return err
		}
	}

	db = conn
	return nil
}
