package main

import "fmt"

type Point struct{ X, Y int }

func incPtr(n *int) {
	*n = *n + 1
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 08: POINTER ===\n")

	// Test 1: Pointer basics
	fmt.Println("--- Test 1: Pointer basics ---")
	a := 10
	p := &a
	fmt.Printf("a=%d, p=%p, *p=%d\n", a, p, *p)
	*p = 20
	fmt.Printf("after *p=20 -> a=%d\n", a)
	fmt.Println()

	// Test 2: Nil pointer
	fmt.Println("--- Test 2: Nil pointer ---")
	var pn *int
	fmt.Printf("pn == nil? %v\n", pn == nil)
	fmt.Println()

	// Test 3: new()
	fmt.Println("--- Test 3: new() ---")
	n := new(int)
	fmt.Printf("n addr=%p val=%d\n", n, *n)
	*n = 7
	fmt.Printf("after *n=7 -> %d\n", *n)
	fmt.Println()

	// Test 4: Function param pointer
	fmt.Println("--- Test 4: Function param pointer ---")
	b := 5
	fmt.Printf("before b=%d\n", b)
	incPtr(&b)
	fmt.Printf("after incPtr b=%d\n", b)
	fmt.Println()

	// Test 5: Pointer to array element
	fmt.Println("--- Test 5: Pointer to array element ---")
	arr := [3]int{1, 2, 3}
	pp := &arr[1]
	fmt.Printf("arr before: %v\n", arr)
	*pp = 99
	fmt.Printf("arr after: %v\n", arr)
	fmt.Println()

	// Test 6: Pointer to struct
	fmt.Println("--- Test 6: Pointer to struct ---")
	pt := &Point{X: 1, Y: 2}
	fmt.Printf("pt before: %+v\n", pt)
	pt.X = 10
	fmt.Printf("pt after: %+v\n", pt)
	fmt.Println()

	// Test 7: Pointer equality
	fmt.Println("--- Test 7: Pointer equality ---")
	x := 1
	y := 1
	px := &x
	py := &y
	fmt.Printf("&x == &y ? %v\n", px == py)
	fmt.Println()

	fmt.Println("=== SEMUA TEST SELESAI ===")
}
