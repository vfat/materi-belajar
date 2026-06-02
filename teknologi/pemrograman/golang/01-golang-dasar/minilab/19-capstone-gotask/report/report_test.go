package report

import (
	"strings"
	"testing"
	"time"

	"example.com/minilab/19-capstone-gotask/task"
)

func makeTask(id int, title string, status task.Status, tag string) *task.Task {
	return &task.Task{
		ID:       id,
		Title:    title,
		Status:   status,
		Deadline: time.Now().Add(24 * time.Hour),
		Tag:      tag,
	}
}

func TestTextFormatter_Format(t *testing.T) {
	tasks := []*task.Task{
		makeTask(1, "Belajar Go", task.StatusPending, "go"),
		makeTask(2, "Buat project", task.StatusDone, "project"),
	}

	f := &TextFormatter{}
	output := f.Format(tasks)

	if !strings.Contains(output, "ID") {
		t.Error("output tidak mengandung header ID")
	}
	if !strings.Contains(output, "Belajar Go") {
		t.Error("output tidak mengandung judul task")
	}
	if !strings.Contains(output, "pending") {
		t.Error("output tidak mengandung status")
	}
}

func TestSummary(t *testing.T) {
	tasks := []*task.Task{
		makeTask(1, "T1", task.StatusPending, ""),
		makeTask(2, "T2", task.StatusDone, ""),
		makeTask(3, "T3", task.StatusDone, ""),
	}

	s := BuildSummary(tasks)

	if s.Total != 3 {
		t.Errorf("Total = %d; want 3", s.Total)
	}
	if s.Done != 2 {
		t.Errorf("Done = %d; want 2", s.Done)
	}
	if s.Pending != 1 {
		t.Errorf("Pending = %d; want 1", s.Pending)
	}
}

func TestExportToFile(t *testing.T) {
	tasks := []*task.Task{
		makeTask(1, "Export test", task.StatusPending, "test"),
	}

	tmpFile := t.TempDir() + "/report.txt"
	f := &TextFormatter{}

	err := ExportToFile(f, tasks, tmpFile)
	if err != nil {
		t.Fatalf("ExportToFile error: %v", err)
	}
}
