package main

import (
	"errors"
	"fmt"
	"os"
)

var ErrNotFound = errors.New("not found")

type MyError struct {
	Msg  string
	Code int
}

func (e *MyError) Error() string {
	return fmt.Sprintf("MyError(code=%d): %s", e.Code, e.Msg)
}

func findItem(name string) error {
	if name == "bob" {
		return nil
	}
	return fmt.Errorf("findItem: %w", ErrNotFound)
}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide: %w", errors.New("divide by zero"))
	}
	return a / b, nil
}

func returnMyError() error {
	return &MyError{Msg: "resource missing", Code: 404}
}

func mayPanic() {
	panic("unexpected panic here")
}

func safeCall() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	mayPanic()
	fmt.Println("this will not run")
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 10: ERROR HANDLING ===\n")

	// Test 1: Basic error and fmt.Errorf
	fmt.Println("--- Test 1: Basic error & wrapping ---")
	err := fmt.Errorf("basic error: %s", "something went wrong")
	fmt.Println(err)
	fmt.Println()

	// Test 2: Multiple return with error
	fmt.Println("--- Test 2: Multiple return (divide) ---")
	if _, err := divide(10, 0); err != nil {
		fmt.Println("divide error:", err)
	}
	if res, err := divide(10, 2); err == nil {
		fmt.Println("10 / 2 =", res)
	}
	fmt.Println()

	// Test 3: Sentinel error + errors.Is
	fmt.Println("--- Test 3: Sentinel error & errors.Is ---")
	err = findItem("alice")
	if err != nil {
		fmt.Println("findItem error:", err)
		if errors.Is(err, ErrNotFound) {
			fmt.Println("-> errors.Is: underlying ErrNotFound detected")
		}
	}
	fmt.Println()

	// Test 4: Custom error type & errors.As
	fmt.Println("--- Test 4: Custom error type & errors.As ---")
	err = returnMyError()
	var me *MyError
	if errors.As(err, &me) {
		fmt.Printf("got MyError: code=%d msg=%s\n", me.Code, me.Msg)
	} else {
		fmt.Println("not MyError, err:", err)
	}
	fmt.Println()

	// Test 5: Panic & Recover
	fmt.Println("--- Test 5: Panic & Recover ---")
	safeCall()
	fmt.Println()

	// Test 6: OS file open error
	fmt.Println("--- Test 6: OS file open error ---")
	if _, err := os.Open("this-file-does-not-exist.txt"); err != nil {
		fmt.Println("open error:", err)
	}
	fmt.Println()

	// Test 7: Error wrapping and unwrapping
	fmt.Println("--- Test 7: Error wrapping & unwrapping ---")
	base := errors.New("root cause")
	wrapped := fmt.Errorf("higher: %w", base)
	if errors.Is(wrapped, base) {
		fmt.Println("errors.Is finds base error")
	}
	fmt.Println()

	fmt.Println("=== SEMUA TEST SELESAI ===")
}
