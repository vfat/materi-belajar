package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"example.com/minilab/20-capstone-golog/analyzer"
	"example.com/minilab/20-capstone-golog/logentry"
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
	Filename string `json:"filename"`
	Total    int    `json:"total"`
	Info     int    `json:"info"`
	Warn     int    `json:"warn"`
	Error    int    `json:"error"`
	ErrorMsg string `json:"error_msg,omitempty"`
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
