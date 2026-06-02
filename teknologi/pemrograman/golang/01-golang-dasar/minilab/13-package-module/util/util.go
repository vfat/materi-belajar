package util

import "fmt"

// Add returns the sum of two integers
func Add(a, b int) int {
	return a + b
}

// Greeting returns a formatted greeting
func Greeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
