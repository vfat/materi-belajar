package task

import "time"

// Status task yang mungkin
type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
)

// Task merepresentasikan satu tugas
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	Deadline  time.Time `json:"deadline"`
	Tag       string    `json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTask membuat Task baru dengan status pending
func NewTask(id int, title string, deadline time.Time, tag string) *Task {
	return &Task{
		ID:        id,
		Title:     title,
		Status:    StatusPending,
		Deadline:  deadline,
		Tag:       tag,
		CreatedAt: time.Now(),
	}
}

// IsOverdue mengembalikan true jika deadline sudah lewat dan task belum selesai
func (t *Task) IsOverdue() bool {
	if t.Status == StatusDone {
		return false
	}
	return time.Now().After(t.Deadline)
}

// Complete mengubah status task menjadi done
func (t *Task) Complete() {
	t.Status = StatusDone
}
