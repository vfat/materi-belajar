package main

import "fmt"

func main() {
	fmt.Println("=== MANUAL TEST MATERI 03: OPERATOR & EKSPRESI ===\n")

	// Test 1: Operator Aritmatika
	a, b := 10, 3
	fmt.Println("--- Test 1: Operator Aritmatika ---")
	fmt.Printf("%d + %d = %d\n", a, b, a+b)
	fmt.Printf("%d - %d = %d\n", a, b, a-b)
	fmt.Printf("%d * %d = %d\n", a, b, a*b)
	fmt.Printf("%d / %d = %d\n", a, b, a/b)
	fmt.Printf("%d %% %d = %d\n", a, b, a%b)
	fmt.Printf("%d / %d (float) = %.2f\n\n", a, b, float64(a)/float64(b))

	// Test 2: Operator Perbandingan
	x, y := 5, 7
	fmt.Println("--- Test 2: Operator Perbandingan ---")
	fmt.Printf("%d == %d -> %v\n", x, y, x == y)
	fmt.Printf("%d != %d -> %v\n", x, y, x != y)
	fmt.Printf("%d > %d -> %v\n", x, y, x > y)
	fmt.Printf("%d < %d -> %v\n", x, y, x < y)
	fmt.Printf("%d >= %d -> %v\n", x, y, x >= y)
	fmt.Printf("%d <= %d -> %v\n\n", x, y, x <= y)

	// Test 3: Operator Logika
	umur := 20
	punyaKTP := true
	punyaUang := false
	fmt.Println("--- Test 3: Operator Logika ---")
	bisaBeli := umur >= 17 && punyaKTP
	bisaMasuk := punyaKTP || punyaUang
	tidak := !false
	fmt.Printf("umur>=17 && punyaKTP -> %v\n", bisaBeli)
	fmt.Printf("punyaKTP || punyaUang -> %v\n", bisaMasuk)
	fmt.Printf("!false -> %v\n\n", tidak)

	// Test 4: Operator Penugasan
	fmt.Println("--- Test 4: Operator Penugasan ---")
	z := 10
	fmt.Printf("awal z = %d\n", z)
	z += 5
	fmt.Printf("z +=5 -> %d\n", z)
	z -= 3
	fmt.Printf("z -=3 -> %d\n", z)
	z *= 2
	fmt.Printf("z *=2 -> %d\n", z)
	z /= 4
	fmt.Printf("z /=4 -> %d\n", z)
	z %= 5
	fmt.Printf("z %%=5 -> %d\n\n", z)

	// Test 5: Increment & Decrement
	fmt.Println("--- Test 5: Increment & Decrement ---")
	counter := 0
	counter++
	fmt.Printf("counter++ -> %d\n", counter)
	counter++
	fmt.Printf("counter++ -> %d\n", counter)
	counter--
	fmt.Printf("counter-- -> %d\n\n", counter)

	// Test 6: Operator Bitwise
	fmt.Println("--- Test 6: Operator Bitwise ---")
	aa := 5 // 0101
	bb := 3 // 0011
	fmt.Printf("a & b = %d\n", aa&bb)
	fmt.Printf("a | b = %d\n", aa|bb)
	fmt.Printf("a ^ b = %d\n", aa^bb)
	fmt.Printf("a &^ b = %d\n", aa&^bb)
	fmt.Printf("1 << 2 = %d\n", 1<<2)
	fmt.Printf("4 >> 1 = %d\n\n", 4>>1)

	// Test 7: Presedensi Operator
	fmt.Println("--- Test 7: Presedensi Operator ---")
	hasil := 2 + 3*4
	hasil2 := (2 + 3) * 4
	fmt.Printf("2 + 3 * 4 = %d\n", hasil)
	fmt.Printf("(2 + 3) * 4 = %d\n\n", hasil2)

	// Test 8: Contoh Gabungan
	fmt.Println("--- Test 8: Contoh Gabungan ---")
	nama := "Rina"
	nilaiTeori := 85
	nilaiPraktik := 90
	total := nilaiTeori + nilaiPraktik
	rata := float64(total) / 2.0
	lulusTeori := nilaiTeori >= 75
	lulusPraktik := nilaiPraktik >= 75
	lulusSemuanya := lulusTeori && lulusPraktik
	fmt.Println("=== Hasil Belajar ===")
	fmt.Printf("Nama: %s\n", nama)
	fmt.Printf("Nilai Teori: %d, Praktik: %d\n", nilaiTeori, nilaiPraktik)
	fmt.Printf("Total: %d\n", total)
	fmt.Printf("Rata-rata: %.1f\n", rata)
	fmt.Printf("Lulus? %v\n\n", lulusSemuanya)

	fmt.Println("=== SEMUA TEST SELESAI ===")
}
