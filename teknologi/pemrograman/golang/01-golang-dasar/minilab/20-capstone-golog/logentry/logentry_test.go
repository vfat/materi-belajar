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
