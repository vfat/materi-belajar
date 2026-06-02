package report

import (
	"fmt"
	"os"
	"strings"

	"example.com/minilab/19-capstone-gotask/task"
)

// Formatter adalah interface untuk memformat daftar task
type Formatter interface {
	Format(tasks []*task.Task) string
}

// Summary berisi ringkasan statistik task
type Summary struct {
	Total   int
	Done    int
	Pending int
	Overdue int
}

// TextFormatter memformat task sebagai tabel plain text
type TextFormatter struct{}

// Format menghasilkan tabel teks dari daftar task
func (f *TextFormatter) Format(tasks []*task.Task) string {
	var sb strings.Builder
	header := fmt.Sprintf("%-4s  %-8s  %-12s  %-10s  %s\n",
		"ID", "STATUS", "DEADLINE", "TAG", "JUDUL")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(header)
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for _, t := range tasks {
		line := fmt.Sprintf("%-4d  %-8s  %-12s  %-10s  %s\n",
			t.ID,
			string(t.Status),
			t.Deadline.Format("2006-01-02"),
			t.Tag,
			t.Title,
		)
		sb.WriteString(line)
	}
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	return sb.String()
}

// BuildSummary menghitung statistik dari daftar task
func BuildSummary(tasks []*task.Task) Summary {
	s := Summary{Total: len(tasks)}
	for _, t := range tasks {
		switch t.Status {
		case task.StatusDone:
			s.Done++
		case task.StatusPending:
			s.Pending++
			if t.IsOverdue() {
				s.Overdue++
			}
		}
	}
	return s
}

// ExportToFile menulis hasil format ke file menggunakan Formatter yang diberikan
func ExportToFile(f Formatter, tasks []*task.Task, path string) error {
	content := f.Format(tasks)
	return os.WriteFile(path, []byte(content), 0644)
}
