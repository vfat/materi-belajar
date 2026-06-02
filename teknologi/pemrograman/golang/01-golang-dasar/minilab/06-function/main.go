package main

import "fmt"

func main() {
	fmt.Println("=== MANUAL TEST MATERI 06: FUNCTION ===\n")

	// Test 1: Function declaration & call
	fmt.Println("--- Test 1: Function Declaration ---")	
	greet("Budi")
	fmt.Println()

	// Test 2: Function with return
	fmt.Println("--- Test 2: Function with Return ---")
	fmt.Printf("3 + 5 = %d\n", add(3, 5))
	fmt.Println()

	// Test 3: Multiple returns
	fmt.Println("--- Test 3: Multiple Returns ---")
	q, r := divMod(17, 3)
	fmt.Printf("17 / 3 = %d remainder %d\n", q, r)
	fmt.Println()

	// Test 4: Named return values
	fmt.Println("--- Test 4: Named Returns ---")
	sum, cnt := sumAndCount([]int{1, 2, 3, 4})
	fmt.Printf("sum=%d count=%d\n", sum, cnt)
	fmt.Println()

	// Test 5: Variadic function
	fmt.Println("--- Test 5: Variadic ---")
	fmt.Printf("sumVariadic(1,2,3,4) = %d\n", sumVariadic(1, 2, 3, 4))
	nums := []int{5, 6, 7}
	fmt.Printf("sumVariadic(nums...) = %d\n", sumVariadic(nums...))
	fmt.Println()

	// Test 6: Defer order
	fmt.Println("--- Test 6: Defer ---")
	deferDemo()
	fmt.Println()

	// Test 7: Anonymous function & closure
	fmt.Println("--- Test 7: Anonymous / Closure ---")
	inc := makeIncrementer(10)
	fmt.Println("inc():", inc())
	fmt.Println("inc():", inc())
	anon := func(msg string) { fmt.Println("Anon:", msg) }
	anon("Hello anonymous")
	fmt.Println()

	// Test 8: Recursion
	fmt.Println("--- Test 8: Recursion ---")
	fmt.Println("factorial(5) =", factorial(5))
	fmt.Println()

	// Test 9: Function as parameter
	fmt.Println("--- Test 9: Function as Parameter ---")
	res := applyFunc(3, 4, func(a, b int) int { return a * b })
	fmt.Println("apply multiply 3*4 =", res)
	fmt.Println()

	fmt.Println("=== SEMUA TEST SELESAI ===")
}

func greet(name string) {
	fmt.Println("Halo,", name)
}

func add(a, b int) int {
	return a + b
}

func divMod(a, b int) (int, int) {
	return a / b, a % b
}

func sumAndCount(nums []int) (sum int, count int) {
	for _, v := range nums {
		sum += v
		count++
	}
	return
}

func sumVariadic(nums ...int) int {
	s := 0
	for _, v := range nums {
		s += v
	}
	return s
}

func deferDemo() {
	fmt.Println("start deferDemo")
	defer fmt.Println("defer 1")
	defer fmt.Println("defer 2")
	fmt.Println("end deferDemo")
}

func makeIncrementer(start int) func() int {
	return func() int {
		start++
		return start
	}
}

func factorial(n int) int {
	if n == 0 {
		return 1
	}
	return n * factorial(n-1)
}

func applyFunc(a, b int, fn func(int, int) int) int {
	return fn(a, b)
}
