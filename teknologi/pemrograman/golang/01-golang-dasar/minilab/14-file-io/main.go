package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 14: FILE I/O ===\n")

	filename := "example.txt"

	fmt.Println("--- Test 1: Write file ---")
	content := []byte("Hello, File I/O\n")
	if err := os.WriteFile(filename, content, 0644); err != nil {
		fmt.Println("write error:", err)
	} else {
		fmt.Println("Wrote to", filename)
	}

	fmt.Println("\n--- Test 2: Read file ---")
	b, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("read error:", err)
	} else {
		fmt.Printf("Content (%d bytes):\n%s", len(b), string(b))
	}

	fmt.Println("\n--- Test 3: Append to file ---")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open for append error:", err)
	} else {
		if _, err := f.WriteString("Appended line\n"); err != nil {
			fmt.Println("append error:", err)
		} else {
			fmt.Println("Appended to", filename)
		}
		f.Close()
	}

	fmt.Println("\n--- Test 4: Read after append ---")
	b2, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("read error:", err)
	} else {
		fmt.Printf("Content (%d bytes):\n%s", len(b2), string(b2))
	}

	fmt.Println("\n--- Test 5: File info (Stat) ---")
	if info, err := os.Stat(filename); err != nil {
		fmt.Println("stat error:", err)
	} else {
		fmt.Printf("Name: %s, Size: %d, Mode: %s\n", info.Name(), info.Size(), info.Mode())
	}

	fmt.Println("\n--- Test 6: Temp file ---")
	tmp, err := os.CreateTemp("", "tmpfile-*.txt")
	if err != nil {
		fmt.Println("create temp error:", err)
	} else {
		if _, err := tmp.WriteString("temporary content\n"); err != nil {
			fmt.Println("write temp error:", err)
		}
		tmp.Close()
		fmt.Println("Temp file:", tmp.Name())
		_ = os.Remove(tmp.Name())
	}

	fmt.Println("\n--- Test 7: Open non-existing ---")
	if _, err := os.ReadFile("this-file-does-not-exist.txt"); err != nil {
		fmt.Println("expected error:", err)
	} else {
		fmt.Println("unexpectedly found file")
	}

	fmt.Println("\n--- Cleanup ---")
	if err := os.Remove(filename); err != nil {
		fmt.Println("remove error:", err)
	} else {
		fmt.Println("Removed", filename)
	}

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
