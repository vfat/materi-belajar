package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"example.com/minilab/20-capstone-golog/analyzer"
	"example.com/minilab/20-capstone-golog/logentry"
	"example.com/minilab/20-capstone-golog/parser"
	"example.com/minilab/20-capstone-golog/report"
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
	files       []string
	filterLevel logentry.Level
	keyword     string
	exportPath  string
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
