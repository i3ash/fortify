package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestComputeDigest_Consistency(t *testing.T) {
	input := []byte("test data")
	d1 := ComputeDigest(input)
	d2 := ComputeDigest(input)

	if d1 != d2 {
		t.Errorf("digest should be deterministic: %q vs %q", d1, d2)
	}
}

func TestComputeDigest_DifferentInputs(t *testing.T) {
	d1 := ComputeDigest([]byte("hello"))
	d2 := ComputeDigest([]byte("world"))

	if d1 == d2 {
		t.Error("different inputs should produce different digests")
	}
}

func TestComputeDigest_Empty(t *testing.T) {
	d := ComputeDigest([]byte{})

	if d == "" {
		t.Error("digest of empty data should not be empty")
	}

	// Verify it's valid base64
	decoded, err := base64.URLEncoding.DecodeString(d)
	if err != nil {
		t.Errorf("digest should be valid base64: %v", err)
	}

	// SHA-512 produces 64 bytes, base64 encoded ~ 88 chars
	if len(decoded) != 64 {
		t.Errorf("SHA-512 digest should be 64 bytes, got %d", len(decoded))
	}
}

func TestComputeDigest_Base64Encoding(t *testing.T) {
	d := ComputeDigest([]byte("test"))

	// Should be URL-safe base64 (no '+' or '/' characters)
	if strings.Contains(d, "+") || strings.Contains(d, "/") {
		t.Error("digest should use URL-safe base64 encoding")
	}

	// Should not have padding issues
	if _, err := base64.URLEncoding.DecodeString(d); err != nil {
		t.Errorf("digest should be valid base64: %v", err)
	}
}

func TestComputeDigest_LargeInput(t *testing.T) {
	input := make([]byte, 1024*1024) // 1MB
	for i := range input {
		input[i] = byte(i & 0xFF)
	}

	d := ComputeDigest(input)
	if d == "" {
		t.Error("digest of large data should not be empty")
	}
}

func TestComputeDigest_UniquePerInput(t *testing.T) {
	inputs := [][]byte{
		[]byte(""),
		[]byte("a"),
		[]byte("ab"),
		[]byte("abc"),
		[]byte("hello"),
		[]byte("world"),
		[]byte("hello world"),
		[]byte("Hello World"),
	}

	digests := make(map[string]bool)
	for _, input := range inputs {
		d := ComputeDigest(input)
		if digests[d] {
			t.Errorf("duplicate digest for different input: %q", input)
		}
		digests[d] = true
	}
}

func TestComputeDigest_TrailingNull(t *testing.T) {
	d1 := ComputeDigest([]byte("test"))
	d2 := ComputeDigest([]byte("test\x00"))

	if d1 == d2 {
		t.Error("trailing null byte should produce different digest")
	}
}