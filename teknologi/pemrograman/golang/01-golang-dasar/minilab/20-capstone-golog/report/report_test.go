package report

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"example.com/minilab/20-capstone-golog/analyzer"
	"example.com/minilab/20-capstone-golog/logentry"
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
