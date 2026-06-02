package main

import (
	"encoding/json"
	"fmt"
)

// Person basic struct
type Person struct {
	Name string
	Age  int
}

// Counter to demonstrate receiver behavior
type Counter struct {
	Value int
}

func (c Counter) AddVal(n int) {
	c.Value += n
}

func (c *Counter) AddPtr(n int) {
	c.Value += n
}

// Embedded structs
type Address struct {
	City   string
	Street string
}

type Employee struct {
	Person
	Address
	Position string
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 07: STRUCT ===\n")

	// Test 1: Basic struct
	fmt.Println("--- Test 1: Basic struct ---")
	var p Person
	p.Name = "Budi"
	p.Age = 25
	fmt.Printf("p: %+v (type %T)\n", p, p)
	fmt.Println()

	// Test 2: Literal init
	fmt.Println("--- Test 2: Literal init ---")
	p2 := Person{Name: "Siti", Age: 22}
	fmt.Printf("p2: %+v\n", p2)
	fmt.Println()

	// Test 3: Pointer to struct
	fmt.Println("--- Test 3: Pointer to struct ---")
	p3 := &Person{Name: "Ani"}
	fmt.Printf("before: p3: %+v\n", p3)
	p3.Age = 30
	fmt.Printf("after: p3: %+v\n", p3)
	fmt.Println()

	// Test 4: Anonymous struct
	fmt.Println("--- Test 4: Anonymous struct ---")
	anon := struct {
		X int
		Y int
	}{X: 1, Y: 2}
	fmt.Printf("anon: %+v\n", anon)
	fmt.Println()

	// Test 5: Value vs Pointer receiver
	fmt.Println("--- Test 5: Value vs Pointer receiver ---")
	c := Counter{Value: 0}
	c.AddVal(10)
	fmt.Printf("After AddVal(10) -> %d (value receiver, original unchanged)\n", c.Value)
	c.AddPtr(10)
	fmt.Printf("After AddPtr(10) -> %d (pointer receiver, original changed)\n", c.Value)
	fmt.Println()

	// Test 6: Embedded struct
	fmt.Println("--- Test 6: Embedded struct ---")
	e := Employee{
		Person:   Person{Name: "Rudi", Age: 28},
		Address:  Address{City: "Jakarta", Street: "Jl. Sudirman"},
		Position: "Engineer",
	}
	fmt.Printf("employee: %+v\n", e)
	fmt.Printf("Access embedded fields: Name=%s, City=%s\n", e.Name, e.Address.City)
	fmt.Println()

	// Test 7: JSON marshal/unmarshal
	fmt.Println("--- Test 7: JSON Marshal/Unmarshal ---")
	tg := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{Name: "Tono", Age: 40}
	b, _ := json.Marshal(tg)
	fmt.Println("json:", string(b))
	var tg2 struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	_ = json.Unmarshal(b, &tg2)
	fmt.Printf("unmarshal: %+v\n", tg2)
	fmt.Println()

	// Test 8: Struct comparability
	fmt.Println("--- Test 8: Struct comparability ---")
	a1 := Person{Name: "A", Age: 1}
	a2 := Person{Name: "A", Age: 1}
	fmt.Printf("a1 == a2 ? %v\n", a1 == a2)
	fmt.Println()

	// Test 9: Zero value struct
	fmt.Println("--- Test 9: Zero value struct ---")
	var pzero Person
	fmt.Printf("zero: %+v\n", pzero)
	fmt.Println()

	fmt.Println("=== SEMUA TEST SELESAI ===")
}
