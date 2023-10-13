package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	db *sql.DB
}

var _ Storage = (*DBStorage)(nil)

func NewDBStorage(dbDSN string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("NewDBStorage, Open %w", err)
	}
	return &DBStorage{db: db}, nil
}

func (ds *DBStorage) SaveURL(ctx context.Context, url string) (string, error) {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx. %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO courses.shortener "+
		"(short_url, original_url) VALUES ($1, $2)")
	if err != nil {
		return "", fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()

	sh := generator.RandString(8)
	_, err = stmt.ExecContext(ctx, sh, url)
	if err != nil {
		return "", fmt.Errorf("execContext. %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("tx commit. %w", err)
	}
	return sh, nil
}

func (ds *DBStorage) SaveURLBatch(ctx context.Context, batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx. %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO courses.shortener "+
		"(short_url, original_url) VALUES ($1, $2)")
	if err != nil {
		return nil, fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()
	var bResp []model.BatchRespEntry
	var sh string
	for _, b := range batch {
		sh = generator.RandString(8)
		logger.Log.Info("generated ID is " + string(sh))
		_, err = stmt.ExecContext(ctx, sh, b.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("execContext. %w", err)
		}
		var resp = model.NewBatchRespEntry(b.CorrelationID, sh)
		bResp = append(bResp, resp)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("tx commit. %w", err)
	}
	return bResp, nil
}

func (ds *DBStorage) FindURL(ctx context.Context, shortURL string) (string, error) {
	stmt, err := ds.db.PrepareContext(ctx, "SELECT original_url FROM courses.shortener sh "+
		"WHERE sh.short_url = $1")
	if err != nil {
		return "", fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, shortURL)
	var orig string
	if err := row.Scan(&orig); err != nil {
		return "", fmt.Errorf("cannot scan value. %w", err)
	}
	return orig, nil
}

func (ds *DBStorage) Ping(ctx context.Context) error {
	return ds.db.PingContext(ctx)
}

func (ds *DBStorage) CreateTables() error {
	ddl := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE SCHEMA IF NOT EXISTS courses;
	CREATE TABLE IF NOT EXISTS courses.shortener();
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS id uuid PRIMARY KEY DEFAULT uuid_generate_v4();
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS short_url varchar(8) UNIQUE NOT NULL;
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS original_url varchar UNIQUE NOT NULL;`

	tx, err := ds.db.Begin()
	if err != nil {
		return fmt.Errorf("tx begin. %w", err)
	}
	_, err = tx.Exec(ddl)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("execute ddl. %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx commit. %w", err)
	}
	return nil
}

func (ds *DBStorage) Close() error {
	return ds.db.Close()
}
