package main

import "fmt"

// Speaker example
type Speaker interface {
	Speak() string
}

type Dog struct{ Name string }
func (d Dog) Speak() string { return "Woof, I'm " + d.Name }

type Cat struct{ Name string }
func (c *Cat) Speak() string { return "Meow, I'm " + c.Name }

func say(s Speaker) {
	fmt.Println("say:", s.Speak())
}

// Stringer example
type Person struct{ Name string }
func (p Person) String() string { return "Person(" + p.Name + ")" }

func main() {
	fmt.Println("=== MANUAL TEST MATERI 09: INTERFACE ===\n")

	// Test 1: Basic interface assignment
	fmt.Println("--- Test 1: Basic interface assignment ---")
	var s Speaker = Dog{Name: "Rex"}
	fmt.Println(s.Speak())
	say(Dog{Name: "Buddy"})
	say(&Cat{Name: "Kitty"})
	fmt.Println()

	// Test 2: Empty interface
	fmt.Println("--- Test 2: Empty interface (any) ---")
	var x interface{}
	x = "hello"
	fmt.Printf("x=%v (type %T)\n", x, x)
	x = 123
	fmt.Printf("x=%v (type %T)\n", x, x)
	fmt.Println()

	// Test 3: Type assertion & switch
	fmt.Println("--- Test 3: Type assertion & type switch ---")
	if str, ok := x.(string); ok {
		fmt.Println("string:", str)
	} else {
		fmt.Println("x is not string")
	}
	switch v := x.(type) {
	case int:
		fmt.Println("type switch: int", v)
	case string:
		fmt.Println("type switch: string", v)
	default:
		fmt.Println("type switch: other", v)
	}
	fmt.Println()

	// Test 4: nil interface gotcha
	fmt.Println("--- Test 4: nil interface gotcha ---")
	var sNil Speaker
	fmt.Println("sNil == nil ?", sNil == nil)
	var cPtr *Cat
	var sPtr Speaker = cPtr
	fmt.Println("sPtr == nil ? (interface with nil pointer)", sPtr == nil)
	fmt.Println()

	// Test 5: fmt.Stringer implementation
	fmt.Println("--- Test 5: fmt.Stringer ---")
	p := Person{Name: "Alicia"}
	fmt.Println("Person printing:", p)
	fmt.Println()

	// Test 6: Slice of interface
	fmt.Println("--- Test 6: Slice of interface ---")
	animals := []Speaker{Dog{Name: "Rex"}, &Cat{Name: "Mia"}}
	for _, a := range animals {
		fmt.Println(a.Speak())
	}
	fmt.Println()

	// Test 7: Type switch in loop (empty interface)
	fmt.Println("--- Test 7: Type switch in loop (empty interface) ---")
	items := []interface{}{"abc", 10, 3.14}
	for _, it := range items {
		switch v := it.(type) {
		case string:
			fmt.Println("string value:", v)
		case int:
			fmt.Println("int value:", v)
		default:
			fmt.Println("other:", v)
		}
	}
	fmt.Println()

	fmt.Println("=== SEMUA TEST SELESAI ===")
}
