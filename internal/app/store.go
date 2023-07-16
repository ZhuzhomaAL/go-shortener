package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

func saveURL(db *sql.DB, fWriter *Writer, urlList *sync.Map, shortURL string, fullURL string) error {
	switch {
	case db != nil:
		err := saveToDB(db, fullURL, shortURL)
		if err != nil {
			return err
		}
	case fWriter != nil:
		err := saveToFile(fWriter, shortURL, fullURL)
		if err != nil {
			return err
		}
		fallthrough
	case urlList != nil:
		err := saveToMemory(shortURL, fullURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func getURL(db *sql.DB, shortURL string) (string, error) {
	if db != nil {
		return readFromDB(db, shortURL)
	} else {
		return readFromMemory(shortURL)
	}
}

func readFromMemory(shortURL string) (string, error) {
	fullURL, ok := urlList.Load(shortURL)
	if !ok {
		return "", errors.New("failed to load fullURL from memory")
	}
	return fmt.Sprintf("%v", fullURL), nil
}

func readFromDB(db *sql.DB, shortURL string) (string, error) {
	var fullURL string
	err := db.QueryRowContext(
		context.Background(),
		`SELECT full_url FROM short_url WHERE short_url = $1`, shortURL,
	).Scan(&fullURL)
	if err != nil {
		return "", err
	}

	return fullURL, nil
}

func saveToMemory(shortURL string, fullURL string) error {
	urlList.Store(shortURL, fullURL)
	return nil
}

func saveToFile(writer *Writer, shortURL string, fullURL string) error {
	id := uuid.New()
	fileURL := &URL{
		id,
		shortURL,
		fullURL,
	}

	return writer.WriteFile(fileURL)

}

func saveToDB(db *sql.DB, fullURL string, shortURL string) error {
	query := `INSERT INTO short_url(full_url, short_url) VALUES ($1, $2)`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, fullURL, shortURL)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func saveBatch(db *sql.DB, fWriter *Writer, urlList *sync.Map, batchURL []batchURL) error {
	switch {
	case db != nil:
		err := saveBatchToDB(db, batchURL)
		if err != nil {
			return err
		}
	case fWriter != nil:
		err := saveBatchToFile(fWriter, batchURL)
		if err != nil {
			return err
		}
		fallthrough
	case urlList != nil:
		err := saveBatchToMemory(batchURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveBatchToMemory(batchURL []batchURL) error {
	for _, URL := range batchURL {
		urlList.Store(URL.ShortURL, URL.OriginalURL)
	}

	return nil
}

func saveBatchToFile(writer *Writer, batchURL []batchURL) error {
	for _, item := range batchURL {
		id := uuid.New()
		fileURL := &URL{
			id,
			item.ShortURL,
			item.OriginalURL,
		}
		err := writer.WriteFile(fileURL)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveBatchToDB(db *sql.DB, batchURL []batchURL) error {
	query := "INSERT INTO short_url(full_url, short_url) VALUES "
	var inserts []string
	var params []interface{}
	var i int
	for _, u := range batchURL {
		inserts = append(inserts, fmt.Sprintf("($%d, $%d)", i+1, i+2))
		i = i + 2
		params = append(params, u.OriginalURL, u.ShortURL)
	}
	queryVals := strings.Join(inserts, ",")
	query = query + queryVals
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, params...)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}
