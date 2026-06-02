package parser

import (
	"strings"
	"testing"

	"example.com/minilab/20-capstone-golog/logentry"
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
