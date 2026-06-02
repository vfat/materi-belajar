package task

import (
	"errors"
	"time"
)

// Manager mengelola koleksi task
type Manager struct {
	tasks  []*Task
	nextID int
}

// NewManager membuat Manager baru
func NewManager() *Manager {
	return &Manager{nextID: 1}
}

// AddTask menambah task baru ke Manager
func (m *Manager) AddTask(title string, deadline time.Time, tag string) (*Task, error) {
	if title == "" {
		return nil, errors.New("title tidak boleh kosong")
	}
	task := NewTask(m.nextID, title, deadline, tag)
	m.tasks = append(m.tasks, task)
	m.nextID++
	return task, nil
}

// CompleteTask menandai task sebagai selesai berdasarkan ID
func (m *Manager) CompleteTask(id int) error {
	task := m.findByID(id)
	if task == nil {
		return errors.New("task tidak ditemukan")
	}
	task.Complete()
	return nil
}

// DeleteTask menghapus task berdasarkan ID
func (m *Manager) DeleteTask(id int) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return errors.New("task tidak ditemukan")
}

// ListTasks mengembalikan semua task, bisa difilter berdasar tag
func (m *Manager) ListTasks(filterTag string) []*Task {
	if filterTag == "" {
		return m.tasks
	}
	var result []*Task
	for _, t := range m.tasks {
		if t.Tag == filterTag {
			result = append(result, t)
		}
	}
	return result
}

// OverdueTasks mengembalikan task yang sudah overdue (belum selesai dan deadline lewat)
func (m *Manager) OverdueTasks() []*Task {
	var result []*Task
	for _, t := range m.tasks {
		if t.IsOverdue() {
			result = append(result, t)
		}
	}
	return result
}

// SetTasks mengganti seluruh daftar task (dipakai saat load dari storage)
func (m *Manager) SetTasks(tasks []*Task) {
	m.tasks = tasks
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	m.nextID = maxID + 1
}

// findByID mencari task berdasarkan ID, return nil jika tidak ketemu
func (m *Manager) findByID(id int) *Task {
	for _, t := range m.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}
