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
