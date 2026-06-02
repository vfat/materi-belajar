package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"example.com/minilab/19-capstone-gotask/task"
)

// Storage menangani baca/tulis task ke file JSON
type Storage struct {
	path string
}

// NewStorage membuat Storage dengan path file yang ditentukan
func NewStorage(path string) *Storage {
	return &Storage{path: path}
}

// Save menyimpan slice task ke file JSON
func (s *Storage) Save(tasks []*task.Task) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// Load membaca task dari file JSON
// Jika file tidak ada, return slice kosong (bukan error)
func (s *Storage) Load() ([]*task.Task, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*task.Task{}, nil
		}
		return nil, err
	}

	var tasks []*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
