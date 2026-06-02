package storage

import (
	"os"
	"testing"
	"time"

	"example.com/minilab/19-capstone-gotask/task"
)

func TestStorage_SaveLoad(t *testing.T) {
	tmpFile := t.TempDir() + "/tasks_test.json"
	store := NewStorage(tmpFile)

	deadline := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	tasks := []*task.Task{
		{ID: 1, Title: "Test Task", Status: task.StatusPending, Deadline: deadline, Tag: "test"},
	}

	err := store.Save(tasks)
	if err != nil {
		t.Fatalf("Save error: %v", err)
	}

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("file tidak terbentuk setelah Save")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Load() = %d tasks; want 1", len(loaded))
	}
	if loaded[0].Title != "Test Task" {
		t.Errorf("Title = %q; want %q", loaded[0].Title, "Test Task")
	}
	if loaded[0].Status != task.StatusPending {
		t.Errorf("Status = %q; want pending", loaded[0].Status)
	}
}

func TestStorage_LoadFileKosong(t *testing.T) {
	tmpFile := t.TempDir() + "/tidak_ada.json"
	store := NewStorage(tmpFile)

	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("Load file tidak ada harus return kosong, bukan error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Load() = %d; want 0", len(tasks))
	}
}
