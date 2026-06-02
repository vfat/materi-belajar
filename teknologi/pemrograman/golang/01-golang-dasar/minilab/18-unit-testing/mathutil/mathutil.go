package mathutil

import "errors"

// Add returns the sum of two integers
func Add(a, b int) int {
	return a + b
}

// Multiply returns the product of two integers
func Multiply(a, b int) int {
	return a * b
}

// Divide returns integer division result or error if division by zero
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}
	return a / b, nil
}

// Max returns the maximum value from a list of integers. Returns 0 if empty.
func Max(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, v := range nums[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// Min returns the minimum value from a list of integers. Returns 0 if empty.
func Min(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, v := range nums[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
