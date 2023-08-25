package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type MemoryReader struct {
	URLList *sync.Map
}

func (mr *MemoryReader) GetURL(ctx context.Context, shortURL string) (string, error) {
	fullURL, ok := mr.URLList.Load(shortURL)
	if !ok {
		return "", errors.New("failed to load fullURL from memory")
	}
	return fmt.Sprintf("%v", fullURL), nil
}

type MemoryWriter struct {
	URLList *sync.Map
}

func (mw *MemoryWriter) SaveURL(ctx context.Context, URL URL) error {
	mw.URLList.Store(URL.ShortURL, URL.OriginalURL)
	return nil
}

func (mw *MemoryWriter) SaveBatch(ctx context.Context, batchURL []URL) error {
	for _, URL := range batchURL {
		mw.URLList.Store(URL.ShortURL, URL.OriginalURL)
	}

	return nil
}
