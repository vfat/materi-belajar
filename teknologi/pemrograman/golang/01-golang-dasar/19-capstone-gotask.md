---
topik: Capstone Project — gotask CLI Task Manager
urutan: 19 dari 20
posisi: capstone
sebelumnya: Unit Testing
---

> 🔗 **Lanjutan dari:** Unit Testing  
> ← Kembali ke: `18-unit-testing.md`

# Capstone Project — `gotask` CLI Task Manager

## Tujuan Belajar

- Menerapkan **semua 18 materi** dasar Go dalam satu proyek nyata
- Memahami dan mempraktikkan **Test-Driven Development (TDD)**
- Membangun aplikasi CLI lengkap dari nol sampai bisa di-build
- Memahami arsitektur multi-package dalam proyek Go nyata
- Merasakan siklus pengembangan: setup → test → implement → refactor → build

---

## Gambaran Proyek

`gotask` adalah aplikasi command-line untuk mengelola tugas pribadi. Data disimpan secara persisten di file JSON, ada pengecekan deadline otomatis di background, dan laporan bisa diekspor ke file teks.

```
$ gotask add "Belajar Go interfaces" --deadline 2026-06-05 --tag belajar
✅ Task #1 ditambahkan.

$ gotask list
ID  STATUS   DEADLINE    TAG      JUDUL
─────────────────────────────────────────────────────
 1  pending  2026-06-05  belajar  Belajar Go interfaces

$ gotask done 1
✅ Task #1 selesai.

$ gotask report --export laporan.txt
📄 Laporan diekspor ke laporan.txt
```

---

## Peta Materi — 18 Topik Dipakai Di Mana

| No  | Topik                   | Dipakai Di                                          |
|-----|-------------------------|-----------------------------------------------------|
| 01  | Pengenalan & Instalasi  | `go mod init`, setup project                        |
| 02  | Variabel & Tipe Data    | Field `Task`: ID, Title, Status, Deadline           |
| 03  | Operator & Ekspresi     | Hitung sisa hari, perbandingan deadline             |
| 04  | Control Flow            | Routing perintah CLI (`add`, `done`, `list`, dll.)  |
| 05  | Array, Slice & Map      | `[]Task` untuk daftar tugas, `map` per tag          |
| 06  | Function                | `AddTask()`, `CompleteTask()`, `ListTasks()`        |
| 07  | Struct                  | Struct `Task`, `Report`, `OverdueResult`            |
| 08  | Pointer                 | Pointer receiver saat update task                   |
| 09  | Interface               | Interface `Formatter` (plain text vs JSON)          |
| 10  | Error Handling          | Validasi input, error baca/tulis file               |
| 11  | String & Formatting     | Format tampilan task ke terminal                    |
| 12  | Time & Date             | Parse deadline, hitung overdue, timestamp           |
| 13  | Package & Module        | Package `task`, `storage`, `report`, `checker`, `cli` |
| 14  | File I/O                | Baca/tulis `tasks.json`, ekspor `report.txt`        |
| 15  | JSON Handling           | Marshal/Unmarshal data task                         |
| 16  | Goroutine               | Background overdue checker saat startup             |
| 17  | Channel                 | Kirim hasil overdue ke main goroutine               |
| 18  | Unit Testing            | Test setiap fungsi dan package                      |

---

## Konsep TDD (Test-Driven Development)

TDD berarti **tulis test dulu** sebelum menulis implementasi. Siklus TDD disebut **Red → Green → Refactor**:

```
    ┌─────────────────────────────────────┐
    │                                     │
    ▼                                     │
  🔴 RED          Tulis test → jalankan → GAGAL
    │                                     │
    ▼                                     │
  🟢 GREEN        Tulis kode minimal → test LULUS
    │                                     │
    ▼                                     │
  🔵 REFACTOR     Perbaiki kode → test tetap LULUS
    │                                     │
    └─────────────────────────────────────┘
```

**Mengapa TDD?**
- Memastikan setiap fitur punya test sebelum selesai
- Mencegah regresi (kode baru merusak kode lama)
- Mendorong desain kode yang mudah di-test (dan otomatis lebih bersih)
- Dokumentasi hidup: test menjelaskan perilaku yang diharapkan

---

## Fase 1 — Setup Project

### Langkah 1.1: Buat folder project

```bash
mkdir -p ~/workspace/gotask
cd ~/workspace/gotask
```

### Langkah 1.2: Inisialisasi module

```bash
go mod init github.com/kamu/gotask
```

File `go.mod` akan terbentuk:

```
module github.com/kamu/gotask

go 1.21
```

### Langkah 1.3: Buat struktur folder

```bash
mkdir -p task storage report checker cli
```

Struktur akhir yang dituju:

```
gotask/
├── go.mod
├── main.go
├── cli/
│   ├── cli.go
│   └── cli_test.go
├── task/
│   ├── task.go         ← Struct & interface
│   ├── task_test.go
│   ├── manager.go      ← CRUD logic
│   └── manager_test.go
├── storage/
│   ├── storage.go      ← Baca/tulis JSON
│   └── storage_test.go
├── report/
│   ├── report.go       ← Format & ekspor laporan
│   └── report_test.go
└── checker/
    ├── checker.go      ← Goroutine + channel overdue
    └── checker_test.go
```

---

## Fase 2 — Domain: Struct Task (TDD)

### 🔴 RED — Tulis test dulu

Buat file `task/task_test.go`:

```go
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
```

Jalankan test — **harus GAGAL** karena kode belum ada:

```bash
go test ./task/...
# task [build failed]
# ./task/task_test.go:8:11: undefined: NewTask
```

### 🟢 GREEN — Implementasi minimal

Buat file `task/task.go`:

```go
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
```

Jalankan test lagi — **harus LULUS**:

```bash
go test ./task/...
# ok  github.com/kamu/gotask/task  0.001s
```

### 🔵 REFACTOR

Tidak ada yang perlu direfaktor sekarang. Kode sudah bersih.

---

## Fase 3 — Manager: CRUD Task (TDD)

### 🔴 RED — Tulis test

Buat file `task/manager_test.go`:

```go
package task

import (
	"testing"
	"time"
)

func newDeadline(hoursFromNow float64) time.Time {
	return time.Now().Add(time.Duration(hoursFromNow * float64(time.Hour)))
}

func TestManager_AddTask(t *testing.T) {
	m := NewManager()
	deadline := newDeadline(48)

	task, err := m.AddTask("Belajar Struct", deadline, "go")
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}
	if task.ID != 1 {
		t.Errorf("ID pertama = %d; want 1", task.ID)
	}

	// Tambah kedua
	task2, _ := m.AddTask("Belajar Interface", deadline, "go")
	if task2.ID != 2 {
		t.Errorf("ID kedua = %d; want 2", task2.ID)
	}
}

func TestManager_AddTask_TitleKosong(t *testing.T) {
	m := NewManager()
	_, err := m.AddTask("", time.Now().Add(time.Hour), "")
	if err == nil {
		t.Error("harus error jika title kosong")
	}
}

func TestManager_CompleteTask(t *testing.T) {
	m := NewManager()
	task, _ := m.AddTask("Test", newDeadline(24), "")

	err := m.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("CompleteTask error: %v", err)
	}

	tasks := m.ListTasks("")
	if tasks[0].Status != StatusDone {
		t.Errorf("Status = %q; want done", tasks[0].Status)
	}
}

func TestManager_CompleteTask_TidakAda(t *testing.T) {
	m := NewManager()
	err := m.CompleteTask(999)
	if err == nil {
		t.Error("harus error jika ID tidak ada")
	}
}

func TestManager_ListTasks(t *testing.T) {
	m := NewManager()
	m.AddTask("Task 1", newDeadline(24), "go")
	m.AddTask("Task 2", newDeadline(48), "belajar")
	m.AddTask("Task 3", newDeadline(72), "go")

	// List semua
	all := m.ListTasks("")
	if len(all) != 3 {
		t.Errorf("len(ListTasks(\"\")) = %d; want 3", len(all))
	}

	// Filter berdasar tag
	goTasks := m.ListTasks("go")
	if len(goTasks) != 2 {
		t.Errorf("len(ListTasks(\"go\")) = %d; want 2", len(goTasks))
	}
}

func TestManager_DeleteTask(t *testing.T) {
	m := NewManager()
	task, _ := m.AddTask("Hapus ini", newDeadline(24), "")

	err := m.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("DeleteTask error: %v", err)
	}

	all := m.ListTasks("")
	if len(all) != 0 {
		t.Errorf("task seharusnya sudah dihapus, len = %d", len(all))
	}
}

func TestManager_OverdueTasks(t *testing.T) {
	m := NewManager()
	m.AddTask("Sudah lewat", newDeadline(-24), "")    // overdue
	m.AddTask("Belum lewat", newDeadline(24), "")     // ok
	m.AddTask("Sudah selesai", newDeadline(-24), "")  // overdue tapi done

	// selesaikan task ke-3
	tasks := m.ListTasks("")
	m.CompleteTask(tasks[2].ID)

	overdue := m.OverdueTasks()
	if len(overdue) != 1 {
		t.Errorf("OverdueTasks() = %d; want 1", len(overdue))
	}
}
```

```bash
go test ./task/...
# build failed — Manager belum ada
```

### 🟢 GREEN — Implementasi

Buat file `task/manager.go`:

```go
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
	// Update nextID agar tidak bentrok
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
```

```bash
go test ./task/...
# ok  github.com/kamu/gotask/task  0.002s
```

---

## Fase 4 — Storage: JSON (TDD)

### 🔴 RED — Tulis test

Buat file `storage/storage_test.go`:

```go
package storage

import (
	"os"
	"testing"
	"time"

	"github.com/kamu/gotask/task"
)

func TestStorage_SaveLoad(t *testing.T) {
	// Gunakan file sementara untuk test
	tmpFile := t.TempDir() + "/tasks_test.json"
	store := NewStorage(tmpFile)

	deadline := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	tasks := []*task.Task{
		{ID: 1, Title: "Test Task", Status: task.StatusPending, Deadline: deadline, Tag: "test"},
	}

	// Simpan
	err := store.Save(tasks)
	if err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Pastikan file terbentuk
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("file tidak terbentuk setelah Save")
	}

	// Load ulang
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

	// File tidak ada → return slice kosong, bukan error
	tasks, err := store.Load()
	if err != nil {
		t.Fatalf("Load file tidak ada harus return kosong, bukan error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Load() = %d; want 0", len(tasks))
	}
}
```

```bash
go test ./storage/...
# build failed
```

### 🟢 GREEN — Implementasi

Buat file `storage/storage.go`:

```go
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/kamu/gotask/task"
)

// Storage menangani baca/tulis task ke file JSON
type Storage struct {
	path string
}

// NewStorage membuat Storage dengan path file yang ditentukan
func NewStorage(path string) *Storage {
	return &Storage{path: path}
}

// Save menyimpan slice task ke file JSON
func (s *Storage) Save(tasks []*task.Task) error {
	// Pastikan direktori ada
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// Load membaca task dari file JSON
// Jika file tidak ada, return slice kosong (bukan error)
func (s *Storage) Load() ([]*task.Task, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*task.Task{}, nil
		}
		return nil, err
	}

	var tasks []*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
```

```bash
go test ./storage/...
# ok  github.com/kamu/gotask/storage  0.003s
```

---

## Fase 5 — Formatter Interface (TDD)

### 🔴 RED — Tulis test

Buat file `report/report_test.go`:

```go
package report

import (
	"strings"
	"testing"
	"time"

	"github.com/kamu/gotask/task"
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

	// Header harus ada
	if !strings.Contains(output, "ID") {
		t.Error("output tidak mengandung header ID")
	}
	// Judul task harus ada
	if !strings.Contains(output, "Belajar Go") {
		t.Error("output tidak mengandung judul task")
	}
	// Status harus ada
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
```

```bash
go test ./report/...
# build failed
```

### 🟢 GREEN — Implementasi

Buat file `report/report.go`:

```go
package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/kamu/gotask/task"
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
```

```bash
go test ./report/...
# ok  github.com/kamu/gotask/report  0.002s
```

---

## Fase 6 — Checker: Goroutine + Channel (TDD)

### 🔴 RED — Tulis test

Buat file `checker/checker_test.go`:

```go
package checker

import (
	"testing"
	"time"

	"github.com/kamu/gotask/task"
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

	// RunCheckerAsync harus mengirim hasilnya via channel dan menutup channel saat selesai
	ch := RunCheckerAsync(tasks)

	var received []*task.Task
	for t := range ch {
		received = append(received, t)
	}

	if len(received) != 1 {
		t.Errorf("RunCheckerAsync hasil = %d; want 1", len(received))
	}
}
```

```bash
go test ./checker/...
# build failed
```

### 🟢 GREEN — Implementasi

Buat file `checker/checker.go`:

```go
package checker

import "github.com/kamu/gotask/task"

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
```

```bash
go test ./checker/...
# ok  github.com/kamu/gotask/checker  0.002s
```

---

## Fase 7 — CLI: Routing Perintah (TDD)

### 🔴 RED — Tulis test

Buat file `cli/cli_test.go`:

```go
package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/kamu/gotask/task"
)

func newTestApp(tmpDir string) *App {
	dataFile := tmpDir + "/tasks.json"
	return NewApp(dataFile)
}

func TestApp_HandleAdd(t *testing.T) {
	app := newTestApp(t.TempDir())

	// Simulasi: gotask add "Judul task" --deadline 2026-06-10 --tag belajar
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

	// Tambah dulu
	deadline := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	app.Handle([]string{"add", "Task satu", "--deadline", deadline, "--tag", "go"})
	app.Handle([]string{"add", "Task dua", "--deadline", deadline, "--tag", "belajar"})

	// List semua
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

// Test untuk memastikan data persisten (disimpan dan dimuat ulang)
func TestApp_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	deadline := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	// App pertama: tambah task
	app1 := newTestApp(tmpDir)
	app1.Handle([]string{"add", "Persisten task", "--deadline", deadline, "--tag", "test"})

	// App kedua: baca dari file yang sama
	app2 := newTestApp(tmpDir)
	output, err := app2.Handle([]string{"list"})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if !strings.Contains(output, "Persisten task") {
		t.Errorf("task tidak persisten, output: %q", output)
	}
}

// Pastikan interface Formatter terpenuhi
var _ interface{ Format(tasks []*task.Task) string } = nil
```

```bash
go test ./cli/...
# build failed
```

### 🟢 GREEN — Implementasi

Buat file `cli/cli.go`:

```go
package cli

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/kamu/gotask/report"
	"github.com/kamu/gotask/storage"
	"github.com/kamu/gotask/task"
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
	// Muat data dari file jika ada
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

	// Simpan setelah setiap operasi yang mengubah data
	if command == "add" || command == "done" || command == "delete" {
		if saveErr := a.store.Save(a.manager.ListTasks("")); saveErr != nil {
			return output, fmt.Errorf("warning: gagal menyimpan: %w", saveErr)
		}
	}

	return output, nil
}

// parseArgs mengekstrak nilai flag sederhana dari args
// Contoh: ["Judul", "--deadline", "2026-06-05", "--tag", "go"]
func parseArgs(args []string) (title string, deadline time.Time, tag string, err error) {
	if len(args) == 0 {
		return "", time.Time{}, "", errors.New("judul task wajib diisi")
	}

	title = args[0]
	deadline = time.Now().Add(7 * 24 * time.Hour) // default: 7 hari ke depan

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
```

```bash
go test ./cli/...
# ok  github.com/kamu/gotask/cli  0.004s
```

---

## Fase 8 — main.go: Entry Point + Overdue Checker

### Buat file `main.go`

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kamu/gotask/checker"
	"github.com/kamu/gotask/cli"
	"github.com/kamu/gotask/storage"
)

func main() {
	// Tentukan lokasi file data: ~/.gotask/tasks.json
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: tidak bisa mendapatkan home directory:", err)
		os.Exit(1)
	}
	dataFile := filepath.Join(homeDir, ".gotask", "tasks.json")

	// Jalankan overdue checker secara async sebelum proses perintah
	store := storage.NewStorage(dataFile)
	tasks, _ := store.Load()
	overdueCh := checker.RunCheckerAsync(tasks)

	// Terima dan tampilkan peringatan overdue dari goroutine
	var overdueWarnings []string
	for t := range overdueCh {
		overdueWarnings = append(overdueWarnings,
			fmt.Sprintf("  ⚠️  Task #%d overdue: %q", t.ID, t.Title))
	}
	if len(overdueWarnings) > 0 {
		fmt.Println("── Peringatan Overdue ──────────────────────────")
		for _, w := range overdueWarnings {
			fmt.Println(w)
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	// Proses argumen CLI
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	app := cli.NewApp(dataFile)
	output, err := app.Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error:", err)
		printUsage()
		os.Exit(1)
	}

	// Trim trailing whitespace sebelum cetak
	fmt.Println(strings.TrimRight(output, "\n"))
}

func printUsage() {
	fmt.Println(`
Penggunaan:
  gotask add <judul> [--deadline YYYY-MM-DD] [--tag <tag>]
  gotask list [--tag <tag>]
  gotask done <id>
  gotask delete <id>
  gotask report [--export <file.txt>]
`)
}
```

---

## Fase 9 — Verifikasi Semua Test & Build

### Langkah 9.1: Jalankan seluruh test

```bash
go test ./...
```

Output yang diharapkan:
```
ok  github.com/kamu/gotask/checker  0.002s
ok  github.com/kamu/gotask/cli      0.005s
ok  github.com/kamu/gotask/report   0.002s
ok  github.com/kamu/gotask/storage  0.003s
ok  github.com/kamu/gotask/task     0.002s
```

### Langkah 9.2: Jalankan dengan race detector

```bash
go test -race ./...
```

### Langkah 9.3: Lihat coverage

```bash
go test -cover ./...
```

### Langkah 9.4: Build binary

```bash
go build -o gotask .
```

### Langkah 9.5: Coba jalankan

```bash
# Tambah beberapa task
./gotask add "Selesaikan capstone Go" --deadline 2026-06-10 --tag belajar
./gotask add "Buat README project" --deadline 2026-06-07 --tag docs
./gotask add "Review kode" --deadline 2026-05-29 --tag review  # sengaja overdue

# Lihat semua task
./gotask list

# Selesaikan task 1
./gotask done 1

# Lihat laporan
./gotask report

# Ekspor laporan
./gotask report --export laporan.txt
cat laporan.txt
```

Contoh output saat ada overdue:
```
── Peringatan Overdue ──────────────────────────
  ⚠️  Task #3 overdue: "Review kode"
────────────────────────────────────────────────
📊 Ringkasan:
  Total: 3 | Selesai: 1 | Pending: 2 | Overdue: 1
```

### Langkah 9.6: Install ke PATH (opsional)

```bash
go install .
# Sekarang bisa dipakai dari mana saja:
gotask list
```

---

## Manual Test

Semua konsep dalam materi ini sudah diimplementasi di folder praktek:

📁 **Lokasi:** `minilab/go/19-capstone-gotask/`

### Struktur Folder

```
minilab/go/19-capstone-gotask/
├── 19-test              ← binary hasil build, langsung bisa dijalankan
├── go.mod
├── main.go
├── task/
│   ├── task.go
│   ├── task_test.go
│   ├── manager.go
│   └── manager_test.go
├── storage/
│   ├── storage.go
│   └── storage_test.go
├── report/
│   ├── report.go
│   └── report_test.go
├── checker/
│   ├── checker.go
│   └── checker_test.go
└── cli/
    ├── cli.go
    └── cli_test.go
```

### Cara Menjalankan Test

```bash
cd minilab/go/19-capstone-gotask

# Semua test sekaligus
go test ./...

# Verbose (lihat nama test satu per satu)
go test -v ./...

# Dengan race detector
go test -race ./...

# Coverage per package
go test -cover ./...

# Coverage HTML report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
```

### Cara Build Ulang & Jalankan Binary

```bash
cd minilab/go/19-capstone-gotask

# Build ulang binary
go build -o 19-test .

# Jalankan langsung
./19-test add "Selesaikan capstone Go" --deadline 2026-06-10 --tag belajar
./19-test add "Buat README project" --deadline 2026-06-07 --tag docs
./19-test add "Review kode" --deadline 2026-05-29 --tag review

./19-test list
./19-test done 1
./19-test report
./19-test report --export laporan.txt
cat laporan.txt
```

---

## ➡️ Selanjutnya

**[Capstone Project 2 — Mini Log Analyzer]**  
→ Lanjut ke: `20-capstone-loganalyzer.md`
