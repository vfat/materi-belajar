package mathutil

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct{
		name string
		a, b int
		want int
	}{
		{"pos", 2, 3, 5},
		{"neg", -1, -2, -3},
		{"zero", 0, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			if got := Add(tt.a, tt.b); got != tt.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	if got := Multiply(3, 4); got != 12 {
		t.Fatalf("Multiply failed: got %d want 12", got)
	}
}

func TestDivide(t *testing.T) {
	res, err := Divide(10, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 5 {
		t.Fatalf("Divide wrong result: %d", res)
	}

	_, err = Divide(1, 0)
	if err == nil {
		t.Fatalf("expected error for divide by zero")
	}
}

func TestMaxMin(t *testing.T) {
	if got := Max(1, 5, 3, 9, 2); got != 9 {
		t.Fatalf("Max wrong: %d", got)
	}
	if got := Min(1, 5, 3, 9, 2); got != 1 {
		t.Fatalf("Min wrong: %d", got)
	}

	// empty
	if got := Max(); got != 0 {
		t.Fatalf("Max empty should be 0: %d", got)
	}
	if got := Min(); got != 0 {
		t.Fatalf("Min empty should be 0: %d", got)
	}
}

// Example for documentation/testing
func ExampleAdd() {
	fmt.Println(Add(2,3))
	// Output:
	// 5
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(100, 200)
	}
}
