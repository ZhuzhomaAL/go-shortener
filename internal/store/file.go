package store

import (
	"context"
	"github.com/ZhuzhomaAL/go-shortener/internal/file"
	"github.com/google/uuid"
)

type FileReader struct {
	MemoryReader *MemoryReader
}

func (fr *FileReader) GetURL(ctx context.Context, shortURL string) (string, error) {
	return fr.MemoryReader.GetURL(ctx, shortURL)
}

type FileWriter struct {
	MemoryWriter *MemoryWriter
	Writer       *file.Writer
}

func (fw *FileWriter) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	err := fw.MemoryWriter.SaveURL(ctx, shortURL, fullURL)
	if err != nil {
		return err
	}
	id := uuid.New()
	fileURL := &file.URL{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: fullURL,
	}

	return fw.Writer.WriteFile(fileURL)
}

func (fw *FileWriter) SaveBatch(ctx context.Context, batchURL []URL) error {
	for _, item := range batchURL {
		err := fw.SaveURL(ctx, item.ShortURL, item.OriginalURL)
		if err != nil {
			return err
		}
	}

	return nil
}
