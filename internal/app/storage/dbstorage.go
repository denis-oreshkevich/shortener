package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	db *sql.DB
}

//var _ Storage = (*FileStorage)(nil)

func NewDBStorage(dbDSN string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("NewDBStorage, Open %w", err)
	}
	return &DBStorage{db: db}, nil
}

func (ds *DBStorage) Ping() error {
	return ds.db.Ping()
}

func (ds *DBStorage) Close() error {
	return ds.db.Close()
}
