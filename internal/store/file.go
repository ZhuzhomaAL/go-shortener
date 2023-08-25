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

func (fw *FileWriter) SaveURL(ctx context.Context, URL URL) error {
	err := fw.MemoryWriter.SaveURL(ctx, URL)
	if err != nil {
		return err
	}
	id := uuid.New()
	fileURL := &file.URL{
		ID:          id,
		ShortURL:    URL.ShortURL,
		OriginalURL: URL.OriginalURL,
	}

	return fw.Writer.WriteFile(fileURL)
}

func (fw *FileWriter) SaveBatch(ctx context.Context, batchURL []URL) error {
	for _, item := range batchURL {
		err := fw.SaveURL(ctx, item)
		if err != nil {
			return err
		}
	}

	return nil
}
