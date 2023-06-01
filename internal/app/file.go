package app

import (
	"encoding/json"
	"github.com/google/uuid"
	"os"
)

type URL struct {
	ID          uuid.UUID `json:"id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type Writer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewFileWriter(fileName string) (*Writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Writer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (w *Writer) WriteFile(file *URL) error {

	return w.encoder.Encode(&file)
}

func (w *Writer) Close() error {
	return w.file.Close()
}

type Reader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewFileReader(fileName string) (*Reader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (r *Reader) ReadFile() (*URL, error) {
	url := &URL{}
	if err := r.decoder.Decode(&url); err != nil {
		return nil, err
	}

	return url, nil
}

func (r *Reader) Close() error {
	return r.file.Close()
}
