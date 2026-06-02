package main

import (
	"fmt"
	"sort"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 05: ARRAY, SLICE & MAP ===\n")

	// Test 1: Array
	fmt.Println("--- Test 1: Array ---")
	var arr [5]int = [5]int{10, 20, 30, 40, 50}
	fmt.Println("arr:", arr)
	fmt.Println("arr[0]:", arr[0], "arr[4]:", arr[4])
	fmt.Println("len:", len(arr))
	fmt.Println()

	// Test 2: Slice creation & properties
	fmt.Println("--- Test 2: Slice (create, len, cap, append) ---")
	s := []int{1, 2, 3}
	fmt.Printf("s: %v len=%d cap=%d\n", s, len(s), cap(s))
	s = append(s, 4, 5)
	fmt.Printf("after append: %v len=%d cap=%d\n", s, len(s), cap(s))
	s2 := make([]int, 3, 6)
	fmt.Printf("s2: %v len=%d cap=%d\n", s2, len(s2), cap(s2))
	fmt.Println()

	// Test 3: Copy
	fmt.Println("--- Test 3: Copy ---")
	src := []int{7, 8, 9}
	dst := make([]int, len(src))
	copied := copy(dst, src)
	fmt.Println("src:", src)
	fmt.Println("dst:", dst, "copied:", copied)
	fmt.Println()

	// Test 4: Remove & Insert
	fmt.Println("--- Test 4: Remove & Insert ---")
	srm := []int{1, 2, 3, 4, 5}
	idx := 2 // remove element at index 2 (value 3)
	srm = append(srm[:idx], srm[idx+1:]...)
	fmt.Println("after remove idx 2:", srm)

	sins := []int{1, 2, 4, 5}
	insertIdx := 2
	sins = append(sins[:insertIdx], append([]int{99}, sins[insertIdx:]...)...)
	fmt.Println("after insert 99 at idx 2:", sins)
	fmt.Println()

	// Test 5: Subslice & underlying array
	fmt.Println("--- Test 5: Subslice & underlying array ---")
	arr2 := [6]int{10, 20, 30, 40, 50, 60}
	ss := arr2[1:4] // [20,30,40]
	fmt.Println("arr2:", arr2)
	fmt.Println("ss:", ss)
	ss[1] = 300
	fmt.Println("after ss[1]=300")
	fmt.Println("arr2:", arr2)
	fmt.Println("ss:", ss)
	fmt.Println()

	// Test 6: Map operations
	fmt.Println("--- Test 6: Map operations ---")
	m := map[string]int{"Budi": 25, "Siti": 22}
	fmt.Println("m:", m)
	m["Rudi"] = 30
	fmt.Println("after add Rudi:", m)
	val, ok := m["Ani"]
	fmt.Println("get Ani:", val, "ok?", ok)
	if _, ok2 := m["Budi"]; ok2 {
		fmt.Println("Budi ada:", m["Budi"])
	}
	delete(m, "Siti")
	fmt.Println("after delete Siti:", m)
	fmt.Println()

	// Test 7: Iterate map unsorted vs sorted
	fmt.Println("--- Test 7: Iterate map (sorted) ---")
	m2 := map[string]int{"Matematika": 90, "Bahasa": 85, "IPA": 88}
	fmt.Println("unsorted iteration:")
	for k, v := range m2 {
		fmt.Printf("%s:%d ", k, v)
	}
	fmt.Println("\nsorted iteration:")
	keys := make([]string, 0, len(m2))
	for k := range m2 {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s:%d ", k, m2[k])
	}
	fmt.Println("\n")

	// Test 8: Find max/min from slice
	fmt.Println("--- Test 8: Find max & min in slice ---")
	nums := []int{5, 3, 9, 1, 7}
	max := nums[0]
	min := nums[0]
	for _, v := range nums {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	fmt.Println("nums:", nums, "max:", max, "min:", min)
	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
