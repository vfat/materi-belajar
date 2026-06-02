---
topik: Capstone Project 2 — golog CLI Log Analyzer
urutan: 20 dari 20
posisi: capstone
sebelumnya: Capstone gotask
---

> 🔗 **Lanjutan dari:** Capstone gotask  
> ← Kembali ke: `19-capstone-gotask.md`

# Capstone Project 2 — `golog` CLI Log Analyzer

## Tujuan Belajar

- Mengulang dan memperdalam **semua 18 materi** dasar Go dari sudut pandang berbeda
- Mempraktikkan TDD pada domain **pemrosesan teks dan file**
- Memahami pola **concurrent fan-out** untuk memproses banyak file sekaligus
- Membangun tool yang relevan untuk dunia nyata (DevOps, Security, Monitoring)

---

## Gambaran Proyek

`golog` adalah tool CLI untuk menganalisis satu atau banyak file log. Tool ini membaca baris-baris log, mendeteksi level (INFO/WARN/ERROR), menghitung statistik, memfilter berdasar level atau kata kunci, dan mengekspor laporan — semua diproses secara concurrent jika ada banyak file.

```
$ golog analyze server.log
──────────────────────────────────────────────────
File     : server.log
Total    : 120 baris
INFO     : 80  (66.7%)
WARN     : 25  (20.8%)
ERROR    : 15  (12.5%)
──────────────────────────────────────────────────

$ golog analyze server.log --filter ERROR
[ERROR] 2026-06-01 10:23:45 - Database connection timeout
[ERROR] 2026-06-01 10:24:01 - Failed to write cache
...

$ golog analyze *.log --export report.json
📄 Laporan diekspor ke: report.json
```

---

## Perbedaan dari Capstone 1

| Aspek              | gotask (19)                   | golog (20)                      |
|--------------------|-------------------------------|---------------------------------|
| Domain             | CRUD / state management       | Parsing teks / analisis data    |
| Concurrency        | Background goroutine + signal | Fan-out: N file diproses paralel|
| Storage            | Baca/tulis JSON (R/W)         | Hanya baca file log (R only)    |
| Interface          | Formatter untuk output        | Parser untuk input log          |
| Fokus testing      | State & persistence           | Parsing & agregasi data         |

---

## Peta Materi — 18 Topik Dipakai Di Mana

| No  | Topik                   | Dipakai Di                                              |
|-----|-------------------------|---------------------------------------------------------|
| 01  | Pengenalan & Instalasi  | `go mod init`, setup project                            |
| 02  | Variabel & Tipe Data    | Struct `LogEntry`, `Stats`, konstanta level             |
| 03  | Operator & Ekspresi     | Hitung persentase, perbandingan count                   |
| 04  | Control Flow            | Routing perintah CLI, switch pada level log             |
| 05  | Array, Slice & Map      | `[]LogEntry`, `map[Level]int` untuk hitungan per level  |
| 06  | Function                | `ParseLine()`, `Analyze()`, `FilterByLevel()`           |
| 07  | Struct                  | Struct `LogEntry`, `Stats`, `FileResult`                |
| 08  | Pointer                 | Pointer receiver pada `Analyzer`, pass stats by pointer |
| 09  | Interface               | Interface `Parser` (plain text vs JSON log)             |
| 10  | Error Handling          | Handle error buka file, parse baris invalid             |
| 11  | String & Formatting     | Parse baris log, format output ke terminal              |
| 12  | Time & Date             | Parse timestamp dari baris log, filter by time range    |
| 13  | Package & Module        | Package `logentry`, `parser`, `analyzer`, `report`, `cli` |
| 14  | File I/O                | Baca file log, ekspor laporan                           |
| 15  | JSON Handling           | Export laporan ke JSON, parse log format JSON           |
| 16  | Goroutine               | Fan-out: setiap file diproses di goroutine terpisah     |
| 17  | Channel                 | Kumpulkan hasil dari semua goroutine via channel        |
| 18  | Unit Testing            | Test parsing, agregasi, filtering, concurrent behavior  |

---

## Konsep Baru: Fan-Out Pattern

Berbeda dengan capstone 1 yang memakai goroutine sebagai background checker, di sini goroutine dipakai untuk **fan-out**: memproses N file secara paralel dan mengumpulkan hasilnya.

```
main()
  │
  ├── go analyzeFile("app.log")   ──► FileResult ──►┐
  ├── go analyzeFile("error.log") ──► FileResult ──►├─► resultCh ──► aggregate
  └── go analyzeFile("web.log")   ──► FileResult ──►┘
```

```go
// Fan-out: kirim setiap file ke goroutine terpisah
resultCh := make(chan FileResult, len(files))
for _, file := range files {
    go func(f string) {
        resultCh <- analyzeFile(f)
    }(file)
}

// Kumpulkan semua hasil
for range files {
    result := <-resultCh
    // proses result...
}
```

---

## Fase 1 — Setup Project

### Langkah 1.1: Buat folder project

```bash
mkdir -p ~/workspace/golog
cd ~/workspace/golog
```

### Langkah 1.2: Inisialisasi module

```bash
go mod init github.com/kamu/golog
```

### Langkah 1.3: Buat struktur folder

```bash
mkdir -p logentry parser analyzer report cli
```

Struktur akhir yang dituju:

```
golog/
├── go.mod
├── main.go
├── logentry/
│   ├── logentry.go       ← Struct LogEntry, Level, konstanta
│   └── logentry_test.go
├── parser/
│   ├── parser.go         ← Interface Parser + implementasi plain text
│   └── parser_test.go
├── analyzer/
│   ├── analyzer.go       ← Hitung statistik, filter, agregasi
│   └── analyzer_test.go
├── report/
│   ├── report.go         ← Format output text & JSON
│   └── report_test.go
└── cli/
    ├── cli.go            ← Routing perintah + fan-out goroutine
    └── cli_test.go
```

---

## Fase 2 — Domain: LogEntry (TDD)

### 🔴 RED — Tulis test dulu

Buat file `logentry/logentry_test.go`:

```go
package logentry

import (
	"testing"
	"time"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelUnknown, "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level.String() = %q; want %q", got, tt.want)
		}
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"INFO", LevelInfo},
		{"info", LevelInfo},
		{"WARN", LevelWarn},
		{"WARNING", LevelWarn},
		{"ERROR", LevelError},
		{"ERR", LevelError},
		{"DEBUG", LevelUnknown},
		{"", LevelUnknown},
	}
	for _, tt := range tests {
		if got := ParseLevel(tt.input); got != tt.want {
			t.Errorf("ParseLevel(%q) = %v; want %v", tt.input, got, tt.want)
		}
	}
}

func TestNewLogEntry(t *testing.T) {
	ts := time.Now()
	entry := NewLogEntry(LevelError, "database timeout", ts)

	if entry.Level != LevelError {
		t.Errorf("Level = %v; want ERROR", entry.Level)
	}
	if entry.Message != "database timeout" {
		t.Errorf("Message = %q; want %q", entry.Message, "database timeout")
	}
	if entry.Timestamp.IsZero() {
		t.Error("Timestamp tidak boleh zero")
	}
}
```

```bash
go test ./logentry/...
# build failed — logentry belum ada
```

### 🟢 GREEN — Implementasi

Buat file `logentry/logentry.go`:

```go
package logentry

import (
	"strings"
	"time"
)

// Level merepresentasikan tingkat keparahan log
type Level int

const (
	LevelUnknown Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String mengubah Level ke representasi string
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel mengubah string menjadi Level
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR", "ERR":
		return LevelError
	default:
		return LevelUnknown
	}
}

// LogEntry merepresentasikan satu baris log yang sudah diparse
type LogEntry struct {
	Level     Level
	Message   string
	Timestamp time.Time
	Raw       string // baris asli sebelum diparse
}

// NewLogEntry membuat LogEntry baru
func NewLogEntry(level Level, message string, ts time.Time) *LogEntry {
	return &LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: ts,
	}
}
```

```bash
go test ./logentry/...
# ok  github.com/kamu/golog/logentry  0.001s
```

### 🔵 REFACTOR

Tidak ada yang perlu diubah. Kode sudah bersih.

---

## Fase 3 — Parser: Interface + Implementasi (TDD)

### 🔴 RED — Tulis test

Buat file `parser/parser_test.go`:

```go
package parser

import (
	"strings"
	"testing"

	"github.com/kamu/golog/logentry"
)

func TestPlainTextParser_Parse(t *testing.T) {
	p := &PlainTextParser{}

	tests := []struct {
		name      string
		line      string
		wantLevel logentry.Level
		wantErr   bool
	}{
		{
			name:      "baris INFO valid",
			line:      "2026-06-01 10:00:00 [INFO] Server started",
			wantLevel: logentry.LevelInfo,
		},
		{
			name:      "baris ERROR valid",
			line:      "2026-06-01 10:05:00 [ERROR] Database connection failed",
			wantLevel: logentry.LevelError,
		},
		{
			name:      "baris WARN valid",
			line:      "2026-06-01 10:03:00 [WARN] Memory usage high",
			wantLevel: logentry.LevelWarn,
		},
		{
			name:    "baris kosong",
			line:    "",
			wantErr: true,
		},
		{
			name:    "baris tanpa format yang dikenal",
			line:    "ini bukan log",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := p.Parse(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) harus error", tt.line)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.line, err)
			}
			if entry.Level != tt.wantLevel {
				t.Errorf("Level = %v; want %v", entry.Level, tt.wantLevel)
			}
		})
	}
}

func TestPlainTextParser_ParseLines(t *testing.T) {
	p := &PlainTextParser{}
	input := `2026-06-01 10:00:00 [INFO] started
2026-06-01 10:01:00 [ERROR] failed
baris tidak valid
2026-06-01 10:02:00 [WARN] warning`

	entries, skipped := p.ParseLines(strings.NewReader(input))

	if len(entries) != 3 {
		t.Errorf("ParseLines() = %d entries; want 3", len(entries))
	}
	if skipped != 1 {
		t.Errorf("skipped = %d; want 1", skipped)
	}
}
```

```bash
go test ./parser/...
# build failed
```

### 🟢 GREEN — Implementasi

Buat file `parser/parser.go`:

```go
package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kamu/golog/logentry"
)

// Parser adalah interface untuk memparse baris log menjadi LogEntry
type Parser interface {
	Parse(line string) (*logentry.LogEntry, error)
}

// PlainTextParser memparse log dalam format:
// "2006-01-02 15:04:05 [LEVEL] message"
type PlainTextParser struct{}

// Parse memparse satu baris log
func (p *PlainTextParser) Parse(line string) (*logentry.LogEntry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, errors.New("baris kosong")
	}

	// Format yang diharapkan: "2006-01-02 15:04:05 [LEVEL] message"
	// Minimal harus ada [LEVEL]
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("format tidak dikenal: %q", line)
	}

	levelStr := line[start+1 : end]
	level := logentry.ParseLevel(levelStr)
	if level == logentry.LevelUnknown {
		return nil, fmt.Errorf("level tidak dikenal: %q", levelStr)
	}

	// Parse timestamp (opsional — jika gagal, pakai zero time)
	var ts time.Time
	if start >= 19 {
		ts, _ = time.Parse("2006-01-02 15:04:05", strings.TrimSpace(line[:start]))
	}

	// Ambil message setelah "] "
	message := ""
	if end+2 < len(line) {
		message = strings.TrimSpace(line[end+1:])
	}

	entry := logentry.NewLogEntry(level, message, ts)
	entry.Raw = line
	return entry, nil
}

// ParseLines membaca semua baris dari reader dan memparse yang valid
// Mengembalikan slice entries dan jumlah baris yang dilewati (tidak bisa diparse)
func (p *PlainTextParser) ParseLines(r io.Reader) ([]*logentry.LogEntry, int) {
	var entries []*logentry.LogEntry
	skipped := 0

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		entry, err := p.Parse(scanner.Text())
		if err != nil {
			skipped++
			continue
		}
		entries = append(entries, entry)
	}
	return entries, skipped
}
```

```bash
go test ./parser/...
# ok  github.com/kamu/golog/parser  0.002s
```

---

## Fase 4 — Analyzer: Statistik & Filter (TDD)

### 🔴 RED — Tulis test

Buat file `analyzer/analyzer_test.go`:

```go
package analyzer

import (
	"testing"
	"time"

	"github.com/kamu/golog/logentry"
)

func makeEntry(level logentry.Level, msg string) *logentry.LogEntry {
	return logentry.NewLogEntry(level, msg, time.Now())
}

func TestAnalyze_Stats(t *testing.T) {
	entries := []*logentry.LogEntry{
		makeEntry(logentry.LevelInfo, "start"),
		makeEntry(logentry.LevelInfo, "running"),
		makeEntry(logentry.LevelWarn, "memory high"),
		makeEntry(logentry.LevelError, "db fail"),
		makeEntry(logentry.LevelError, "timeout"),
	}

	stats := Analyze(entries)

	if stats.Total != 5 {
		t.Errorf("Total = %d; want 5", stats.Total)
	}
	if stats.Count[logentry.LevelInfo] != 2 {
		t.Errorf("INFO count = %d; want 2", stats.Count[logentry.LevelInfo])
	}
	if stats.Count[logentry.LevelWarn] != 1 {
		t.Errorf("WARN count = %d; want 1", stats.Count[logentry.LevelWarn])
	}
	if stats.Count[logentry.LevelError] != 2 {
		t.Errorf("ERROR count = %d; want 2", stats.Count[logentry.LevelError])
	}
}

func TestAnalyze_Percentage(t *testing.T) {
	entries := []*logentry.LogEntry{
		makeEntry(logentry.LevelInfo, "a"),
		makeEntry(logentry.LevelError, "b"),
	}
	stats := Analyze(entries)

	if stats.Percent(logentry.LevelInfo) != 50.0 {
		t.Errorf("INFO percent = %.1f; want 50.0", stats.Percent(logentry.LevelInfo))
	}
}

func TestFilterByLevel(t *testing.T) {
	entries := []*logentry.LogEntry{
		makeEntry(logentry.LevelInfo, "info msg"),
		makeEntry(logentry.LevelError, "error msg"),
		makeEntry(logentry.LevelWarn, "warn msg"),
		makeEntry(logentry.LevelError, "another error"),
	}

	errors := FilterByLevel(entries, logentry.LevelError)
	if len(errors) != 2 {
		t.Errorf("FilterByLevel(ERROR) = %d; want 2", len(errors))
	}
}

func TestFilterByKeyword(t *testing.T) {
	entries := []*logentry.LogEntry{
		makeEntry(logentry.LevelError, "database connection failed"),
		makeEntry(logentry.LevelError, "disk write error"),
		makeEntry(logentry.LevelInfo, "database query ok"),
	}

	results := FilterByKeyword(entries, "database")
	if len(results) != 2 {
		t.Errorf("FilterByKeyword(\"database\") = %d; want 2", len(results))
	}
}

func TestAnalyze_EmptyEntries(t *testing.T) {
	stats := Analyze([]*logentry.LogEntry{})
	if stats.Total != 0 {
		t.Errorf("Total = %d; want 0", stats.Total)
	}
	if stats.Percent(logentry.LevelError) != 0 {
		t.Errorf("Percent pada entries kosong harus 0")
	}
}
```

```bash
go test ./analyzer/...
# build failed
```

### 🟢 GREEN — Implementasi

Buat file `analyzer/analyzer.go`:

```go
package analyzer

import (
	"strings"

	"github.com/kamu/golog/logentry"
)

// Stats menyimpan hasil analisis dari sekumpulan LogEntry
type Stats struct {
	Total int
	Count map[logentry.Level]int
}

// Percent menghitung persentase sebuah level dari total
func (s *Stats) Percent(level logentry.Level) float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Count[level]) / float64(s.Total) * 100
}

// Analyze menghitung statistik dari slice LogEntry
func Analyze(entries []*logentry.LogEntry) *Stats {
	stats := &Stats{
		Total: len(entries),
		Count: make(map[logentry.Level]int),
	}
	for _, e := range entries {
		stats.Count[e.Level]++
	}
	return stats
}

// FilterByLevel mengembalikan entries yang memiliki level tertentu
func FilterByLevel(entries []*logentry.LogEntry, level logentry.Level) []*logentry.LogEntry {
	var result []*logentry.LogEntry
	for _, e := range entries {
		if e.Level == level {
			result = append(result, e)
		}
	}
	return result
}

// FilterByKeyword mengembalikan entries yang mengandung kata kunci di message
func FilterByKeyword(entries []*logentry.LogEntry, keyword string) []*logentry.LogEntry {
	keyword = strings.ToLower(keyword)
	var result []*logentry.LogEntry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Message), keyword) ||
			strings.Contains(strings.ToLower(e.Raw), keyword) {
			result = append(result, e)
		}
	}
	return result
}
```

```bash
go test ./analyzer/...
# ok  github.com/kamu/golog/analyzer  0.002s
```

---

## Fase 5 — Report: Format Output (TDD)

### 🔴 RED — Tulis test

Buat file `report/report_test.go`:

```go
package report

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kamu/golog/analyzer"
	"github.com/kamu/golog/logentry"
)

func makeStats(info, warn, errCount int) *analyzer.Stats {
	total := info + warn + errCount
	return &analyzer.Stats{
		Total: total,
		Count: map[logentry.Level]int{
			logentry.LevelInfo:  info,
			logentry.LevelWarn:  warn,
			logentry.LevelError: errCount,
		},
	}
}

func TestFileResult_TextFormat(t *testing.T) {
	result := &FileResult{
		Filename: "server.log",
		Stats:    makeStats(80, 25, 15),
		Entries: []*logentry.LogEntry{
			logentry.NewLogEntry(logentry.LevelError, "db timeout", time.Now()),
		},
	}

	output := result.TextFormat()

	if !strings.Contains(output, "server.log") {
		t.Error("output tidak mengandung nama file")
	}
	if !strings.Contains(output, "120") {
		t.Error("output tidak mengandung total (120)")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("output tidak mengandung INFO")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("output tidak mengandung ERROR")
	}
}

func TestExportJSON(t *testing.T) {
	results := []*FileResult{
		{
			Filename: "app.log",
			Stats:    makeStats(10, 2, 1),
		},
	}

	tmpFile := t.TempDir() + "/report.json"
	err := ExportJSON(results, tmpFile)
	if err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	data, _ := os.ReadFile(tmpFile)
	var loaded []map[string]interface{}
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("hasil bukan JSON valid: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("len(loaded) = %d; want 1", len(loaded))
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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kamu/golog/analyzer"
	"github.com/kamu/golog/logentry"
)

// FileResult menyimpan hasil analisis satu file log
type FileResult struct {
	Filename string
	Stats    *analyzer.Stats
	Entries  []*logentry.LogEntry // entries yang akan ditampilkan (sudah difilter)
	Err      error                // error jika file gagal dibaca
}

// TextFormat menghasilkan tampilan teks dari hasil analisis satu file
func (r *FileResult) TextFormat() string {
	var sb strings.Builder
	sep := strings.Repeat("─", 50)

	sb.WriteString(sep + "\n")
	if r.Err != nil {
		sb.WriteString(fmt.Sprintf("File     : %s\n", r.Filename))
		sb.WriteString(fmt.Sprintf("Error    : %v\n", r.Err))
		sb.WriteString(sep + "\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("File     : %s\n", r.Filename))
	sb.WriteString(fmt.Sprintf("Total    : %d baris\n", r.Stats.Total))
	sb.WriteString(fmt.Sprintf("INFO     : %-4d (%.1f%%)\n",
		r.Stats.Count[logentry.LevelInfo],
		r.Stats.Percent(logentry.LevelInfo)))
	sb.WriteString(fmt.Sprintf("WARN     : %-4d (%.1f%%)\n",
		r.Stats.Count[logentry.LevelWarn],
		r.Stats.Percent(logentry.LevelWarn)))
	sb.WriteString(fmt.Sprintf("ERROR    : %-4d (%.1f%%)\n",
		r.Stats.Count[logentry.LevelError],
		r.Stats.Percent(logentry.LevelError)))
	sb.WriteString(sep + "\n")

	if len(r.Entries) > 0 {
		sb.WriteString("\nHasil Filter:\n")
		for _, e := range r.Entries {
			sb.WriteString("  " + e.Raw + "\n")
		}
	}

	return sb.String()
}

// jsonFileResult adalah versi JSON-friendly dari FileResult
type jsonFileResult struct {
	Filename string         `json:"filename"`
	Total    int            `json:"total"`
	Info     int            `json:"info"`
	Warn     int            `json:"warn"`
	Error    int            `json:"error"`
	ErrorMsg string         `json:"error_msg,omitempty"`
}

// ExportJSON mengekspor slice FileResult ke file JSON
func ExportJSON(results []*FileResult, path string) error {
	var out []jsonFileResult
	for _, r := range results {
		jr := jsonFileResult{Filename: r.Filename}
		if r.Err != nil {
			jr.ErrorMsg = r.Err.Error()
		} else {
			jr.Total = r.Stats.Total
			jr.Info = r.Stats.Count[logentry.LevelInfo]
			jr.Warn = r.Stats.Count[logentry.LevelWarn]
			jr.Error = r.Stats.Count[logentry.LevelError]
		}
		out = append(out, jr)
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

```bash
go test ./report/...
# ok  github.com/kamu/golog/report  0.002s
```

---

## Fase 6 — CLI: Fan-Out Concurrent (TDD)

### 🔴 RED — Tulis test

Buat file `cli/cli_test.go`:

```go
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeLogFile membuat file log sementara untuk test
func makeLogFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("gagal buat log file: %v", err)
	}
	return path
}

const sampleLog = `2026-06-01 10:00:00 [INFO] Server started
2026-06-01 10:01:00 [INFO] Request received
2026-06-01 10:02:00 [WARN] Memory usage at 80%
2026-06-01 10:03:00 [ERROR] Database timeout
2026-06-01 10:04:00 [ERROR] Retry failed
baris tidak valid
`

func TestApp_HandleAnalyze_SingleFile(t *testing.T) {
	dir := t.TempDir()
	logPath := makeLogFile(t, dir, "test.log", sampleLog)

	app := NewApp()
	output, err := app.Handle([]string{"analyze", logPath})
	if err != nil {
		t.Fatalf("Handle analyze error: %v", err)
	}

	if !strings.Contains(output, "test.log") {
		t.Error("output tidak mengandung nama file")
	}
	if !strings.Contains(output, "5") { // 5 baris valid
		t.Error("output tidak mengandung total baris valid")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("output tidak mengandung ERROR")
	}
}

func TestApp_HandleAnalyze_WithFilterLevel(t *testing.T) {
	dir := t.TempDir()
	logPath := makeLogFile(t, dir, "test.log", sampleLog)

	app := NewApp()
	output, err := app.Handle([]string{"analyze", logPath, "--filter", "ERROR"})
	if err != nil {
		t.Fatalf("Handle analyze --filter ERROR error: %v", err)
	}

	if !strings.Contains(output, "Database timeout") {
		t.Error("output tidak mengandung pesan error yang difilter")
	}
}

func TestApp_HandleAnalyze_WithKeyword(t *testing.T) {
	dir := t.TempDir()
	logPath := makeLogFile(t, dir, "test.log", sampleLog)

	app := NewApp()
	output, err := app.Handle([]string{"analyze", logPath, "--keyword", "database"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(strings.ToLower(output), "database") {
		t.Error("output tidak mengandung keyword 'database'")
	}
}

func TestApp_HandleAnalyze_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	log1 := makeLogFile(t, dir, "app.log", sampleLog)
	log2 := makeLogFile(t, dir, "web.log", `2026-06-01 10:00:00 [INFO] web started
2026-06-01 10:01:00 [ERROR] 404 not found
`)

	app := NewApp()
	output, err := app.Handle([]string{"analyze", log1, log2})
	if err != nil {
		t.Fatalf("Handle multiple files error: %v", err)
	}

	if !strings.Contains(output, "app.log") {
		t.Error("output tidak mengandung app.log")
	}
	if !strings.Contains(output, "web.log") {
		t.Error("output tidak mengandung web.log")
	}
}

func TestApp_HandleAnalyze_FileNotFound(t *testing.T) {
	app := NewApp()
	output, err := app.Handle([]string{"analyze", "/tmp/tidakada.log"})
	// Tidak return error — file yang gagal tetap dilaporkan di output
	if err != nil {
		t.Fatalf("harus tidak error jika file tidak ada: %v", err)
	}
	if !strings.Contains(output, "tidakada.log") {
		t.Error("output harus menyebut nama file yang gagal")
	}
}

func TestApp_HandleAnalyze_ExportJSON(t *testing.T) {
	dir := t.TempDir()
	logPath := makeLogFile(t, dir, "test.log", sampleLog)
	exportPath := filepath.Join(dir, "report.json")

	app := NewApp()
	_, err := app.Handle([]string{"analyze", logPath, "--export", exportPath})
	if err != nil {
		t.Fatalf("Handle --export error: %v", err)
	}

	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("file report.json tidak terbentuk")
	}
}

func TestApp_HandlePerintahTidakDikenal(t *testing.T) {
	app := NewApp()
	_, err := app.Handle([]string{"tidakada"})
	if err == nil {
		t.Error("harus error untuk perintah tidak dikenal")
	}
}

func TestApp_HandleAnalyze_TanpaFile(t *testing.T) {
	app := NewApp()
	_, err := app.Handle([]string{"analyze"})
	if err == nil {
		t.Error("harus error jika tidak ada file yang diberikan")
	}
}
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
	"os"
	"strings"

	"github.com/kamu/golog/analyzer"
	"github.com/kamu/golog/logentry"
	"github.com/kamu/golog/parser"
	"github.com/kamu/golog/report"
)

// App adalah entry point utama untuk routing CLI
type App struct{}

// NewApp membuat App baru
func NewApp() *App {
	return &App{}
}

// Handle memproses argumen CLI
func (a *App) Handle(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("tidak ada perintah")
	}

	switch args[0] {
	case "analyze":
		return a.handleAnalyze(args[1:])
	default:
		return "", fmt.Errorf("perintah tidak dikenal: %q (coba: analyze)", args[0])
	}
}

// options menyimpan flag hasil parsing argumen
type options struct {
	files      []string
	filterLevel logentry.Level
	keyword    string
	exportPath string
}

// parseOptions memisahkan file paths dan flag dari args
func parseOptions(args []string) (options, error) {
	var opts options
	opts.filterLevel = logentry.LevelUnknown

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--filter":
			if i+1 >= len(args) {
				return opts, errors.New("--filter butuh nilai level (INFO/WARN/ERROR)")
			}
			i++
			opts.filterLevel = logentry.ParseLevel(args[i])
			if opts.filterLevel == logentry.LevelUnknown {
				return opts, fmt.Errorf("level tidak dikenal: %q", args[i])
			}
		case "--keyword":
			if i+1 >= len(args) {
				return opts, errors.New("--keyword butuh nilai")
			}
			i++
			opts.keyword = args[i]
		case "--export":
			if i+1 >= len(args) {
				return opts, errors.New("--export butuh path file")
			}
			i++
			opts.exportPath = args[i]
		default:
			// Bukan flag — anggap sebagai path file
			if !strings.HasPrefix(args[i], "--") {
				opts.files = append(opts.files, args[i])
			}
		}
		i++
	}
	return opts, nil
}

// analyzeFile memproses satu file log dan mengembalikan FileResult
// Fungsi ini dipanggil dari goroutine.
func analyzeFile(path string, opts options) *report.FileResult {
	f, err := os.Open(path)
	if err != nil {
		return &report.FileResult{Filename: path, Err: err}
	}
	defer f.Close()

	p := &parser.PlainTextParser{}
	entries, _ := p.ParseLines(f)
	stats := analyzer.Analyze(entries)

	// Terapkan filter jika ada
	var filtered []*logentry.LogEntry
	if opts.filterLevel != logentry.LevelUnknown {
		filtered = analyzer.FilterByLevel(entries, opts.filterLevel)
	} else if opts.keyword != "" {
		filtered = analyzer.FilterByKeyword(entries, opts.keyword)
	}

	return &report.FileResult{
		Filename: path,
		Stats:    stats,
		Entries:  filtered,
	}
}

func (a *App) handleAnalyze(args []string) (string, error) {
	opts, err := parseOptions(args)
	if err != nil {
		return "", err
	}
	if len(opts.files) == 0 {
		return "", errors.New("berikan minimal satu file log. Contoh: golog analyze app.log")
	}

	// Fan-out: proses setiap file di goroutine terpisah
	resultCh := make(chan *report.FileResult, len(opts.files))
	for _, file := range opts.files {
		go func(path string) {
			resultCh <- analyzeFile(path, opts)
		}(file)
	}

	// Kumpulkan semua hasil — urutan sesuai kedatangan (concurrent)
	results := make([]*report.FileResult, len(opts.files))
	for i := range opts.files {
		results[i] = <-resultCh
	}

	// Bangun output teks
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(r.TextFormat())
	}

	// Export JSON jika diminta
	if opts.exportPath != "" {
		if err := report.ExportJSON(results, opts.exportPath); err != nil {
			return sb.String(), fmt.Errorf("gagal ekspor: %w", err)
		}
		sb.WriteString(fmt.Sprintf("\n📄 Laporan diekspor ke: %s\n", opts.exportPath))
	}

	return sb.String(), nil
}
```

```bash
go test ./cli/...
# ok  github.com/kamu/golog/cli  0.005s
```

---

## Fase 7 — main.go: Entry Point

### Buat file `main.go`

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kamu/golog/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	app := cli.NewApp()
	output, err := app.Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error:", err)
		printUsage()
		os.Exit(1)
	}

	fmt.Print(strings.TrimRight(output, "\n") + "\n")
}

func printUsage() {
	fmt.Print(`
Penggunaan:
  golog analyze <file.log> [file2.log ...] [opsi]

Opsi:
  --filter  <INFO|WARN|ERROR>   Tampilkan hanya baris dengan level ini
  --keyword <kata>              Tampilkan hanya baris yang mengandung kata ini
  --export  <file.json>         Ekspor ringkasan ke file JSON

Contoh:
  golog analyze server.log
  golog analyze server.log --filter ERROR
  golog analyze *.log --export report.json
`)
}
```

---

## Fase 8 — Verifikasi Semua Test & Build

### Langkah 8.1: Jalankan seluruh test

```bash
go test ./...
```

Output yang diharapkan:
```
ok  github.com/kamu/golog/analyzer  0.002s
ok  github.com/kamu/golog/cli       0.005s
ok  github.com/kamu/golog/logentry  0.001s
ok  github.com/kamu/golog/parser    0.002s
ok  github.com/kamu/golog/report    0.002s
```

### Langkah 8.2: Jalankan dengan race detector

```bash
go test -race ./...
```

### Langkah 8.3: Lihat coverage

```bash
go test -cover ./...
```

### Langkah 8.4: Build binary

```bash
go build -o golog .
```

### Langkah 8.5: Buat sample log dan coba

Buat file `sample.log`:

```
2026-06-01 10:00:00 [INFO] Server started on port 8080
2026-06-01 10:01:00 [INFO] Connected to database
2026-06-01 10:02:00 [WARN] Memory usage at 75%
2026-06-01 10:03:00 [INFO] Request GET /api/users - 200 OK
2026-06-01 10:04:00 [ERROR] Database connection timeout after 30s
2026-06-01 10:05:00 [ERROR] Retry 1/3 failed
2026-06-01 10:06:00 [WARN] High CPU usage detected
2026-06-01 10:07:00 [ERROR] Service unavailable: cache miss
2026-06-01 10:08:00 [INFO] Health check passed
```

```bash
# Analisis dasar
./golog analyze sample.log

# Filter hanya ERROR
./golog analyze sample.log --filter ERROR

# Cari keyword
./golog analyze sample.log --keyword database

# Ekspor ke JSON
./golog analyze sample.log --export report.json
cat report.json

# Analisis banyak file sekaligus (concurrent)
./golog analyze sample.log sample.log --export multi-report.json
```

Contoh output:
```
──────────────────────────────────────────────────
File     : sample.log
Total    : 9 baris
INFO     : 4   (44.4%)
WARN     : 2   (22.2%)
ERROR    : 3   (33.3%)
──────────────────────────────────────────────────
```

### Langkah 8.6: Install ke PATH (opsional)

```bash
go install .
golog analyze /var/log/syslog --filter ERROR
```

---

## Manual Test

Semua konsep dalam materi ini sudah diimplementasi di folder praktek:

📁 **Lokasi:** `minilab/go/20-capstone-golog/`

### Struktur Folder

```
minilab/go/20-capstone-golog/
├── 20-test              ← binary hasil build, langsung bisa dijalankan
├── go.mod
├── main.go
├── sample.log           ← file log contoh untuk tes manual
├── logentry/
│   ├── logentry.go
│   └── logentry_test.go
├── parser/
│   ├── parser.go
│   └── parser_test.go
├── analyzer/
│   ├── analyzer.go
│   └── analyzer_test.go
├── report/
│   ├── report.go
│   └── report_test.go
└── cli/
    ├── cli.go
    └── cli_test.go
```

### Cara Menjalankan Test

```bash
cd minilab/go/20-capstone-golog

# Semua test sekaligus
go test ./...

# Verbose
go test -v ./...

# Dengan race detector
go test -race ./...

# Coverage per package
go test -cover ./...
```

### Cara Build Ulang & Jalankan Binary

```bash
cd minilab/go/20-capstone-golog

# Build ulang
go build -o 20-test .

# Analisis dasar
./20-test analyze sample.log

# Filter level
./20-test analyze sample.log --filter ERROR

# Cari keyword
./20-test analyze sample.log --keyword database

# Ekspor ke JSON
./20-test analyze sample.log --export laporan.json
cat laporan.json
```

---

## 🏁 Selesai!

Selamat! Kamu telah menyelesaikan seluruh rangkaian **Belajar Golang Dasar** — 18 topik inti ditambah 2 capstone project.

📋 Lihat ringkasan perjalanan belajarmu di: `context.md`

### Langkah Selanjutnya

🎓 **Lanjutkan dengan:**
- Go intermediate: Context, sync.Mutex, WaitGroup lanjutan
- Web development dengan Go (`net/http`, Gin, Echo)
- Database programming (SQL, GORM, sqlx)
- REST API development
- Deployment: Docker, binary distribution

### Referensi

- https://go.dev/doc/
- https://quii.gitbook.io/learn-go-with-tests/
- https://gobyexample.com/
