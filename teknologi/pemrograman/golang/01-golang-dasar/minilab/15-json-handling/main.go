package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type Person struct {
	Name  string   `json:"name"`
	Age   int      `json:"age,omitempty"`
	Email string   `json:"email,omitempty"`
	Tags  []string `json:"tags,omitempty"`
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 15: JSON HANDLING ===\n")

	fmt.Println("--- Test 1: Marshal struct & MarshalIndent ---")
	p := Person{Name: "Gopher", Age: 5, Tags: []string{"go", "dev"}}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		fmt.Println("marshal error:", err)
	} else {
		fmt.Println(string(b))
	}

	fmt.Println("\n--- Test 2: Unmarshal to struct ---")
	js := `{"name":"Gopher","age":5,"tags":["go","tools"]}`
	var p2 Person
	if err := json.Unmarshal([]byte(js), &p2); err != nil {
		fmt.Println("unmarshal error:", err)
	} else {
		fmt.Printf("%+v\n", p2)
	}

	fmt.Println("\n--- Test 3: Unmarshal to map[string]interface{} ---")
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		fmt.Println("unmarshal map error:", err)
	} else {
		fmt.Printf("%v\n", m)
	}

	fmt.Println("\n--- Test 4: json.Decoder (stream) ---")
	arrJs := `[{"name":"A"},{"name":"B"}]`
	dec := json.NewDecoder(bytes.NewBufferString(arrJs))
	var persons []Person
	if err := dec.Decode(&persons); err != nil {
		fmt.Println("decoder error:", err)
	} else {
		fmt.Printf("%#v\n", persons)
	}

	fmt.Println("\n--- Test 5: json.Encoder ---")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode([]Person{p, p2}); err != nil {
		fmt.Println("encode error:", err)
	}

	fmt.Println("\n--- Test 6: Invalid JSON error handling ---")
	bad := `{"name": "bad", "age": }`
	var x interface{}
	if err := json.Unmarshal([]byte(bad), &x); err != nil {
		fmt.Println("expected error:", err)
	} else {
		fmt.Println("unexpected success")
	}

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
