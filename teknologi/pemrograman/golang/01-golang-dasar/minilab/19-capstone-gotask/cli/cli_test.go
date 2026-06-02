package cli

import (
	"strings"
	"testing"
	"time"

	"example.com/minilab/19-capstone-gotask/task"
)

func newTestApp(tmpDir string) *App {
	dataFile := tmpDir + "/tasks.json"
	return NewApp(dataFile)
}

func TestApp_HandleAdd(t *testing.T) {
	app := newTestApp(t.TempDir())

	deadline := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	output, err := app.Handle([]string{"add", "Judul task", "--deadline", deadline, "--tag", "belajar"})

	if err != nil {
		t.Fatalf("Handle add error: %v", err)
	}
	if !strings.Contains(output, "ditambahkan") {
		t.Errorf("output tidak mengandung 'ditambahkan': %q", output)
	}
}

func TestApp_HandleList(t *testing.T) {
	app := newTestApp(t.TempDir())

	deadline := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	app.Handle([]string{"add", "Task satu", "--deadline", deadline, "--tag", "go"})
	app.Handle([]string{"add", "Task dua", "--deadline", deadline, "--tag", "belajar"})

	output, err := app.Handle([]string{"list"})
	if err != nil {
		t.Fatalf("Handle list error: %v", err)
	}
	if !strings.Contains(output, "Task satu") {
		t.Error("output tidak mengandung 'Task satu'")
	}
}

func TestApp_HandleDone(t *testing.T) {
	app := newTestApp(t.TempDir())
	deadline := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	app.Handle([]string{"add", "Selesaikan ini", "--deadline", deadline, "--tag", ""})

	output, err := app.Handle([]string{"done", "1"})
	if err != nil {
		t.Fatalf("Handle done error: %v", err)
	}
	if !strings.Contains(output, "selesai") {
		t.Errorf("output tidak mengandung 'selesai': %q", output)
	}
}

func TestApp_HandlePerintahTidakDikenal(t *testing.T) {
	app := newTestApp(t.TempDir())
	_, err := app.Handle([]string{"tidakada"})
	if err == nil {
		t.Error("harus error untuk perintah tidak dikenal")
	}
}

func TestApp_HandleAdd_TanpaJudul(t *testing.T) {
	app := newTestApp(t.TempDir())
	_, err := app.Handle([]string{"add"})
	if err == nil {
		t.Error("harus error jika judul tidak diberikan")
	}
}

func TestApp_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	deadline := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	app1 := newTestApp(tmpDir)
	app1.Handle([]string{"add", "Persisten task", "--deadline", deadline, "--tag", "test"})

	app2 := newTestApp(tmpDir)
	output, err := app2.Handle([]string{"list"})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if !strings.Contains(output, "Persisten task") {
		t.Errorf("task tidak persisten, output: %q", output)
	}
}

var _ interface{ Format(tasks []*task.Task) string } = nil
