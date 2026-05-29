package gf256

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestNewPolynomial_Degree0(t *testing.T) {
	coeffs, err := NewPolynomial(0x42, 0, nil)
	if err != nil {
		t.Fatalf("NewPolynomial failed: %v", err)
	}
	if len(coeffs) != 1 {
		t.Fatalf("expected 1 coefficient, got %d", len(coeffs))
	}
	if coeffs[0] != 0x42 {
		t.Errorf("expected coefficient 0x42, got %#02x", coeffs[0])
	}
}

func TestNewPolynomial_Degree1(t *testing.T) {
	coeffs, err := NewPolynomial(0x42, 1, rand.Reader)
	if err != nil {
		t.Fatalf("NewPolynomial failed: %v", err)
	}
	if len(coeffs) != 2 {
		t.Fatalf("expected 2 coefficients, got %d", len(coeffs))
	}
	if coeffs[0] != 0x42 {
		t.Errorf("expected intercept 0x42, got %#02x", coeffs[0])
	}
}

func TestNewPolynomial_Degree3(t *testing.T) {
	coeffs, err := NewPolynomial(0x01, 3, rand.Reader)
	if err != nil {
		t.Fatalf("NewPolynomial failed: %v", err)
	}
	if len(coeffs) != 4 {
		t.Fatalf("expected 4 coefficients, got %d", len(coeffs))
	}
	if coeffs[0] != 0x01 {
		t.Errorf("expected intercept 0x01, got %#02x", coeffs[0])
	}
	// Highest coefficient should be non-zero
	if coeffs[3] == 0 {
		t.Error("highest degree coefficient should be non-zero")
	}
}

func TestNewPolynomial_LeadingCoefficientNonZero(t *testing.T) {
	// Run multiple times to verify leading coefficient is always non-zero
	for i := 0; i < 20; i++ {
		coeffs, err := NewPolynomial(0x00, 2, rand.Reader)
		if err != nil {
			t.Fatalf("NewPolynomial failed: %v", err)
		}
		if coeffs[2] == 0 {
			t.Error("leading coefficient should never be zero")
		}
	}
}

func TestPolynomialEvaluate_Zero(t *testing.T) {
	// P(0) should always equal the intercept (constant term)
	coeffs := PolynomialCoefficients{0x42, 0x13, 0x37}
	result := PolynomialEvaluate(coeffs, 0)
	if result != coeffs[0] {
		t.Errorf("P(0) = %#02x, expected intercept %#02x", result, coeffs[0])
	}
}

func TestPolynomialEvaluate_Linear(t *testing.T) {
	// P(x) = 2x + 3
	coeffs := PolynomialCoefficients{3, 2}
	tests := []struct {
		x, expected uint8
	}{
		{0, 3},
		{1, 1}, // 2*1 + 3 = 2 XOR 3 = 1
		{2, 7}, // 2*2 + 3 = 4 XOR 3 = 7
	}
	for _, tt := range tests {
		result := PolynomialEvaluate(coeffs, tt.x)
		if result != tt.expected {
			t.Errorf("P(%d) = %#02x, expected %#02x", tt.x, result, tt.expected)
		}
	}
}

func TestPolynomialEvaluate_Quadratic(t *testing.T) {
	// P(x) = 3x^2 + 2x + 1 (using GF arithmetic)
	coeffs := PolynomialCoefficients{1, 2, 3}
	for x := uint8(0); x < 10; x++ {
		result := PolynomialEvaluate(coeffs, x)
		// Manual check: result = 3x^2 + 2x + 1
		x2 := Multiply(x, x)
		expected := Add(Add(Multiply(3, x2), Multiply(2, x)), 1)
		if result != expected {
			t.Errorf("P(%d) = %#02x, expected %#02x", x, result, expected)
		}
	}
}

func TestPolynomialEvaluate_DifferentDegree(t *testing.T) {
	coeffs2 := PolynomialCoefficients{5, 3}       // degree 1
	coeffs5 := PolynomialCoefficients{1, 2, 3, 4, 5, 6} // degree 5

	for x := uint8(0); x < 5; x++ {
		r2 := PolynomialEvaluate(coeffs2, x)
		r5 := PolynomialEvaluate(coeffs5, x)
		// Just check they don't panic and return values
		if r5 == r2 && x != 0 {
			// Different polynomials at same x should differ (unlikely coincidence)
			t.Logf("P2(%d)=%#02x, P5(%d)=%#02x (coincidence check)", x, r2, x, r5)
		}
	}
}

func TestInterpolatePolynomial_SinglePoint(t *testing.T) {
	result := InterpolatePolynomial(
		[]uint8{5},
		[]uint8{0x42},
		5,
	)
	if result != 0x42 {
		t.Errorf("single point interpolation at known x = %#02x, expected %#02x", result, 0x42)
	}
}

func TestInterpolatePolynomial_Linear(t *testing.T) {
	// Points from P(x) = 2x + 3
	xs := []uint8{1, 2}
	ys := []uint8{
		PolynomialEvaluate(PolynomialCoefficients{3, 2}, 1),
		PolynomialEvaluate(PolynomialCoefficients{3, 2}, 2),
	}

	// Interpolate at x=0 should give intercept 3
	result := InterpolatePolynomial(xs, ys, 0)
	if result != 3 {
		t.Errorf("interpolated P(0) = %#02x, expected 3", result)
	}

	// Interpolate at x=3 should give P(3)
	expected := PolynomialEvaluate(PolynomialCoefficients{3, 2}, 3)
	result = InterpolatePolynomial(xs, ys, 3)
	if result != expected {
		t.Errorf("interpolated P(3) = %#02x, expected %#02x", result, expected)
	}
}

func TestInterpolatePolynomial_Quadratic(t *testing.T) {
	// P(x) = 3x^2 + 2x + 1
	coeffs := PolynomialCoefficients{1, 2, 3}
	xs := []uint8{1, 2, 3}
	ys := make([]uint8, 3)
	for i, x := range xs {
		ys[i] = PolynomialEvaluate(coeffs, x)
	}

	// Interpolate at 3 different points should recover the function
	for x := uint8(0); x < 5; x++ {
		result := InterpolatePolynomial(xs, ys, x)
		expected := PolynomialEvaluate(coeffs, x)
		if result != expected {
			t.Errorf("interpolated P(%d) = %#02x, expected %#02x", x, result, expected)
		}
	}
}

func TestInterpolatePolynomial_SSSRecovery(t *testing.T) {
	// Simulate SSS: split secret at intercept using 3 shares (degree 2)
	secret := uint8(0x7F)
	coeffs, err := NewPolynomial(secret, 2, rand.Reader)
	if err != nil {
		t.Fatalf("NewPolynomial failed: %v", err)
	}

	// Generate 3 sample points
	xs := []uint8{1, 2, 3}
	ys := make([]uint8, 3)
	for i, x := range xs {
		ys[i] = PolynomialEvaluate(coeffs, x)
	}

	// Recover secret at x=0 (the intercept)
	recovered := InterpolatePolynomial(xs, ys, 0)
	if recovered != secret {
		t.Errorf("SSS recovery: got %#02x, expected secret %#02x", recovered, secret)
	}
}

func TestInterpolatePolynomial_5PointsThreshold3(t *testing.T) {
	// Simulate SSS with 5 shares, threshold 3 (degree 2)
	secret := uint8(0xAB)
	coeffs, err := NewPolynomial(secret, 2, rand.Reader)
	if err != nil {
		t.Fatalf("NewPolynomial failed: %v", err)
	}

	// Generate 5 points
	allXs := []uint8{1, 2, 3, 4, 5}
	allYs := make([]uint8, 5)
	for i, x := range allXs {
		allYs[i] = PolynomialEvaluate(coeffs, x)
	}

	// Test all combinations of 3 points out of 5
	combinator := [][]int{
		{0, 1, 2}, {0, 1, 3}, {0, 1, 4},
		{0, 2, 3}, {0, 2, 4}, {0, 3, 4},
		{1, 2, 3}, {1, 2, 4}, {1, 3, 4},
		{2, 3, 4},
	}

	for _, idx := range combinator {
		xs := []uint8{allXs[idx[0]], allXs[idx[1]], allXs[idx[2]]}
		ys := []uint8{allYs[idx[0]], allYs[idx[1]], allYs[idx[2]]}

		recovered := InterpolatePolynomial(xs, ys, 0)
		if recovered != secret {
			t.Errorf("subset %v: recovered %#02x, expected secret %#02x", idx, recovered, secret)
		}
	}

	// With only 2 points (below threshold), should NOT recover the secret
	recovered2 := InterpolatePolynomial(allXs[:2], allYs[:2], 0)
	if recovered2 == secret {
		t.Log("2 points coincidentally recovered the secret (rare)")
	}
}

func TestInterpolatePolynomial_AtSamplePoint(t *testing.T) {
	// Interpolating at a sample point should return the corresponding y value
	xs := []uint8{1, 5, 9}
	ys := []uint8{0x11, 0x55, 0x99}

	for i := range xs {
		result := InterpolatePolynomial(xs, ys, xs[i])
		if result != ys[i] {
			t.Errorf("interpolated at x=%d should return y=%#02x, got %#02x", xs[i], ys[i], result)
		}
	}
}

func TestInterpolatePolynomial_EmptyInput(t *testing.T) {
	result := InterpolatePolynomial([]uint8{}, []uint8{}, 0)
	if result != 0 {
		t.Errorf("empty input should return 0, got %#02x", result)
	}
}

func TestInterpolatePolynomial_ZeroYValues(t *testing.T) {
	// P(x) = 0x^2 + 0x + 5 = just intercept 5
	coeffs := PolynomialCoefficients{5, 0, 0}
	xs := []uint8{1, 2, 3}
	ys := make([]uint8, 3)
	for i, x := range xs {
		ys[i] = PolynomialEvaluate(coeffs, x)
	}

	result := InterpolatePolynomial(xs, ys, 0)
	if result != 5 {
		t.Errorf("interpolation with zero coefficients: got %#02x, expected 5", result)
	}
}

func TestPolynomialIntegrity_SSSRoundTrip(t *testing.T) {
	// Full SSS-style test: create polynomial, evaluate, interpolate
	for run := 0; run < 10; run++ {
		secret := uint8(run * 17)
		coeffs, err := NewPolynomial(secret, 2, rand.Reader)
		if err != nil {
			t.Fatalf("NewPolynomial failed: %v", err)
		}

		// Generate shares
		xs := []uint8{1, 2, 3}
		ys := make([]uint8, 3)
		for i, x := range xs {
			ys[i] = PolynomialEvaluate(coeffs, x)
		}

		// Recover
		recovered := InterpolatePolynomial(xs, ys, 0)
		if recovered != secret {
			t.Errorf("run %d: round-trip failed: got %#02x, expected %#02x", run, recovered, secret)
		}
	}
}

func TestNewPolynomial_NilReader(t *testing.T) {
	// Degree 0 doesn't need a reader
	coeffs, err := NewPolynomial(0x42, 0, nil)
	if err != nil {
		t.Fatalf("NewPolynomial with nil reader should work for degree 0: %v", err)
	}
	if coeffs[0] != 0x42 {
		t.Errorf("expected 0x42, got %#02x", coeffs[0])
	}
}

func TestNewPolynomial_GeneratesRandomValues(t *testing.T) {
	// Multiple calls should produce different polynomials (different random coefficients)
	intercepts := []uint8{0x00, 0x01, 0xFF}
	for _, intercept := range intercepts {
		results := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			coeffs, err := NewPolynomial(intercept, 3, rand.Reader)
			if err != nil {
				t.Fatalf("NewPolynomial failed: %v", err)
			}
			results[i] = coeffs
		}
		// At least some coefficients should differ between runs
		allSame := true
		for i := 1; i < len(results); i++ {
			if !bytes.Equal(results[0], results[i]) {
				allSame = false
				break
			}
		}
		if allSame && intercept != 0 {
			t.Errorf("polynomial generation should produce different random coefficients")
		}
	}
}

func TestHornerMethod_Efficiency(t *testing.T) {
	// High-degree polynomial evaluation using Horner's method
	coeffs := make(PolynomialCoefficients, 20)
	for i := range coeffs {
		coeffs[i] = uint8(i * 17)
	}

	// Should not panic
	result := PolynomialEvaluate(coeffs, 0xAB)
	_ = result
}

func TestInterpolatePolynomial_LargeSampleSet(t *testing.T) {
	// Interpolation with more samples than needed still works
	coeffs := PolynomialCoefficients{0x7F, 0x3A, 0xC5}
	xs := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ys := make([]uint8, len(xs))
	for i, x := range xs {
		ys[i] = PolynomialEvaluate(coeffs, x)
	}

	// Using all 10 points should still recover the intercept correctly
	result := InterpolatePolynomial(xs, ys, 0)
	if result != 0x7F {
		t.Errorf("large sample interpolation: got %#02x, expected 0x7F", result)
	}
}