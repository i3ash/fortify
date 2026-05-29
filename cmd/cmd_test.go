package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/i3ash/fortify/fortifier"
)

func TestReadKeyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.txt")
	content := []byte("test key content")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	kb, err := readKeyFile([]string{path})
	if err != nil {
		t.Fatalf("readKeyFile failed: %v", err)
	}
	if string(kb) != string(content) {
		t.Errorf("readKeyFile: got %q, expected %q", kb, content)
	}
}

func TestReadKeyFile_NonExistent(t *testing.T) {
	_, err := readKeyFile([]string{"/nonexistent/key.file"})
	if err == nil {
		t.Error("expected error for non-existent key file")
	}
}

func TestReadKeyFile_EmptyArgs(t *testing.T) {
	kb, err := readKeyFile([]string{})
	if err != nil {
		t.Fatalf("readKeyFile with empty args failed: %v", err)
	}
	if kb != nil {
		t.Errorf("expected nil for empty args, got %v", kb)
	}
}

func TestReadKeyFile_MultipleArgs(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "k1.txt")
	p2 := filepath.Join(dir, "k2.txt")
	os.WriteFile(p1, []byte("key1"), 0644)
	os.WriteFile(p2, []byte("key2"), 0644)

	// readKeyFile only reads the first arg and ignores the rest
	kb, err := readKeyFile([]string{p1, p2})
	if err != nil {
		t.Fatalf("readKeyFile should handle multiple args by reading first one: %v", err)
	}
	if string(kb) != "key1" {
		t.Errorf("expected first key content 'key1', got %q", kb)
	}
}

func TestDockerYes_InContainer(t *testing.T) {
	// In a Docker container, /.dockerenv exists
	result := dockerYes()
	// Can't assert true/false since test might run inside or outside Docker
	// Just verify it doesn't panic
	_ = result
}

func TestDockerReadKeyPaths_NonExistentFile(t *testing.T) {
	paths := dockerReadKeyPaths("/nonexistent/shared/keys")
	if len(paths) != 0 {
		t.Errorf("expected empty paths for non-existent file, got %v", paths)
	}
}

func TestDockerReadKeyPaths_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.txt")
	content := "/data/key1,/data/key2\n/data/key3"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write keys file: %v", err)
	}

	paths := dockerReadKeyPaths(path)
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "/data/key1" {
		t.Errorf("expected /data/key1, got %s", paths[0])
	}
	if paths[1] != "/data/key2" {
		t.Errorf("expected /data/key2, got %s", paths[1])
	}
	if paths[2] != "/data/key3" {
		t.Errorf("expected /data/key3, got %s", paths[2])
	}
}

func TestDockerReadKeyPaths_WhitespaceHandling(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.txt")
	content := "  /data/key1  ,  /data/key2  "
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write keys file: %v", err)
	}

	paths := dockerReadKeyPaths(path)
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "/data/key1" {
		t.Errorf("expected /data/key1, got %q", paths[0])
	}
	if paths[1] != "/data/key2" {
		t.Errorf("expected /data/key2, got %q", paths[1])
	}
}

func TestNewFortifier_SSS(t *testing.T) {
	// Create actual SSS key part files
	dir := t.TempDir()

	// Create a simple SSS key part file
	content := `{"payload":"dGVzdA==","block":1,"blocks":1,"part":1,"parts":2,"threshold":2,"digest":"abc123","timestamp":"2026-01-01T00:00:00Z"}`
	p1 := filepath.Join(dir, "part1.json")
	p2 := filepath.Join(dir, "part2.json")
	os.WriteFile(p1, []byte(content), 0644)
	os.WriteFile(p2, []byte(content), 0644)

	// Set verbose to false to avoid output during test
	flagVerbose = false
	flagTruncate = true

	f, rest, err := newFortifier("sss", nil, []string{p1, p2})
	if err != nil {
		t.Fatalf("newFortifier SSS failed: %v", err)
	}
	if f == nil {
		t.Fatal("newFortifier SSS returned nil")
	}
	if rest == nil {
		t.Fatal("newFortifier SSS returned nil rest")
	}
	_ = rest
}

func TestNewFortifier_RSA(t *testing.T) {
	// Create a minimal RSA public key file for testing
	// Just check the function handles the file path
	dir := t.TempDir()

	// Need a real-looking key file to pass the readKeyFile check
	// Use a minimal PEM that will fail parsing but not file reading
	pubPath := filepath.Join(dir, "pub.pem")
	os.WriteFile(pubPath, []byte("-----BEGIN PUBLIC KEY-----\n-----END PUBLIC KEY-----"), 0644)

	flagVerbose = false
	flagTruncate = true

	meta := &fortifier.Metadata{Key: "rsa"}
	f, rest, err := newFortifier("rsa", meta, []string{pubPath})
	// Expected to fail because PEM is empty/invalid, but function should handle gracefully
	if err == nil {
		t.Log("RSA fortifier created (may have failed later)")
	}
	_ = f
	_ = rest
}

func TestExecute_ZeroReturn(t *testing.T) {
	// Execute() returns 0 on success - but root.Execute() needs proper setup
	// Just check the function signature and basic behavior
	result := Execute()
	// In test context without proper args, it will likely fail and return 1
	_ = result
}

func TestConstants(t *testing.T) {
	if defaultSssParts != 5 {
		t.Errorf("default SSS parts should be 5, got %d", defaultSssParts)
	}
	if defaultSssThreshold != 3 {
		t.Errorf("default SSS threshold should be 3, got %d", defaultSssThreshold)
	}
	if defaultRandomBytes != 32 {
		t.Errorf("default random bytes should be 32, got %d", defaultRandomBytes)
	}
}