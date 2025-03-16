// Package gf256 represents elements in the Galois Field GF(2^8)
package gf256

import "crypto/subtle"

// IrreduciblePolynomial use the irreducible polynomial x^8 + x^4 + x^3 + x + 1 (0x11B in hex)
const IrreduciblePolynomial uint8 = 0x1B

var multiplicationTable [256][256]uint8
var inverseTable [256]uint8

func init() {
	multiplicationTable = GenerateMultiplicationTable()
	inverseTable = GenerateInverseTable()
}

// Add performs addition in GF(2^8), which is just XOR
func Add(a, b uint8) uint8 {
	return a ^ b
}

// Subtract in GF(2^8) is the same as addition (XOR)
func Subtract(a, b uint8) uint8 {
	return a ^ b
}

// Multiply performs multiplication in GF(2^8)
func Multiply(a, b uint8) uint8 {
	return multiplicationTable[a][b]
}

// Inverse returns multiplicative inverse in GF(2^8) using the Extended Euclidean Algorithm
func Inverse(a uint8) uint8 {
	if a == 0 {
		panic("Cannot compute inverse of 0 in GF(2^8)")
	}
	return inverseTable[a]
}

// Divide performs division in GF(2^8): a/b = a * b^(-1)
func Divide(a, b uint8) uint8 {
	if b == 0 {
		panic("Division by zero in GF(2^8)")
	}
	c := Multiply(a, Inverse(b))
	c = uint8(subtle.ConstantTimeSelect(subtle.ConstantTimeByteEq(a, 0), 0, int(c)))
	return c
}

// GenerateMultiplicationTable creates a 256x256 lookup table for multiplication
func GenerateMultiplicationTable() [256][256]uint8 {
	var table [256][256]uint8
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			table[i][j] = MultiplyDo(uint8(i), uint8(j))
		}
	}
	return table
}

// GenerateInverseTable creates a 256-element lookup table for multiplicative inverses
func GenerateInverseTable() [256]uint8 {
	var table [256]uint8
	// Zero has no inverse, set to 0
	table[0] = 0
	for i := 1; i < 256; i++ {
		table[i] = InverseDo(uint8(i))
	}
	return table
}

func MultiplyDo(a, b uint8) uint8 {
	var product uint8 = 0
	var temp = b
	for i := 0; i < 8; i++ {
		// If the i-th bit of 'a' is set
		if (a & (1 << i)) != 0 {
			product ^= temp
		}
		// Check if high bit is set before shifting
		highBitSet := (temp & 0x80) != 0
		// Shift left
		temp <<= 1
		// If high bit was set, XOR with the irreducible polynomial
		if highBitSet {
			temp ^= IrreduciblePolynomial
		}
	}
	return product
}

// InverseDo calculates the multiplicative inverse in GF(2^8) using the Extended Euclidean Algorithm
func InverseDo(a uint8) uint8 {
	if a == 0 {
		panic("Cannot compute inverse of 0 in GF(2^8)")
	}
	// Using Fermat's Little Theorem: a^(p-1) ≡ 1 (mod p)
	// For GF(2^8), a^(2^8-1) = a^255 ≡ 1, thus a^254 ≡ a^(-1)
	// We can compute this efficiently using square-and-multiply algorithm
	result := uint8(1) // Start with 1 as the identity element in GF(2^8)
	base := a
	for i := 0; i < 8; i++ {
		if (254>>i)&1 == 1 { // If the i-th bit of 254 is set
			result = Multiply(result, base)
		}
		base = Multiply(base, base) // Square the base on each iteration
	}
	return result
}
