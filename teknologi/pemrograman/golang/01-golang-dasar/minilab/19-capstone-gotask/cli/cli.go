package cli

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"example.com/minilab/19-capstone-gotask/report"
	"example.com/minilab/19-capstone-gotask/storage"
	"example.com/minilab/19-capstone-gotask/task"
)

// App adalah entry point utama untuk routing CLI
type App struct {
	manager *task.Manager
	store   *storage.Storage
}

// NewApp membuat App baru dengan file penyimpanan yang ditentukan
func NewApp(dataFile string) *App {
	app := &App{
		manager: task.NewManager(),
		store:   storage.NewStorage(dataFile),
	}
	tasks, _ := app.store.Load()
	app.manager.SetTasks(tasks)
	return app
}

// Handle memproses slice argumen CLI dan mengembalikan output string
func (a *App) Handle(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("tidak ada perintah")
	}

	command := args[0]
	rest := args[1:]

	var output string
	var err error

	switch command {
	case "add":
		output, err = a.handleAdd(rest)
	case "list":
		output, err = a.handleList(rest)
	case "done":
		output, err = a.handleDone(rest)
	case "delete":
		output, err = a.handleDelete(rest)
	case "report":
		output, err = a.handleReport(rest)
	default:
		return "", fmt.Errorf("perintah tidak dikenal: %q (coba: add, list, done, delete, report)", command)
	}

	if err != nil {
		return "", err
	}

	if command == "add" || command == "done" || command == "delete" {
		if saveErr := a.store.Save(a.manager.ListTasks("")); saveErr != nil {
			return output, fmt.Errorf("warning: gagal menyimpan: %w", saveErr)
		}
	}

	return output, nil
}

// parseArgs mengekstrak nilai flag sederhana dari args
func parseArgs(args []string) (title string, deadline time.Time, tag string, err error) {
	if len(args) == 0 {
		return "", time.Time{}, "", errors.New("judul task wajib diisi")
	}

	title = args[0]
	deadline = time.Now().Add(7 * 24 * time.Hour)

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--deadline":
			if i+1 >= len(args) {
				return "", time.Time{}, "", errors.New("--deadline butuh nilai tanggal (format: 2006-01-02)")
			}
			i++
			deadline, err = time.Parse("2006-01-02", args[i])
			if err != nil {
				return "", time.Time{}, "", fmt.Errorf("format deadline salah, gunakan: 2006-01-02")
			}
		case "--tag":
			if i+1 < len(args) {
				i++
				tag = args[i]
			}
		}
	}
	return title, deadline, tag, nil
}

func (a *App) handleAdd(args []string) (string, error) {
	title, deadline, tag, err := parseArgs(args)
	if err != nil {
		return "", err
	}
	t, err := a.manager.AddTask(title, deadline, tag)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("✅ Task #%d ditambahkan: %q", t.ID, t.Title), nil
}

func (a *App) handleList(args []string) (string, error) {
	var filterTag string
	for i := 0; i < len(args); i++ {
		if args[i] == "--tag" && i+1 < len(args) {
			filterTag = args[i+1]
			i++
		}
	}

	tasks := a.manager.ListTasks(filterTag)
	if len(tasks) == 0 {
		return "📭 Tidak ada task.", nil
	}
	f := &report.TextFormatter{}
	return f.Format(tasks), nil
}

func (a *App) handleDone(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("gunakan: done <id>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("ID harus angka: %q", args[0])
	}
	if err := a.manager.CompleteTask(id); err != nil {
		return "", err
	}
	return fmt.Sprintf("✅ Task #%d selesai.", id), nil
}

func (a *App) handleDelete(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("gunakan: delete <id>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("ID harus angka: %q", args[0])
	}
	if err := a.manager.DeleteTask(id); err != nil {
		return "", err
	}
	return fmt.Sprintf("🗑️  Task #%d dihapus.", id), nil
}

func (a *App) handleReport(args []string) (string, error) {
	tasks := a.manager.ListTasks("")
	summary := report.BuildSummary(tasks)

	var exportPath string
	for i := 0; i < len(args); i++ {
		if args[i] == "--export" && i+1 < len(args) {
			exportPath = args[i+1]
			i++
		}
	}

	output := fmt.Sprintf(
		"📊 Ringkasan:\n  Total: %d | Selesai: %d | Pending: %d | Overdue: %d\n",
		summary.Total, summary.Done, summary.Pending, summary.Overdue,
	)

	if exportPath != "" {
		f := &report.TextFormatter{}
		if err := report.ExportToFile(f, tasks, exportPath); err != nil {
			return output, fmt.Errorf("gagal ekspor: %w", err)
		}
		output += fmt.Sprintf("📄 Laporan diekspor ke: %s\n", exportPath)
	}

	return output, nil
}
