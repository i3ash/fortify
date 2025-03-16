package gf256

import (
	"strconv"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct{ a, b, expected uint8 }{
		{0, 0, 0},
		{1, 0, 1},
		{0, 1, 1},
		{255, 255, 0},
		{0x0F, 0xF0, 0xFF},
		{0x53, 0xCA, 0x99},
		{0xAA, 0x55, 0xFF},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := Add(test.a, test.b)
			if result != test.expected {
				t.Errorf("Add(%#02x, %#02x) = %#02x; expected %#02x", test.a, test.b, result, test.expected)
			}
			result = Add(test.b, test.a)
			if result != test.expected {
				t.Errorf("Add(%#02x, %#02x) = %#02x; expected %#02x (commutativity check)", test.b, test.a, result, test.expected)
			}
			result = Subtract(test.a, test.b)
			if result != test.expected {
				t.Errorf("Subtract(%#02x, %#02x) = %#02x; expected %#02x", test.a, test.b, result, test.expected)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct{ a, b, expected uint8 }{
		{0, 0, 0},
		{0, 1, 0},
		{1, 1, 1},
		{2, 3, 6},
		{3, 7, 9},
		{0xFF, 0x01, 0xFF},
		{0x53, 0xCA, 0x01},
		{0x02, 0x02, 0x04},
		{0x02, 0x80, 0x1B},
		{0x53, 0xCA, MultiplyDo(0x53, 0xCA)}, // Double-check with direct calculation
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := Multiply(test.a, test.b)
			if result != test.expected {
				t.Errorf("Multiply(%#02x, %#02x) = %#02x; expected %#02x", test.a, test.b, result, test.expected)
			}
			// Multiplication should be commutative
			result = Multiply(test.b, test.a)
			if result != test.expected {
				t.Errorf("Multiply(%#02x, %#02x) = %#02x; expected %#02x (commutativity check)", test.b, test.a, result, test.expected)
			}
		})

	}
}

func TestDivide(t *testing.T) {
	tests := []struct{ a, b, expected uint8 }{
		{0, 1, 0}, // 0 divided by anything is 0
		{1, 1, 1}, // 1/1 = 1
		{2, 1, 2}, // a/1 = a
		{2, 2, 1}, // a/a = 1
		{0x53, 0xCA, Multiply(0x53, Inverse(0xCA))}, // Cross-check with multiply and inverse
		{Multiply(0x53, 0xCA), 0x53, 0xCA},          // Cross-check with multiply
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := Divide(test.a, test.b)
			if result != test.expected {
				t.Errorf("Divide(%#02x, %#02x) = %#02x; expected %#02x", test.a, test.b, result, test.expected)
			}
		})
	}
	// Division by zero should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Divide(1, 0) did not panic as expected")
		}
	}()
	Divide(1, 0)
}

func TestInverse(t *testing.T) {
	tests := []struct{ a, expected uint8 }{
		{1, 1},
		{0x53, 0xCA},
		{0xFF, 0x1C},
		{0x55, 0x24},
	}
	for i, test := range tests {
		t.Run("idx."+strconv.Itoa(i), func(t *testing.T) {
			result := Inverse(test.a)
			if result != test.expected {
				t.Errorf("Inverse(%#02x) = %#02x; expected %#02x", test.a, result, test.expected)
			}
			// Verify: a * a^(-1) = 1
			product := Multiply(test.a, result)
			if product != 1 {
				t.Errorf("Multiply(%#02x, Inverse(%#02x)) = %#02x; expected 1", test.a, test.a, product)
			}
		})
	}
	t.Run("1_255", func(t *testing.T) {
		// Test all non-zero elements have valid inverses
		for a := 1; a < 256; a++ {
			inv := Inverse(uint8(a))
			product := Multiply(uint8(a), inv)
			if product != 1 {
				t.Errorf("Multiply(%#02x, Inverse(%#02x)) = %#02x; expected 1", a, a, product)
			}
		}
	})
	// Test inverse of 0 should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Inverse(0) did not panic as expected")
		}
	}()
	Inverse(0)
}
