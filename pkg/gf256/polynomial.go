package gf256

import (
	"fmt"
	"io"
)

type PolynomialCoefficients = []uint8

func NewPolynomial(intercept uint8, degree uint8, random io.Reader) (PolynomialCoefficients, error) {
	if degree == 0 {
		return []byte{intercept}, nil
	}
	coefficients := make([]byte, degree+1)
	coefficients[0] = intercept
	// Read all coefficients at once
	if _, err := io.ReadFull(random, coefficients[1:]); err != nil {
		return nil, fmt.Errorf("failed to read random coefficients: %w", err)
	}
	// Handle zero coefficients
	for i := 1; i <= int(degree); i++ {
		// Only ensure the highest degree coefficient is non-zero to maintain actual degree
		if i == int(degree) && coefficients[i] == 0 {
			const maxAttempts = 50 // Set maximum attempts to avoid infinite loops
			nonZero := make([]byte, 1)
			for attempt := 0; attempt < maxAttempts; attempt++ {
				nonZero[0] = 0 // Reset to zero
				if _, err := io.ReadFull(random, nonZero); err != nil {
					return coefficients, fmt.Errorf("failed to read non-zero coefficient: %w", err)
				}
				if nonZero[0] != 0 {
					coefficients[i] = nonZero[0]
					break
				}
			}
			// If still zero after multiple attempts, set default non-zero value
			if coefficients[i] == 0 {
				coefficients[i] = 1 // Use 1 as default non-zero value
			}
		}
	}
	return coefficients, nil
}

// PolynomialEvaluate returns the value of the polynomial for the given x
// using Horner's method for efficient computation.
func PolynomialEvaluate(coefficients PolynomialCoefficients, x uint8) uint8 {
	// Handle empty polynomial
	if len(coefficients) == 0 {
		return 0
	}
	// Boundary condition optimization
	if x == 0 {
		return coefficients[0]
	}
	degree := len(coefficients) - 1
	y := coefficients[degree]
	// Use Horner's method to calculate polynomial value, skip zero coefficients
	for i := degree - 1; i >= 0; i-- {
		if coefficients[i] == 0 && i > 0 {
			// If intermediate coefficient is zero, just multiply by x
			y = Multiply(y, x)
		} else {
			y = Add(Multiply(y, x), coefficients[i])
		}
	}
	return y
}

// InterpolatePolynomial takes N sample points and returns
// the value at a given x using a lagrange interpolation.
func InterpolatePolynomial(xSamples, ySamples []uint8, x uint8) uint8 {
	limit := len(xSamples)
	if limit > len(ySamples) {
		limit = len(ySamples)
	}
	if limit == 0 {
		return 0
	}
	// Check if x equals any sample point x value, if so return corresponding y value
	for i := 0; i < limit; i++ {
		if x == xSamples[i] {
			return ySamples[i]
		}
	}
	// Pre-calculate denominators to avoid repeated computation
	denominators := make([]uint8, limit)
	for i := 0; i < limit; i++ {
		denominator := uint8(1)
		for j := 0; j < limit; j++ {
			if i == j {
				continue
			}
			denominator = Multiply(denominator, Add(xSamples[i], xSamples[j]))
		}
		denominators[i] = denominator
	}
	var result uint8
	for i := 0; i < limit; i++ {
		if ySamples[i] == 0 {
			continue // Skip calculation if y value is 0
		}
		basis := uint8(1)
		for j := 0; j < limit; j++ {
			if i == j {
				continue
			}
			basis = Multiply(basis, Add(x, xSamples[j]))
		}
		// Use pre-calculated denominator
		term := Divide(basis, denominators[i])
		group := Multiply(ySamples[i], term)
		result = Add(result, group)
	}
	return result
}
