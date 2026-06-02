package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 04: CONTROL FLOW ===\n")

	// Test 1: If/Else
	fmt.Println("--- Test 1: If/Else ---")
	nilai := 85
	if nilai >= 90 {
		fmt.Println("Grade: A")
	} else if nilai >= 80 {
		fmt.Println("Grade: B")
	} else {
		fmt.Println("Grade: C")
	}
	fmt.Println()

	// Test 2: If dengan inisialisasi
	fmt.Println("--- Test 2: If with init ---")
	if nama := "Budi"; len(nama) > 5 {
		fmt.Println("Nama panjang:", nama)
	} else {
		fmt.Println("Nama pendek:", nama)
	}
	if angka, err := strconv.Atoi("123"); err == nil {
		fmt.Printf("Parsed: %d (tipe: %T)\n", angka, angka)
	} else {
		fmt.Println("Error:", err)
	}
	fmt.Println()

	// Test 3: Switch
	fmt.Println("--- Test 3: Switch ---")
	hari := 3
	switch hari {
	case 1:
		fmt.Println("Senin")
	case 2:
		fmt.Println("Selasa")
	case 3:
		fmt.Println("Rabu")
	default:
		fmt.Println("Tidak valid")
	}

	nilai2 := 55
	switch {
	case nilai2 >= 70:
		fmt.Println("Cukup")
		fallthrough
	case nilai2 >= 50:
		fmt.Println("Lulus")
	default:
		fmt.Println("Tidak lulus")
	}
	fmt.Println()

	// Test 4: For
	fmt.Println("--- Test 4: For ---")
	for i := 0; i < 3; i++ {
		fmt.Println("Iterasi ke:", i)
	}
	fmt.Println("While-like:")
	j := 0
	for j < 3 {
		fmt.Println(j)
		j++
	}
	fmt.Println()

	// Test 5: Break & Continue
	fmt.Println("--- Test 5: Break & Continue ---")
	for i := 0; i < 5; i++ {
		if i == 3 {
			fmt.Println("break at i=3")
			break
		}
		if i%2 == 0 {
			fmt.Println("even:", i)
			continue
		}
		fmt.Println("odd:", i)
	}
	fmt.Println()

	// Test 6: Nested Loop
	fmt.Println("--- Test 6: Nested Loop ---")
	for a := 1; a <= 2; a++ {
		for b := 1; b <= 2; b++ {
			fmt.Printf("%d x %d = %d\n", a, b, a*b)
		}
		fmt.Println()
	}

	// Test 7: Label Break
	fmt.Println("--- Test 7: Label Break ---")
	outer:
	for i := 0; i < 3; i++ {
		for k := 0; k < 3; k++ {
			if k == 1 {
				fmt.Println("break outer")
				break outer
			}
			fmt.Println(i, k)
		}
	}
	fmt.Println()

	// Test 8: For Range
	fmt.Println("--- Test 8: For Range ---")
	buah := []string{"Apel", "Jeruk", "Mangga"}
	for idx, val := range buah {
		fmt.Printf("%d: %s\n", idx, val)
	}

	umur := map[string]int{"Budi": 25, "Siti": 22}
	for nama, umurVal := range umur {
		fmt.Printf("%s: %d\n", nama, umurVal)
	}

	for idx, ch := range "Halo" {
		fmt.Printf("%d: %c\n", idx, ch)
	}

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
