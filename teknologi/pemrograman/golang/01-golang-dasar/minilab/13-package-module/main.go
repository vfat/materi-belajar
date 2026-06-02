package main

import (
	"fmt"
	"example.com/minilab/13-package-module/util"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 13: PACKAGE & MODULE ===\n")

	fmt.Println("--- Test 1: Local subpackage usage ---")
	fmt.Println("Add 2 + 3 =", util.Add(2, 3))

	fmt.Println("\n--- Test 2: Greeting from package ---")
	fmt.Println(util.Greeting("Gopher"))

	fmt.Println("\n--- Test 3: go.mod presence ---")
	fmt.Println("Module path: example.com/minilab/13-package-module")

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
