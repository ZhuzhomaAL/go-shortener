package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type MemoryReader struct {
	UrlList *sync.Map
}

func (mr *MemoryReader) GetURL(ctx context.Context, shortURL string) (string, error) {
	fullURL, ok := mr.UrlList.Load(shortURL)
	if !ok {
		return "", errors.New("failed to load fullURL from memory")
	}
	return fmt.Sprintf("%v", fullURL), nil
}

type MemoryWriter struct {
	UrlList *sync.Map
}

func (mw *MemoryWriter) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	mw.UrlList.Store(shortURL, fullURL)
	return nil
}

func (mw *MemoryWriter) SaveBatch(ctx context.Context, batchURL []URL) error {
	for _, URL := range batchURL {
		mw.UrlList.Store(URL.ShortURL, URL.OriginalURL)
	}

	return nil
}
