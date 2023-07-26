package store

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

type URL struct {
	ID          string
	OriginalURL string
	ShortURL    string
	UserID      uuid.UUID
}

type ConflictError struct {
	ShortURL string
	Err      error
}

type Reader interface {
	GetURL(ctx context.Context, shortURL string) (string, error)
}

type PingableReader interface {
	Reader
	Pingable
}

type Pingable interface {
	Ping(ctx context.Context) error
}

type UsersURLGetter interface {
	GetURLsByUserID(ctx context.Context, userID string) ([]URL, error)
}

type UserIDReader interface {
	Reader
	UsersURLGetter
}

type Writer interface {
	SaveURL(ctx context.Context, URL URL) error
	SaveBatch(ctx context.Context, batchURL []URL) error
}

func (ce *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %v", ce.Err)
}
