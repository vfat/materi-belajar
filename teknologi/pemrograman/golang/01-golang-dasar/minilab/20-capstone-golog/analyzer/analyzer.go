package analyzer

import (
	"strings"

	"example.com/minilab/20-capstone-golog/logentry"
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
