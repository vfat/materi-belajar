package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 11: STRING & FORMATTING ===\n")

	fmt.Println("--- Test 1: Basic formatting ---")
	name := "Gopher"
	age := 5
	fmt.Printf("Hello, %s. Age: %d\n", name, age)

	fmt.Println("\n--- Test 2: Sprintf and verbs ---")
	s := fmt.Sprintf("Pi approx: %.2f", 3.14159)
	fmt.Println(s)

	fmt.Println("\n--- Test 3: strconv conversions ---")
	n := 42
	sn := strconv.Itoa(n)
	fmt.Printf("String: %s (len=%d)\n", sn, len(sn))

	fmt.Println("\n--- Test 4: Time formatting ---")
	t := time.Date(2026, 5, 30, 14, 0, 0, 0, time.UTC)
	fmt.Println("RFC3339:", t.Format(time.RFC3339))
	fmt.Println("Custom:", t.Format("02-01-2006 15:04"))

	fmt.Println("\n--- Test 5: Width & padding ---")
	fmt.Printf("|%6d|\n", 42)
	fmt.Printf("|%-6s|\n", "go")

	fmt.Println("\n--- Test 6: Print with %#v / %T ---")
	fmt.Printf("Value: %#v, Type: %T\n", name, name)

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
