package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lucasgiovanella/secret-go/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	// Cria o arquivo se não existir
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar banco: %w", err)
		}
		file.Close()
	}
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, err
	}
	return db, nil
}

func (d *DB) initSchema() error {
	_, err := d.conn.Exec(`
	CREATE TABLE IF NOT EXISTS store (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		data BLOB NOT NULL
	);
	`)
	return err
}

func (d *DB) LoadStore() (*model.Store, error) {
	row := d.conn.QueryRow("SELECT data FROM store WHERE id = 1")
	var data []byte
	err := row.Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var store model.Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func (d *DB) SaveStore(store *model.Store) error {
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec(`
	INSERT INTO store (id, data) VALUES (1, ?)
	ON CONFLICT(id) DO UPDATE SET data=excluded.data
	`, data)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}
