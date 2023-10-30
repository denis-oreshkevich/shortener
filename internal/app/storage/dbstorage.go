package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strconv"
	"strings"
)

type DBStorage struct {
	db *sql.DB
}

var _ Storage = (*DBStorage)(nil)

var ErrDBConflict = errors.New("db conflict while executing sql query")

func NewDBStorage(dbDSN string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("NewDBStorage, Open %w", err)
	}
	return &DBStorage{
		db: db,
	}, nil
}

func (ds *DBStorage) SaveURL(ctx context.Context, userID string, url string) (string, error) {
	stmt, err := ds.db.PrepareContext(ctx, "WITH new_row AS ("+
		"INSERT INTO courses.shortener(short_url, original_url, user_id) VALUES ($1, $2, $3) "+
		"ON CONFLICT (original_url) DO NOTHING RETURNING short_url) "+
		"SELECT short_url FROM new_row UNION SELECT short_url FROM courses.shortener "+
		"WHERE courses.shortener.original_url = $2")
	if err != nil {
		return "", fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()

	sh := generator.RandString(8)
	row := stmt.QueryRowContext(ctx, sh, url, userID)
	var res string
	if err = row.Scan(&res); err != nil {
		return "", fmt.Errorf("cannot scan value. %w", err)
	}
	if sh != res {
		err = ErrDBConflict
	}

	return res, err
}

func (ds *DBStorage) SaveURLBatch(ctx context.Context, userID string,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx. %w", err)
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "WITH new_row AS ("+
		"INSERT INTO courses.shortener(short_url, original_url, user_id) VALUES ($1, $2, $3) "+
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
		row := stmt.QueryRowContext(ctx, sh, b.OriginalURL, userID)
		if err := row.Scan(&sh); err != nil {
			return nil, fmt.Errorf("cannot scan value. %w", err)
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

func (ds *DBStorage) FindURL(ctx context.Context, shortURL string) (*OrigURL, error) {
	stmt, err := ds.db.PrepareContext(ctx, "SELECT original_url, is_deleted "+
		"FROM courses.shortener sh WHERE sh.short_url = $1")
	if err != nil {
		return nil, fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, shortURL)
	orig := &OrigURL{}
	if err := row.Scan(&orig.OriginalURL, &orig.DeletedFlag); err != nil {
		return nil, fmt.Errorf("cannot scan value. %w", err)
	}
	if orig.DeletedFlag {
		return nil, ErrResultIsDeleted
	}
	return orig, nil
}

func (ds *DBStorage) FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error) {
	stmt, err := ds.db.PrepareContext(ctx, "SELECT short_url, original_url "+
		"FROM courses.shortener sh WHERE sh.user_id = $1")
	if err != nil {
		return nil, fmt.Errorf("prepare context. %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query context. %w", err)
	}
	defer rows.Close()
	var res = make([]model.URLPair, 0)
	for rows.Next() {
		var sh string
		var orig string
		if err := rows.Scan(&sh, &orig); err != nil {
			return nil, fmt.Errorf("cannot scan value. %w", err)
		}
		p := model.NewURLPair(sh, orig)
		res = append(res, p)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err(). %w", err)
	}
	return res, nil
}

func (ds *DBStorage) Ping(ctx context.Context) error {
	return ds.db.PingContext(ctx)
}

func (ds *DBStorage) DeleteUserURLs(ctx context.Context, bde model.BatchDeleteEntry) error {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx. %w", err)
	}
	defer tx.Rollback()
	var errs []error
	template := "update courses.shortener set is_deleted = true " +
		"where user_id = $1 and short_url in ($2%s)"

	q := ds.buildDeleteQuery(bde, template)
	iDs := buildIDs(bde)

	if _, err := tx.ExecContext(ctx, q, iDs...); err != nil {
		errs = append(errs, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx commit. %w", err)
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func buildIDs(it model.BatchDeleteEntry) []any {
	var iDs = make([]any, len(it.ShortIDs)+1)
	iDs[0] = it.UserID
	for i := 1; i < len(iDs); i++ {
		iDs[i] = it.ShortIDs[i-1]
	}
	return iDs
}

func (ds *DBStorage) buildDeleteQuery(it model.BatchDeleteEntry, template string) string {
	l := len(it.ShortIDs)
	builder := strings.Builder{}
	for i := 3; i <= l+1; i++ {
		builder.WriteString(", $")
		builder.WriteString(strconv.Itoa(i))
	}
	return fmt.Sprintf(template, builder.String())
}

func (ds *DBStorage) CreateTables() error {
	ddl := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE SCHEMA IF NOT EXISTS courses;
	CREATE TABLE IF NOT EXISTS courses.shortener();
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS id uuid PRIMARY KEY DEFAULT uuid_generate_v4();
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS short_url varchar(8) UNIQUE NOT NULL;
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS original_url varchar UNIQUE NOT NULL;
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS user_id uuid NOT NULL;
	ALTER TABLE courses.shortener ADD COLUMN IF NOT EXISTS is_deleted boolean NOT NULL default false;`

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
