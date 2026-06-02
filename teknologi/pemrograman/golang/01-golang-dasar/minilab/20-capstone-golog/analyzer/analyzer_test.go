package analyzer

import (
	"testing"
	"time"

	"example.com/minilab/20-capstone-golog/logentry"
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
