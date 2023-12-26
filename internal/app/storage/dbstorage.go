package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"sync"

	"github.com/denis-oreshkevich/shortener/migration"

	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// DBStorage database storage.
type DBStorage struct {
	db *sql.DB
}

var _ Storage = (*DBStorage)(nil)

var (
	db     *sql.DB
	pgOnce sync.Once

	dbErr error
)

// ErrDBConflict error happens on DB conflict.
var ErrDBConflict = errors.New("db conflict while executing sql query")

// NewDBStorage creates new [*DBStorage].
func NewDBStorage(dbDSN string) (*DBStorage, error) {
	pool, err := initDatasource(dbDSN)
	if err != nil {
		return nil, fmt.Errorf("initPool: %w", err)
	}
	return &DBStorage{
		db: pool,
	}, nil
}

func initDatasource(dbDSN string) (*sql.DB, error) {
	pgOnce.Do(func() {
		pool, err := sql.Open("pgx", dbDSN)
		if err != nil {
			dbErr = fmt.Errorf("pgxpool.New: %w", err)
			return
		}
		if err = pool.Ping(); err != nil {
			dbErr = fmt.Errorf("pool.Ping: %w", err)
			return
		}
		if err = applyMigration(dbDSN, migration.SQLFiles); err != nil {
			dbErr = fmt.Errorf("applyMigration: %w", err)
			return
		}
		db = pool
	})
	return db, dbErr
}

// SaveURL saves original URL to DB and returns short URL.
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

	sh := generator.RandString()
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

// SaveURLBatch saves many URLs to DB and return [[]model.BatchRespEntry] back.
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
		sh = generator.RandString()
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

// FindURL finds original URL in DB by short ID.
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

// FindUserURLs finds user's URLs in DB.
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

// Ping pings DB.
func (ds *DBStorage) Ping(ctx context.Context) error {
	return ds.db.PingContext(ctx)
}

// DeleteUserURLs deletes user's URLs.
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

// applyMigration patches DB.
func applyMigration(dsn string, fsys fs.FS) error {
	//TODO ask about conv between pool
	//db := stdlib.OpenDBFromPool(db, nil)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(fsys)
	goose.SetSequential(true)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose.SetDialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}
	return nil
}

// Close closes connection pool.
func (ds *DBStorage) Close() error {
	return ds.db.Close()
}
