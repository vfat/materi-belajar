package task

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	deadline := time.Now().Add(48 * time.Hour)
	task := NewTask(1, "Belajar Go", deadline, "belajar")

	if task.ID != 1 {
		t.Errorf("ID = %d; want 1", task.ID)
	}
	if task.Title != "Belajar Go" {
		t.Errorf("Title = %q; want %q", task.Title, "Belajar Go")
	}
	if task.Status != StatusPending {
		t.Errorf("Status = %q; want %q", task.Status, StatusPending)
	}
	if task.Tag != "belajar" {
		t.Errorf("Tag = %q; want %q", task.Tag, "belajar")
	}
}

func TestTask_IsOverdue(t *testing.T) {
	tests := []struct {
		name     string
		deadline time.Time
		want     bool
	}{
		{"sudah lewat", time.Now().Add(-24 * time.Hour), true},
		{"belum lewat", time.Now().Add(24 * time.Hour), false},
		{"tepat sekarang (dianggap belum)", time.Now().Add(time.Minute), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(1, "Test", tt.deadline, "")
			if got := task.IsOverdue(); got != tt.want {
				t.Errorf("IsOverdue() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestTask_Complete(t *testing.T) {
	task := NewTask(1, "Test", time.Now().Add(time.Hour), "")
	task.Complete()

	if task.Status != StatusDone {
		t.Errorf("Status setelah Complete() = %q; want %q", task.Status, StatusDone)
	}
}
