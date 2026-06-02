package checker

import "example.com/minilab/19-capstone-gotask/task"

// ScanOverdue memindai dan mengembalikan task-task yang overdue
func ScanOverdue(tasks []*task.Task) []*task.Task {
	var result []*task.Task
	for _, t := range tasks {
		if t.IsOverdue() {
			result = append(result, t)
		}
	}
	return result
}

// RunCheckerAsync menjalankan pengecekan overdue di goroutine terpisah
// dan mengirimkan hasilnya ke channel yang dikembalikan.
// Channel akan ditutup (close) setelah semua hasil terkirim.
func RunCheckerAsync(tasks []*task.Task) <-chan *task.Task {
	ch := make(chan *task.Task)
	go func() {
		defer close(ch)
		for _, t := range ScanOverdue(tasks) {
			ch <- t
		}
	}()
	return ch
}
