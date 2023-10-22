package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	db   *sql.DB
	conf config.Conf
}

var _ Storage = (*DBStorage)(nil)

var ErrDBConflict = errors.New("db conflict while executing sql query")

func NewDBStorage(dbDSN string, conf config.Conf) (*DBStorage, error) {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("NewDBStorage, Open %w", err)
	}
	return &DBStorage{
		db:   db,
		conf: conf,
	}, nil
}

func (ds *DBStorage) SaveURL(ctx context.Context, url string) (string, error) {
	stmt, err := ds.db.PrepareContext(ctx, "WITH new_row AS ("+
		"INSERT INTO courses.shortener(short_url, original_url) VALUES ($1, $2) "+
		"ON CONFLICT (original_url) DO NOTHING RETURNING short_url) "+
		"SELECT short_url FROM new_row UNION SELECT short_url FROM courses.shortener "+
		"WHERE courses.shortener.original_url = $2")
	if err != nil {
		return "", fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()

	sh := generator.RandString(8)
	row := stmt.QueryRowContext(ctx, sh, url)
	var res string
	if err = row.Scan(&res); err != nil {
		return "", fmt.Errorf("cannot scan value. %w", err)
	}
	if sh != res {
		err = ErrDBConflict
	}

	return res, err
}

func (ds *DBStorage) SaveURLBatch(ctx context.Context, batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx. %w", err)
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "WITH new_row AS ("+
		"INSERT INTO courses.shortener(short_url, original_url) VALUES ($1, $2) "+
		"ON CONFLICT (original_url) DO NOTHING RETURNING short_url) "+
		"SELECT short_url FROM new_row UNION SELECT short_url FROM courses.shortener "+
		"WHERE courses.shortener.original_url = $2")

	if err != nil {
		return nil, fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()
	var bResp []model.BatchRespEntry
	var sh string
	for _, b := range batch {
		sh = generator.RandString(8)
		row := stmt.QueryRowContext(ctx, sh, b.OriginalURL)
		if err := row.Scan(&sh); err != nil {
			return nil, fmt.Errorf("cannot scan value. %w", err)
		}
		url := fmt.Sprintf("%s/%s", ds.conf.BaseURL(), sh)
		var resp = model.NewBatchRespEntry(b.CorrelationID, url)
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
