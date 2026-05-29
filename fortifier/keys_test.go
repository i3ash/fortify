package fortifier

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/i3ash/fortify/sss"
)

func TestNewFortifierWithSss_CustomParts(t *testing.T) {
	// Create some SSS parts
	secret := []byte("custom-key-32-bytes!!")
	parts, err := sss.Split(secret, 3, 2)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	f := NewFortifierWithSss(false, true, parts)
	if f == nil {
		t.Fatal("NewFortifierWithSss returned nil")
	}
	if f.meta.Sss.Parts != 3 {
		t.Errorf("expected 3 parts, got %d", f.meta.Sss.Parts)
	}
	if f.meta.Sss.Threshold != 2 {
		t.Errorf("expected 2 threshold, got %d", f.meta.Sss.Threshold)
	}
	if f.meta.Sss.Digest != parts[0].Digest {
		t.Errorf("digest mismatch")
	}
}

func TestSetupSssKey_WithParts(t *testing.T) {
	// Generate a known secret and split it
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("failed to generate secret: %v", err)
	}

	parts, err := sss.Split(secret, 3, 2)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	f := NewFortifierWithSss(false, true, parts)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}

	// The key should be constructed from the parts
	if !bytes.Equal(f.key.raw, secret) {
		t.Errorf("key mismatch: got %x, expected %x", f.key.raw, secret)
	}

	// Block should be initialized
	if f.block == nil {
		t.Fatal("block should be initialized after SetupKey")
	}
}

func TestSetupSssKey_WithParts_NotEnoughShares(t *testing.T) {
	secret := []byte("test-secret-32-bytes-len!!")
	parts, err := sss.Split(secret, 3, 3) // threshold 3
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Only provide 2 parts (below threshold of 3)
	f := NewFortifierWithSss(false, true, parts[:2])
	err = f.SetupKey()
	// Should produce wrong key or error
	if err == nil && bytes.Equal(f.key.raw, secret) {
		t.Error("recovering with fewer shares than threshold should not produce correct key")
	}
}

func TestSetupSssKey_AutoGenerate(t *testing.T) {
	dir := t.TempDir()
	prevDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevDir)

	f := NewFortifierWithSss(false, true, nil)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}

	// Should have generated a 32-byte key
	if len(f.key.raw) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(f.key.raw))
	}

	// Should have created key files
	keyFiles, _ := filepath.Glob("fortified.key*.json")
	if len(keyFiles) == 0 {
		t.Error("expected key files to be created")
	}

	// Should have metadata
	if f.meta.Sss.Digest == "" {
		t.Error("SSS digest should be set")
	}
	if f.meta.Sss.Parts != 2 {
		t.Errorf("expected 2 parts, got %d", f.meta.Sss.Parts)
	}
}

func TestNewFortifierWithRsa_Basic(t *testing.T) {
	// RSA key bytes (just placeholder - the actual key setup needs a real key)
	f := NewFortifierWithRsa(false, nil, []byte("placeholder"))
	if f == nil {
		t.Fatal("NewFortifierWithRsa returned nil")
	}
	if f.key.kind != CipherKeyKindRSA {
		t.Errorf("expected RSA key kind, got %s", f.key.kind)
	}
}

func TestCipherKeyData_NewSha256_Consistency(t *testing.T) {
	key1 := &CipherKeyData{raw: []byte("test-key-32-bytes-long-for-testing!!")}
	key2 := &CipherKeyData{raw: []byte("test-key-32-bytes-long-for-testing!!")}

	h1 := key1.NewSha256()
	h2 := key2.NewSha256()

	sum1 := h1.Sum(nil)
	sum2 := h2.Sum(nil)

	if !bytes.Equal(sum1, sum2) {
		t.Error("SHA-256 should be deterministic for same key")
	}

	// Different keys should produce different hashes
	key3 := &CipherKeyData{raw: []byte("different-key-32-bytes-for-testing!!")}
	h3 := key3.NewSha256()
	sum3 := h3.Sum(nil)

	if bytes.Equal(sum1, sum3) {
		t.Error("different keys should produce different hashes")
	}
}

func TestCipherKeyKind_String(t *testing.T) {
	if CipherKeyKindSSS.String() != "sss" {
		t.Errorf("SSS string: got %s, expected sss", CipherKeyKindSSS.String())
	}
	if CipherKeyKindRSA.String() != "rsa" {
		t.Errorf("RSA string: got %s, expected rsa", CipherKeyKindRSA.String())
	}
}

func TestCipherModeName_String(t *testing.T) {
	if CipherModeAes256CTR.String() != "aes256-ctr" {
		t.Errorf("CTR string: got %s", CipherModeAes256CTR.String())
	}
	if CipherModeAes256CFB.String() != "aes256-cfb" {
		t.Errorf("CFB string: got %s", CipherModeAes256CFB.String())
	}
	if CipherModeAes256OFB.String() != "aes256-ofb" {
		t.Errorf("OFB string: got %s", CipherModeAes256OFB.String())
	}
}

func TestCipherMode_NewEncrypter(t *testing.T) {
	f := NewFortifierWithSss(false, true, nil)

	modes := []CipherModeName{CipherModeAes256CTR, CipherModeAes256CFB, CipherModeAes256OFB}
	for _, mode := range modes {
		enc := NewEncrypter(mode, f)
		if enc == nil {
			t.Errorf("NewEncrypter(%s) returned nil", mode)
		}
	}

	// Unknown mode should return nil
	unknown := NewEncrypter("unknown", f)
	if unknown != nil {
		t.Error("NewEncrypter with unknown mode should return nil")
	}
}

func TestCipherMode_NewDecrypter(t *testing.T) {
	f := NewFortifierWithSss(false, true, nil)

	modes := []CipherModeName{CipherModeAes256CTR, CipherModeAes256CFB, CipherModeAes256OFB}
	for _, mode := range modes {
		dec := NewDecrypter(mode, f)
		if dec == nil {
			t.Errorf("NewDecrypter(%s) returned nil", mode)
		}
	}

	// Unknown mode should return nil
	unknown := NewDecrypter("unknown", f)
	if unknown != nil {
		t.Error("NewDecrypter with unknown mode should return nil")
	}
}

func TestAes256StreamEncrypter_EncryptWithKey(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i)
	}

	dir := t.TempDir()
	plainPath := filepath.Join(dir, "plain.txt")
	encPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")
	plaintext := []byte("Hello with known key!")

	if err := os.WriteFile(plainPath, plaintext, 0644); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	block, err := aes.NewCipher(rawKey)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	f := &Fortifier{
		meta:  &Metadata{Key: CipherKeyKindSSS, Mode: CipherModeAes256CTR},
		key:   &CipherKeyData{kind: CipherKeyKindSSS, raw: rawKey},
		block: block,
	}

	enc := NewEncrypter(CipherModeAes256CTR, f)
	if enc == nil {
		t.Fatal("NewEncrypter returned nil")
	}

	in, err := os.Open(plainPath)
	if err != nil {
		t.Fatalf("failed to open input: %v", err)
	}
	defer in.Close()

	out, err := os.Create(encPath)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	if err := enc.EncryptFile(in, out); err != nil {
		out.Close()
		t.Fatalf("EncryptFile failed: %v", err)
	}
	out.Close()

	// Decrypt
	encIn, err := os.Open(encPath)
	if err != nil {
		t.Fatalf("failed to open encrypted: %v", err)
	}
	defer encIn.Close()

	layout := &FileLayout{}
	if err := layout.ReadHeadIn(encIn); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	f2 := &Fortifier{
		meta:  layout.Metadata(),
		key:   &CipherKeyData{kind: CipherKeyKindSSS, raw: rawKey},
		block: block,
	}

	dec := NewDecrypter(CipherModeAes256CTR, f2)
	decOut, err := os.Create(decPath)
	if err != nil {
		t.Fatalf("failed to create decrypted: %v", err)
	}
	defer decOut.Close()

	if err := dec.DecryptFile(encIn, decOut, layout); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}
	decOut.Close()

	recovered, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read decrypted: %v", err)
	}

	if !bytes.Equal(plaintext, recovered) {
		t.Errorf("recovered text mismatch: got %q, expected %q", recovered, plaintext)
	}
}

func TestFileLayout_WriteHeadOut_NoOutput(t *testing.T) {
	layout := &FileLayout{metadata: &Metadata{Key: CipherKeyKindSSS, Mode: CipherModeAes256CTR}}
	if err := layout.WriteHeadOut(nil); err != nil {
		t.Fatalf("WriteHeadOut with nil output should not error: %v", err)
	}
	// Metadata should still be marshaled
	if len(layout.metadataRaw) == 0 {
		t.Error("metadata raw should be set even with nil output")
	}
}

func TestFileLayout_WriteHeadPlaceHolders_NoOutput(t *testing.T) {
	layout := &FileLayout{metadata: &Metadata{Key: CipherKeyKindSSS, Mode: CipherModeAes256CTR}}
	key := &CipherKeyData{raw: make([]byte, 32)}
	if _, err := rand.Read(key.raw); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	check := key.NewSha256()

	if err := layout.WriteHeadPlaceHolders(nil, key, check, 100); err != nil {
		t.Fatalf("WriteHeadPlaceHolders with nil output should not error: %v", err)
	}
	if layout.dataLength != 100 {
		t.Errorf("expected data length 100, got %d", layout.dataLength)
	}
}

func TestEncryptDecrypt_FileRoundTrip(t *testing.T) {
	// Full round-trip using files (which support io.WriteSeeker)
	plaintext := []byte("File-based round-trip test for write-seeker!")
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	block, err := aes.NewCipher(rawKey)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	encPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	f := NewFortifierWithSss(false, true, nil)
	f.key.raw = rawKey
	f.block = block
	f.meta = &Metadata{Key: CipherKeyKindSSS, Mode: CipherModeAes256CTR}

	enc := NewEncrypter(CipherModeAes256CTR, f)

	in, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("failed to open input: %v", err)
	}
	defer in.Close()

	out, err := os.Create(encPath)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	if err := enc.EncryptFile(in, out); err != nil {
		out.Close()
		t.Fatalf("EncryptFile failed: %v", err)
	}
	out.Close()

	// Decrypt
	encIn, err := os.Open(encPath)
	if err != nil {
		t.Fatalf("failed to open encrypted: %v", err)
	}
	defer encIn.Close()

	layout := &FileLayout{}
	if err := layout.ReadHeadIn(encIn); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	f2 := NewFortifierWithSss(false, true, nil)
	f2.key.raw = rawKey
	f2.block = block
	f2.meta = layout.Metadata()

	dec := NewDecrypter(CipherModeAes256CTR, f2)
	decOut, err := os.Create(decPath)
	if err != nil {
		t.Fatalf("failed to create decrypted: %v", err)
	}
	defer decOut.Close()

	if err := dec.DecryptFile(encIn, decOut, layout); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}
	decOut.Close()

	recovered, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read decrypted: %v", err)
	}

	if !bytes.Equal(plaintext, recovered) {
		t.Errorf("file round-trip failed")
	}
}