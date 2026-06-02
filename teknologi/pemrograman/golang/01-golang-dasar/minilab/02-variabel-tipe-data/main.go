package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 02: VARIABEL & TIPE DATA ===\n")

	// Test 1: Declarasi Variabel dengan var
	fmt.Println("--- Test 1: Declarasi dengan var ---")
	var nama string = "Budi"
	var umur int = 25
	var aktif bool = true
	fmt.Printf("nama: %v (tipe: %T)\n", nama, nama)
	fmt.Printf("umur: %v (tipe: %T)\n", umur, umur)
	fmt.Printf("aktif: %v (tipe: %T)\n", aktif, aktif)
	fmt.Println()

	// Test 2: Declarasi tanpa inisialisasi (Zero Value)
	fmt.Println("--- Test 2: Zero Value (Nilai Default) ---")
	var namaKosong string
	var umurDefault int
	var aktifDefault bool
	fmt.Printf("namaKosong: '%s' (default string kosong)\n", namaKosong)
	fmt.Printf("umurDefault: %d (default 0)\n", umurDefault)
	fmt.Printf("aktifDefault: %v (default false)\n", aktifDefault)
	fmt.Println()

	// Test 3: Short Declaration :=
	fmt.Println("--- Test 3: Short Declaration := ---")
	kota := "Jakarta"
	populasi := 10600000
	metropolitan := true
	fmt.Printf("kota: %v (tipe: %T)\n", kota, kota)
	fmt.Printf("populasi: %v (tipe: %T)\n", populasi, populasi)
	fmt.Printf("metropolitan: %v (tipe: %T)\n", metropolitan, metropolitan)
	fmt.Println()

	// Test 4: Tipe Integer
	fmt.Println("--- Test 4: Tipe Integer ---")
	var a uint8 = 255
	var b int8 = -128
	var c int = 42
	var d int64 = 9223372036854775807
	fmt.Printf("uint8 (max): %v\n", a)
	fmt.Printf("int8 (min): %v\n", b)
	fmt.Printf("int: %v\n", c)
	fmt.Printf("int64 (max): %v\n", d)
	fmt.Println()

	// Test 5: Tipe Float
	fmt.Println("--- Test 5: Tipe Float ---")
	var phi float32 = 3.14
	var pi float64 = 3.141592653589
	var gaji float64 = 5500000.50
	fmt.Printf("phi (float32): %.5f\n", phi)
	fmt.Printf("pi (float64): %.10f\n", pi)
	fmt.Printf("gaji: Rp %.2f\n", gaji)
	fmt.Println()

	// Test 6: Tipe String
	fmt.Println("--- Test 6: Tipe String ---")
	pesan := "Halo Dunia"
	multiline := `Baris pertama
Baris kedua
Baris ketiga`
	fmt.Printf("String biasa: %s\n", pesan)
	fmt.Println("String multi-line:")
	fmt.Println(multiline)
	fmt.Println()

	// Test 7: Tipe Boolean
	fmt.Println("--- Test 7: Tipe Boolean ---")
	lulus := 80 > 60
	tidakAda := 5 == 10
	fmt.Printf("80 > 60 = %v\n", lulus)
	fmt.Printf("5 == 10 = %v\n", tidakAda)
	fmt.Println()

	// Test 8: Konversi Tipe Data
	fmt.Println("--- Test 8: Konversi Tipe Data ---")
	intValue := 42
	floatValue := float64(intValue)
	backToInt := int(floatValue)
	fmt.Printf("int: %v → float64: %.2f → int: %v\n", intValue, floatValue, backToInt)

	// String ke number
	angka, err := strconv.Atoi("123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("String '123' ke int: %v (tipe: %T)\n", angka, angka)
	}

	// Number ke string
	teks := strconv.Itoa(456)
	fmt.Printf("int 456 ke string: '%v' (tipe: %T)\n", teks, teks)
	fmt.Println()

	// Test 9: Konstanta
	fmt.Println("--- Test 9: Konstanta ---")
	const PHI float64 = 3.14159
	const NAMA = "Go Learning"
	const UKURAN = 10
	fmt.Printf("PHI: %v\n", PHI)
	fmt.Printf("NAMA: %v\n", NAMA)
	fmt.Printf("UKURAN: %v\n", UKURAN)
	fmt.Println()

	// Test 10: Contoh Lengkap (dari materi)
	fmt.Println("--- Test 10: Contoh Lengkap (Profil) ---")
	var namaMahasiswa string = "Siti"
	umurMahasiswa := 22
	var tinggi float64 = 165.5
	var mahasiswa bool = true
	var pekerjaan string // zero value ""

	fmt.Println("=== Profil Mahasiswa ===")
	fmt.Printf("Nama: %s\n", namaMahasiswa)
	fmt.Printf("Umur: %d tahun\n", umurMahasiswa)
	fmt.Printf("Tinggi: %.1f cm\n", tinggi)
	fmt.Printf("Status Mahasiswa: %v\n", mahasiswa)
	fmt.Printf("Pekerjaan: %s (kosong)\n", pekerjaan)
	fmt.Println()

	// Test 11: Latihan - Profil Pribadi
	fmt.Println("--- Test 11: LATIHAN - Profil Pribadi ---")
	namaKamu := "Ahmad"
	umurKamu := 23
	tinggiKamu := 170.5
	sedangKuliah := true

	fmt.Println("=== Profil Saya ===")
	fmt.Printf("Nama: %s\n", namaKamu)
	fmt.Printf("Umur: %d tahun\n", umurKamu)
	fmt.Printf("Tinggi: %.1f cm\n", tinggiKamu)
	fmt.Printf("Sedang Kuliah: %v\n", sedangKuliah)
	fmt.Println()

	fmt.Println("=== SEMUA TEST SELESAI ===")
}
