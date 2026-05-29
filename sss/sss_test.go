package sss

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitIntoShares_Basic(t *testing.T) {
	secret := []byte("Hello, Shamir Secret Sharing!")
	parts := uint8(5)
	threshold := uint8(3)

	shares, err := SplitIntoShares(secret, parts, threshold)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	if len(shares) != int(parts) {
		t.Fatalf("expected %d shares, got %d", parts, len(shares))
	}

	// Verify each share has correct length: len(secret) + 1 (for x coordinate)
	for i, share := range shares {
		if len(share) != len(secret)+1 {
			t.Errorf("share[%d] length = %d, expected %d", i, len(share), len(secret)+1)
		}
	}
}

func TestSplitIntoShares_Threshold2(t *testing.T) {
	secret := []byte("Test with threshold=2")
	shares, err := SplitIntoShares(secret, 3, 2)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	// Recover with just 2 shares (threshold)
	recovered, err := CombineFromShares(shares[:2])
	if err != nil {
		t.Fatalf("CombineFromShares failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch: got %q, expected %q", recovered, secret)
	}
}

func TestSplitIntoShares_MaxThreshold(t *testing.T) {
	secret := []byte("Max threshold = parts")
	parts := uint8(5)
	threshold := uint8(5)

	shares, err := SplitIntoShares(secret, parts, threshold)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	recovered, err := CombineFromShares(shares)
	if err != nil {
		t.Fatalf("CombineFromShares failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch: got %q, expected %q", recovered, secret)
	}
}

func TestSplitIntoShares_AnySubset(t *testing.T) {
	secret := []byte("Any 3 of 5 should work")
	parts := uint8(5)
	threshold := uint8(3)

	shares, err := SplitIntoShares(secret, parts, threshold)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	// Try different subsets
	subsets := [][]int{
		{0, 1, 2},
		{2, 3, 4},
		{0, 3, 4},
		{1, 2, 4},
	}

	for _, idx := range subsets {
		subset := make([]Share, len(idx))
		for i, j := range idx {
			subset[i] = shares[j]
		}
		recovered, err := CombineFromShares(subset)
		if err != nil {
			t.Fatalf("CombineFromShares failed with subset %v: %v", idx, err)
		}
		if !bytes.Equal(secret, recovered) {
			t.Errorf("subset %v: recovered secret mismatch", idx)
		}
	}
}

func TestSplitIntoShares_NotEnoughShares(t *testing.T) {
	secret := []byte("Need at least 3 shares")
	shares, err := SplitIntoShares(secret, 5, 3)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	// Try to recover with only 2 shares (below threshold).
	// The function will interpolate but produce a wrong result.
	recovered, err := CombineFromShares(shares[:2])
	if err != nil {
		t.Fatalf("CombineFromShares should not error, but produce wrong result: %v", err)
	}
	if bytes.Equal(secret, recovered) {
		t.Error("recovering with fewer shares than threshold should give wrong result")
	}
}

func TestSplitIntoShares_EmptySecret(t *testing.T) {
	_, err := SplitIntoShares([]byte{}, 3, 2)
	if err != ErrEmptySecret {
		t.Errorf("expected ErrEmptySecret, got %v", err)
	}
}

func TestSplitIntoShares_ThresholdTooSmall(t *testing.T) {
	_, err := SplitIntoShares([]byte("test"), 3, 1)
	if err != ErrThresholdTooSmall {
		t.Errorf("expected ErrThresholdTooSmall, got %v", err)
	}
}

func TestSplitIntoShares_ThresholdExceedsParts(t *testing.T) {
	_, err := SplitIntoShares([]byte("test"), 3, 5)
	if err != ErrInvalidPartsThreshold {
		t.Errorf("expected ErrInvalidPartsThreshold, got %v", err)
	}
}

func TestCombineFromShares_DuplicateShares(t *testing.T) {
	secret := []byte("No duplicates allowed")
	shares, err := SplitIntoShares(secret, 3, 2)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	// Pass the same share twice
	_, err = CombineFromShares([]Share{shares[0], shares[0]})
	if err != ErrDuplicatedShare {
		t.Errorf("expected ErrDuplicatedShare, got %v", err)
	}
}

func TestSplitIntoShares_LargeSecret(t *testing.T) {
	// Test with a larger secret (multiple blocks worth)
	secret := make([]byte, 1024)
	for i := range secret {
		secret[i] = byte(i & 0xFF)
	}

	shares, err := SplitIntoShares(secret, 3, 2)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	recovered, err := CombineFromShares(shares[:2])
	if err != nil {
		t.Fatalf("CombineFromShares failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Error("recovered secret mismatch for large secret")
	}
}

func TestSplit_Basic(t *testing.T) {
	secret := []byte("Integration test for Split()")
	parts := uint8(5)
	threshold := uint8(3)

	ps, err := Split(secret, parts, threshold)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if len(ps) != int(parts) {
		t.Fatalf("expected %d parts, got %d", parts, len(ps))
	}

	// Check that all parts have metadata
	for i, p := range ps {
		if p.Parts != parts {
			t.Errorf("part[%d].Parts = %d, expected %d", i, p.Parts, parts)
		}
		if p.Threshold != threshold {
			t.Errorf("part[%d].Threshold = %d, expected %d", i, p.Threshold, threshold)
		}
		if p.Payload == "" {
			t.Errorf("part[%d].Payload is empty", i)
		}
		if p.Digest == "" {
			t.Errorf("part[%d].Digest is empty", i)
		}
	}
}

func TestCombine_Basic(t *testing.T) {
	secret := []byte("Split then combine")
	ps, err := Split(secret, 5, 3)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	recovered, err := Combine(ps[:3])
	if err != nil {
		t.Fatalf("Combine failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch: got %q, expected %q", recovered, secret)
	}
}

func TestCombine_DigestMismatch(t *testing.T) {
	secret1 := []byte("Secret one")
	secret2 := []byte("Secret two")

	ps1, err := Split(secret1, 3, 2)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	ps2, err := Split(secret2, 3, 2)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Mix parts from different secrets
	mixed := []Part{ps1[0], ps2[1]}
	_, err = Combine(mixed)
	if err == nil {
		t.Error("expected error when combining parts with different digests")
	}
}

func TestCombineKeyFiles_Basic(t *testing.T) {
	secret := []byte("File-based split and combine")
	ps, err := Split(secret, 3, 2)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Write parts to temp files
	dir := t.TempDir()
	var paths []string
	for i, p := range ps[:2] {
		path := filepath.Join(dir, "key.json")
		// Each part gets its own file
		path = filepath.Join(dir, "key"+string(rune('0'+i))+".json")
		data, _ := json.Marshal(p)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("failed to write key file: %v", err)
		}
		paths = append(paths, path)
	}

	parts, err := CombineKeyFiles(paths)
	if err != nil {
		t.Fatalf("CombineKeyFiles failed: %v", err)
	}

	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}

	recovered, err := Combine(parts)
	if err != nil {
		t.Fatalf("Combine failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch")
	}
}

func TestCombineKeyFiles_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := CombineKeyFiles([]string{path})
	if err == nil {
		t.Error("expected error for invalid key file")
	}
}

func TestSplitIntoShares_Deterministic(t *testing.T) {
	// While the shares themselves are random, the combine should always
	// recover the original secret regardless of which subset we use
	secret := []byte{0x01, 0x02, 0x03, 0x04}

	for run := 0; run < 10; run++ {
		shares, err := SplitIntoShares(secret, 5, 3)
		if err != nil {
			t.Fatalf("SplitIntoShares failed: %v", err)
		}

		recovered, err := CombineFromShares(shares[:3])
		if err != nil {
			t.Fatalf("CombineFromShares failed: %v", err)
		}
		if !bytes.Equal(secret, recovered) {
			t.Errorf("run %d: recovered secret mismatch", run)
		}
	}
}

func TestGenerateSecureXCoordinates(t *testing.T) {
	xs, err := generateSecureXCoordinates(10)
	if err != nil {
		t.Fatalf("generateSecureXCoordinates failed: %v", err)
	}

	if len(xs) != 10 {
		t.Fatalf("expected 10 coordinates, got %d", len(xs))
	}

	// Check no duplicates
	seen := make(map[uint8]bool)
	for _, x := range xs {
		if seen[x] {
			t.Errorf("duplicate x coordinate: %d", x)
		}
		if x == 0 {
			t.Errorf("x coordinate should not be zero")
		}
		seen[x] = true
	}
}

func TestPayloadRoundTrip(t *testing.T) {
	secret := []byte("Check payload encoding/decoding")
	ps, err := Split(secret, 3, 2)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Decode payloads and reconstruct
	var shares []Share
	for _, p := range ps {
		share, err := base64.URLEncoding.DecodeString(p.Payload)
		if err != nil {
			t.Fatalf("base64 decode failed: %v", err)
		}
		shares = append(shares, share)
	}

	recovered, err := CombineFromShares(shares[:2])
	if err != nil {
		t.Fatalf("CombineFromShares failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch")
	}
}

func TestSplitIntoShares_SingleByteSecret(t *testing.T) {
	secret := []byte{'A'}
	shares, err := SplitIntoShares(secret, 3, 2)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	recovered, err := CombineFromShares(shares[:2])
	if err != nil {
		t.Fatalf("CombineFromShares failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch: got %q, expected %q", recovered, secret)
	}
}

func TestCombinePartFiles_EmptyOutNoPanic(t *testing.T) {
	// CombinePartFiles with empty out should not panic.
	// Before fix: defer oCloseFn() panics on nil function call.
	dir := t.TempDir()

	p1 := filepath.Join(dir, "a.json")
	os.WriteFile(p1, []byte(`{}`), 0644)
	p2 := filepath.Join(dir, "b.json")
	os.WriteFile(p2, []byte(`{}`), 0644)

	// This should NOT panic. If it does, the test crashes = FAIL.
	err := CombinePartFiles([]string{p1, p2}, "", false, false)
	if err == nil {
		t.Log("CombinePartFiles with empty out returned nil error (acceptable)")
	}
}

func TestSplitIntoShares_AllZerosSecret(t *testing.T) {
	secret := []byte{0, 0, 0, 0, 0}
	shares, err := SplitIntoShares(secret, 3, 2)
	if err != nil {
		t.Fatalf("SplitIntoShares failed: %v", err)
	}

	recovered, err := CombineFromShares(shares[:2])
	if err != nil {
		t.Fatalf("CombineFromShares failed: %v", err)
	}
	if !bytes.Equal(secret, recovered) {
		t.Errorf("recovered secret mismatch for zero secret")
	}
}

// TestSplitIntoFiles_ExactBlockMultiple proves the bug:
// When file size is an exact multiple of fileBlockSize,
// reader.Read returns (0, io.EOF) after the last full block.
// The current code does `if err != nil { return err }` before
// processing the data, causing io.EOF to be returned as an error.
func TestSplitIntoFiles_ExactBlockMultiple(t *testing.T) {
	defer CloseAllFilesForWrite()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.bin")
	prefix := filepath.Join(dir, "share_")

	// Create a file of exactly fileBlockSize bytes (non-zero data)
	data := make([]byte, fileBlockSize)
	for i := range data {
		data[i] = byte(i%251) + 1
	}
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// SplitIntoFiles should succeed without error
	err := SplitIntoFiles(inputPath, 3, 2, prefix, true, false)
	if err != nil {
		t.Fatalf("SplitIntoFiles returned error for file of exact block size: %v", err)
	}

	// Flush split files before reading them back
	CloseAllFilesForWrite()

	// Verify round-trip: combine split parts and compare
	var partFiles []string
	for i := 1; i <= 3; i++ {
		partFiles = append(partFiles, fmt.Sprintf("%s%dof%d.json", prefix, i, 3))
	}
	parts, err := CombineKeyFiles(partFiles)
	if err != nil {
		t.Fatalf("CombineKeyFiles failed: %v", err)
	}
	recovered, err := Combine(parts)
	if err != nil {
		t.Fatalf("Combine failed: %v", err)
	}
	if !bytes.Equal(data, recovered) {
		t.Errorf("recovered data length=%d, expected %d", len(recovered), len(data))
	}
}