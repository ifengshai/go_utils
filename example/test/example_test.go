// math_test.go
package main

import "testing"

func TestAdd(t *testing.T) {
	result := 2 + 3
	expected := 5

	if result != expected {
		t.Errorf("Add(2, 3) returned %d, expected %d", result, expected)
	}
}

// go test -bench=Add -benchmem
func BenchmarkAdd(b *testing.B) {
	result := 2 + 3
	expected := 5

	if result != expected {
		b.Errorf("Add(2, 3) returned %d, expected %d", result, expected)
	}
}
