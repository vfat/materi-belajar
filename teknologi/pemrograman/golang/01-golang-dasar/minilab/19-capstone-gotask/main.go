package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"example.com/minilab/19-capstone-gotask/checker"
	"example.com/minilab/19-capstone-gotask/cli"
	"example.com/minilab/19-capstone-gotask/storage"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: tidak bisa mendapatkan home directory:", err)
		os.Exit(1)
	}
	dataFile := filepath.Join(homeDir, ".gotask", "tasks.json")

	store := storage.NewStorage(dataFile)
	tasks, _ := store.Load()
	overdueCh := checker.RunCheckerAsync(tasks)

	var overdueWarnings []string
	for t := range overdueCh {
		overdueWarnings = append(overdueWarnings,
			fmt.Sprintf("  ⚠️  Task #%d overdue: %q", t.ID, t.Title))
	}
	if len(overdueWarnings) > 0 {
		fmt.Println("── Peringatan Overdue ──────────────────────────")
		for _, w := range overdueWarnings {
			fmt.Println(w)
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	app := cli.NewApp(dataFile)
	output, err := app.Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error:", err)
		printUsage()
		os.Exit(1)
	}

	fmt.Println(strings.TrimRight(output, "\n"))
}

func printUsage() {
	fmt.Print(`
Penggunaan:
  gotask add <judul> [--deadline YYYY-MM-DD] [--tag <tag>]
  gotask list [--tag <tag>]
  gotask done <id>
  gotask delete <id>
  gotask report [--export <file.txt>]
`)
}
