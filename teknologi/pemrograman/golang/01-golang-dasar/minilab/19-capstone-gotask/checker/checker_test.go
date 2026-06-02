package checker

import (
	"testing"
	"time"

	"example.com/minilab/19-capstone-gotask/task"
)

func TestScanOverdue(t *testing.T) {
	tasks := []*task.Task{
		{ID: 1, Title: "Lewat", Status: task.StatusPending, Deadline: time.Now().Add(-24 * time.Hour)},
		{ID: 2, Title: "Oke", Status: task.StatusPending, Deadline: time.Now().Add(24 * time.Hour)},
		{ID: 3, Title: "Selesai lewat", Status: task.StatusDone, Deadline: time.Now().Add(-24 * time.Hour)},
	}

	results := ScanOverdue(tasks)

	if len(results) != 1 {
		t.Errorf("ScanOverdue() = %d; want 1", len(results))
	}
	if results[0].ID != 1 {
		t.Errorf("ID overdue = %d; want 1", results[0].ID)
	}
}

func TestRunCheckerAsync(t *testing.T) {
	tasks := []*task.Task{
		{ID: 1, Title: "Overdue", Status: task.StatusPending, Deadline: time.Now().Add(-1 * time.Hour)},
		{ID: 2, Title: "OK", Status: task.StatusPending, Deadline: time.Now().Add(1 * time.Hour)},
	}

	ch := RunCheckerAsync(tasks)

	var received []*task.Task
	for t := range ch {
		received = append(received, t)
	}

	if len(received) != 1 {
		t.Errorf("RunCheckerAsync hasil = %d; want 1", len(received))
	}
}
