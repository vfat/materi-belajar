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
