package store

import (
	"context"
	"fmt"
)

type URL struct {
	ID          string
	OriginalURL string
	ShortURL    string
}

type ConflictError struct {
	ShortURL string
	Err      error
}

type Reader interface {
	GetURL(ctx context.Context, shortURL string) (string, error)
}

type Writer interface {
	SaveURL(ctx context.Context, shortURL string, fullURL string) error
	SaveBatch(ctx context.Context, batchURL []URL) error
}

func (ce *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %v", ce.Err)
}
