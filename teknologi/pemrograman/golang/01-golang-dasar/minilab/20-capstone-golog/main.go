package main

import (
	"fmt"
	"os"
	"strings"

	"example.com/minilab/20-capstone-golog/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	app := cli.NewApp()
	output, err := app.Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ Error:", err)
		printUsage()
		os.Exit(1)
	}

	fmt.Print(strings.TrimRight(output, "\n") + "\n")
}

func printUsage() {
	fmt.Print(`
Penggunaan:
  20-test analyze <file.log> [file2.log ...] [opsi]

Opsi:
  --filter  <INFO|WARN|ERROR>   Tampilkan hanya baris dengan level ini
  --keyword <kata>              Tampilkan hanya baris yang mengandung kata ini
  --export  <file.json>         Ekspor ringkasan ke file JSON

Contoh:
  20-test analyze server.log
  20-test analyze server.log --filter ERROR
  20-test analyze *.log --export report.json
`)
}
