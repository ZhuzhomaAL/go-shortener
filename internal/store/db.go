package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"strings"
)

type DBReader struct {
	DB *sql.DB
}

func (dbr *DBReader) GetURL(ctx context.Context, shortURL string) (string, error) {
	var fullURL string
	var deleted bool
	err := dbr.DB.QueryRowContext(
		ctx,
		`SELECT full_url, is_deleted FROM short_url WHERE short_url = $1`, shortURL,
	).Scan(&fullURL, &deleted)
	if err != nil {
		return "", err
	}
	if deleted {
		return "", &DeletedURLError{Err: err}
	}

	return fullURL, nil
}

func (dbr *DBReader) GetURLsByUserID(ctx context.Context, userID string) ([]URL, error) {
	urls := make([]URL, 0)
	rows, err := dbr.DB.QueryContext(
		ctx,
		`SELECT full_url, short_url FROM short_url s WHERE s.user_id = $1`, userID,
	)
	if err != nil {
		return urls, err
	}
	defer rows.Close()

	for rows.Next() {
		var u URL
		err := rows.Scan(&u.OriginalURL, &u.ShortURL)
		if err != nil {
			return urls, err
		}
		urls = append(urls, u)
	}
	err = rows.Err()
	if err != nil {
		return urls, err
	}

	return urls, nil
}

func (dbr *DBReader) Ping(ctx context.Context) error {

	return dbr.DB.Ping()
}

type DBWriter struct {
	DB *sql.DB
}

func (dbw *DBWriter) SaveURL(ctx context.Context, URL URL) error {
	query := `INSERT INTO short_url(full_url, short_url, user_id) VALUES ($1, $2, $3)`
	stmt, err := dbw.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, URL.OriginalURL, URL.ShortURL, URL.UserID)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == pgerrcode.UniqueViolation {
			short, err := getShortURLByFull(ctx, dbw.DB, URL.OriginalURL)
			if err != nil {
				return err
			}
			return &ConflictError{ShortURL: short, Err: err}
		}
		return err
	}

	return nil
}

func split(batchURL []URL, size int) [][]URL {
	var chunks [][]URL
	if len(batchURL) <= size {
		chunks = append(chunks, batchURL)
		return chunks
	}
	for i := 0; i < len(batchURL); i += size {
		end := i + size
		if end > len(batchURL) {
			end = len(batchURL)
		}
		chunks = append(chunks, batchURL[i:end])
	}
	return chunks
}

func (dbw *DBWriter) SaveBatch(ctx context.Context, batchURL []URL) error {
	chunks := split(batchURL, 1000)
	tx, err := dbw.DB.Begin()
	if err != nil {
		return err
	}
	for _, chunk := range chunks {
		query := "INSERT INTO short_url(full_url, short_url, user_id) VALUES "
		var inserts []string
		var params []interface{}
		var i int
		for _, u := range chunk {
			inserts = append(inserts, fmt.Sprintf("($%d, $%d, $%d)", i+1, i+2, i+3))
			i = i + 3
			params = append(params, u.OriginalURL, u.ShortURL, u.UserID.String())
		}
		queryVals := strings.Join(inserts, ",")
		query = query + queryVals
		_, err := tx.ExecContext(ctx, query, params...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func getShortURLByFull(ctx context.Context, db *sql.DB, fullURL string) (string, error) {
	var shortURL string
	err := db.QueryRowContext(
		ctx,
		`SELECT short_url FROM short_url WHERE full_url = $1`, fullURL,
	).Scan(&shortURL)
	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (dbw *DBWriter) DeleteURLs(ctx context.Context, batchURL []URL) error {
	chunks := split(batchURL, 1000)
	tx, err := dbw.DB.Begin()
	if err != nil {
		return err
	}
	queryTpl := `WITH _data(short_url) AS (
    VALUES
        %s
	)
	UPDATE short_url AS s
	SET is_deleted = true
	FROM _data
	WHERE s.short_url = _data.short_url`
	for _, chunk := range chunks {
		var inserts []string
		var params []interface{}
		for i, u := range chunk {
			inserts = append(inserts, fmt.Sprintf("($%d)", i+1))
			params = append(params, u.ShortURL)
		}
		queryVals := strings.Join(inserts, ",")
		query := fmt.Sprintf(queryTpl, queryVals)
		_, err := tx.ExecContext(ctx, query, params...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (dbr *DBReader) FilterURLsByUserID(ctx context.Context, userID string, URLs []URL) ([]URL, error) {
	queryTpl := `SELECT short_url FROM short_url s WHERE s.user_id = $1 AND s.short_url IN(%s) AND s.is_deleted = false`
	var inserts []string
	var params []interface{}
	params = append(params, userID)
	for i, u := range URLs {
		inserts = append(inserts, fmt.Sprintf("$%d", i+2))
		params = append(params, u.ShortURL)
	}
	queryVals := strings.Join(inserts, ",")
	query := fmt.Sprintf(queryTpl, queryVals)
	urls := make([]URL, 0)
	rows, err := dbr.DB.QueryContext(ctx, query, params...)
	if err != nil {
		return urls, err
	}
	defer rows.Close()

	for rows.Next() {
		var u URL
		err := rows.Scan(&u.ShortURL)
		if err != nil {
			return urls, err
		}
		urls = append(urls, u)
	}
	err = rows.Err()
	if err != nil {
		return urls, err
	}

	return urls, nil
}
