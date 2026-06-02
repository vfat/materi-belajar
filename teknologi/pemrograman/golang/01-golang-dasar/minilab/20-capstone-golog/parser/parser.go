package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"example.com/minilab/20-capstone-golog/logentry"
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
