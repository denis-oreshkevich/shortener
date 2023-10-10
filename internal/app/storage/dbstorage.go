package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
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

func (ds *DBStorage) SaveURL(url string) (string, error) {
	sh := generator.RandString(8)
	_, err := ds.db.ExecContext(context.TODO(), "INSERT INTO courses.shortener "+
		"(short_url, original_url) VALUES ($1, $2)", sh, url)
	if err != nil {
		return "", fmt.Errorf("saveURL execContext. %w", err)
	}
	return sh, err
}

func (ds *DBStorage) FindURL(shortURL string) (string, error) {
	row := ds.db.QueryRowContext(context.TODO(), "SELECT original_url FROM courses.shortener sh "+
		"where sh.short_url = $1", shortURL)
	var orig string
	if err := row.Scan(&orig); err != nil {
		return "", fmt.Errorf("cannot scan value. %w", err)
	}
	return orig, nil
}

func (ds *DBStorage) Ping() error {
	return ds.db.Ping()
}

func (ds *DBStorage) CreateTables() error {
	ddl := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE SCHEMA IF NOT EXISTS courses;
	CREATE TABLE IF NOT EXISTS courses.shortener();
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS id uuid PRIMARY KEY DEFAULT uuid_generate_v4();
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS short_url varchar(8) UNIQUE NOT NULL;
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS original_url varchar UNIQUE NOT NULL;`
	_, err := ds.db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("execute ddl. %w", err)
	}
	return nil
}

func (ds *DBStorage) Close() error {
	return ds.db.Close()
}
