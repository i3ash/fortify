package fortifier

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestFileLayout_WriteReadRoundTrip(t *testing.T) {
	// Create a layout, write it, read it back, compare
	layout := &FileLayout{
		metadata: &Metadata{
			Key:  CipherKeyKindSSS,
			Mode: CipherModeAes256CTR,
		},
	}

	var buf bytes.Buffer
	if err := layout.WriteHeadOut(&buf); err != nil {
		t.Fatalf("WriteHeadOut failed: %v", err)
	}

	// Read back
	layout2 := &FileLayout{}
	reader := bytes.NewReader(buf.Bytes())
	if err := layout2.ReadHeadIn(reader); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	// Verify basic fields
	if layout.magic != layout2.magic {
		t.Errorf("magic mismatch: %X vs %X", layout.magic, layout2.magic)
	}
	if layout.metadata.Key != layout2.metadata.Key {
		t.Errorf("key kind mismatch: %s vs %s", layout.metadata.Key, layout2.metadata.Key)
	}
	if layout.metadata.Mode != layout2.metadata.Mode {
		t.Errorf("mode mismatch: %s vs %s", layout.metadata.Mode, layout2.metadata.Mode)
	}

	// Verify data start mark
	if !bytes.Equal(layout.dataStartMark, layout2.dataStartMark) {
		t.Errorf("data start mark mismatch")
	}

	// Verify nonce is present and non-zero
	if len(layout2.nonce) != 8 {
		t.Errorf("nonce should be 8 bytes, got %d", len(layout2.nonce))
	}
	// Nonce should be random (check it's not all zeros)
	var nonceSum byte
	for _, b := range layout2.nonce {
		nonceSum |= b
	}
	if nonceSum == 0 {
		t.Error("nonce should not be all zeros")
	}
}

func TestFileLayout_MagicNumber(t *testing.T) {
	layout := &FileLayout{
		metadata: &Metadata{Key: CipherKeyKindSSS, Mode: CipherModeAes256CTR},
	}

	var buf bytes.Buffer
	if err := layout.WriteHeadOut(&buf); err != nil {
		t.Fatalf("WriteHeadOut failed: %v", err)
	}

	// Check magic number
	var magic uint32
	if err := binary.Read(bytes.NewReader(buf.Bytes()), binary.BigEndian, &magic); err != nil {
		t.Fatalf("failed to read magic: %v", err)
	}

	// Magic should have 0x40F1ED00 as base with version in low byte
	if magic&0x7FFFFF00 != FileMagicNumber {
		t.Errorf("magic base mismatch: %X, expected base %X", magic, FileMagicNumber)
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	plaintext := []byte("Hello, fortify encryption test!")

	// Create a temporary input file
	dir := t.TempDir()
	inPath := filepath.Join(dir, "plain.txt")
	outPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Open files
	in, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("failed to open input: %v", err)
	}
	defer in.Close()

	// Encrypt with SSS
	f := NewFortifierWithSss(false, true, nil)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}

	rawKey := make([]byte, len(f.key.raw))
	copy(rawKey, f.key.raw)

	// Write the generated key parts to files
	enc := NewEncrypter(CipherModeAes256CTR, f)
	if enc == nil {
		t.Fatal("NewEncrypter returned nil")
	}

	out, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	if err := enc.EncryptFile(in, out); err != nil {
		out.Close()
		t.Fatalf("EncryptFile failed: %v", err)
	}
	out.Close()

	// Verify encrypted file exists and is non-empty
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("failed to stat encrypted file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("encrypted file is empty")
	}

	// Now decrypt
	encIn, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("failed to open encrypted file: %v", err)
	}
	defer encIn.Close()

	// Read layout from encrypted file
	layout := &FileLayout{}
	if err := layout.ReadHeadIn(encIn); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	// Create decryptor with the raw key we saved.
	// Since raw key already set, we need to init block manually
	// (SetupKey returns early if raw is set).
	f2 := NewFortifierWithSss(false, true, nil)
	f2.key.raw = rawKey
	if f2.block, err = aes.NewCipher(rawKey); err != nil {
		t.Fatalf("failed to create cipher block: %v", err)
	}
	f2.meta = layout.Metadata()

	dec := NewDecrypter(CipherModeAes256CTR, f2)
	if dec == nil {
		t.Fatal("NewDecrypter returned nil")
	}

	decOut, err := os.Create(decPath)
	if err != nil {
		t.Fatalf("failed to create decrypted file: %v", err)
	}
	defer decOut.Close()

	if err := dec.DecryptFile(encIn, decOut, layout); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}
	decOut.Close()

	// Verify decrypted content matches original
	decrypted, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted content mismatch:\n  got:      %q\n  expected: %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_AllModes(t *testing.T) {
	plaintext := []byte("Testing all cipher modes with a longer message that has some padding to work with.")

	modes := []struct {
		name  CipherModeName
		maker func(block cipher.Block, iv []byte) cipher.Stream
	}{
		{CipherModeAes256CTR, cipher.NewCTR},
		{CipherModeAes256CFB, cipher.NewCFBEncrypter},
		{CipherModeAes256OFB, cipher.NewOFB},
	}

	for _, mode := range modes {
		t.Run(string(mode.name), func(t *testing.T) {
			dir := t.TempDir()
			inPath := filepath.Join(dir, "plain.txt")
			outPath := filepath.Join(dir, "encrypted.bin")
			decPath := filepath.Join(dir, "decrypted.txt")

			if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
				t.Fatalf("failed to write input: %v", err)
			}

			in, err := os.Open(inPath)
			if err != nil {
				t.Fatalf("failed to open input: %v", err)
			}
			defer in.Close()

			f := NewFortifierWithSss(false, true, nil)
			if err := f.SetupKey(); err != nil {
				t.Fatalf("SetupKey failed: %v", err)
			}

			rawKey := make([]byte, len(f.key.raw))
			copy(rawKey, f.key.raw)

			// Create encrypter
			var enc Encrypter
			switch mode.name {
			case CipherModeAes256CTR:
				enc = NewAes256EncrypterCTR(f)
			case CipherModeAes256CFB:
				enc = NewAes256EncrypterCFB(f)
			case CipherModeAes256OFB:
				enc = NewAes256EncrypterOFB(f)
			}
			if enc == nil {
				t.Fatal("NewEncrypter returned nil")
			}

			out, err := os.Create(outPath)
			if err != nil {
				t.Fatalf("failed to create output: %v", err)
			}

			if err := enc.EncryptFile(in, out); err != nil {
				out.Close()
				t.Fatalf("EncryptFile failed: %v", err)
			}
			out.Close()

			// Decrypt
			encIn, err := os.Open(outPath)
			if err != nil {
				t.Fatalf("failed to open encrypted file: %v", err)
			}
			defer encIn.Close()

			layout := &FileLayout{}
			if err := layout.ReadHeadIn(encIn); err != nil {
				t.Fatalf("ReadHeadIn failed: %v", err)
			}

			f2 := NewFortifierWithSss(false, true, nil)
			f2.key.raw = rawKey
			if f2.block, err = aes.NewCipher(rawKey); err != nil {
				t.Fatalf("failed to create cipher block: %v", err)
			}
			f2.meta = layout.Metadata()

			var dec Decrypter
			switch mode.name {
			case CipherModeAes256CTR:
				dec = NewAes256DecrypterCTR(f2)
			case CipherModeAes256CFB:
				dec = NewAes256DecrypterCFB(f2)
			case CipherModeAes256OFB:
				dec = NewAes256DecrypterOFB(f2)
			}
			if dec == nil {
				t.Fatal("NewDecrypter returned nil")
			}

			decOut, err := os.Create(decPath)
			if err != nil {
				t.Fatalf("failed to create decrypted file: %v", err)
			}
			defer decOut.Close()

			if err := dec.DecryptFile(encIn, decOut, layout); err != nil {
				t.Fatalf("DecryptFile for mode %s failed: %v", mode.name, err)
			}
			decOut.Close()

			decrypted, err := os.ReadFile(decPath)
			if err != nil {
				t.Fatalf("failed to read decrypted file: %v", err)
			}

			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("mode %s: decrypted content mismatch", mode.name)
			}
		})
	}
}

func TestEncryptDecrypt_EmptyFile(t *testing.T) {
	plaintext := []byte{}

	dir := t.TempDir()
	inPath := filepath.Join(dir, "empty.txt")
	outPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	in, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("failed to open input: %v", err)
	}
	defer in.Close()

	f := NewFortifierWithSss(false, true, nil)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}
	rawKey := make([]byte, len(f.key.raw))
	copy(rawKey, f.key.raw)

	enc := NewEncrypter(CipherModeAes256CTR, f)
	out, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	if err := enc.EncryptFile(in, out); err != nil {
		out.Close()
		t.Fatalf("EncryptFile failed: %v", err)
	}
	out.Close()

	// Decrypt back
	encIn, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("failed to open encrypted file: %v", err)
	}
	defer encIn.Close()

	layout := &FileLayout{}
	if err := layout.ReadHeadIn(encIn); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	f2 := NewFortifierWithSss(false, true, nil)
	f2.key.raw = rawKey
	if f2.block, err = aes.NewCipher(rawKey); err != nil {
		t.Fatalf("failed to create cipher block: %v", err)
	}
	f2.meta = layout.Metadata()

	dec := NewDecrypter(CipherModeAes256CTR, f2)
	decOut, err := os.Create(decPath)
	if err != nil {
		t.Fatalf("failed to create decrypted file: %v", err)
	}
	defer decOut.Close()

	if err := dec.DecryptFile(encIn, decOut, layout); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}
	decOut.Close()

	decrypted, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("empty file round-trip failed")
	}
}

func TestEncryptDecrypt_LargeFile(t *testing.T) {
	// 64KB of data
	plaintext := make([]byte, 64*1024)
	for i := range plaintext {
		plaintext[i] = byte(i & 0xFF)
	}

	dir := t.TempDir()
	inPath := filepath.Join(dir, "large.txt")
	outPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	in, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("failed to open input: %v", err)
	}
	defer in.Close()

	f := NewFortifierWithSss(false, true, nil)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}
	rawKey := make([]byte, len(f.key.raw))
	copy(rawKey, f.key.raw)

	enc := NewEncrypter(CipherModeAes256CTR, f)
	out, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	if err := enc.EncryptFile(in, out); err != nil {
		out.Close()
		t.Fatalf("EncryptFile failed: %v", err)
	}
	out.Close()

	// Decrypt
	encIn, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("failed to open encrypted file: %v", err)
	}
	defer encIn.Close()

	layout := &FileLayout{}
	if err := layout.ReadHeadIn(encIn); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	f2 := NewFortifierWithSss(false, true, nil)
	f2.key.raw = rawKey
	if f2.block, err = aes.NewCipher(rawKey); err != nil {
		t.Fatalf("failed to create cipher block: %v", err)
	}
	f2.meta = layout.Metadata()

	dec := NewDecrypter(CipherModeAes256CTR, f2)
	decOut, err := os.Create(decPath)
	if err != nil {
		t.Fatalf("failed to create decrypted file: %v", err)
	}
	defer decOut.Close()

	if err := dec.DecryptFile(encIn, decOut, layout); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}
	decOut.Close()

	decrypted, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("large file round-trip failed")
	}
}

func TestFileLayout_InvalidMagic(t *testing.T) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 0xDEADBEEF)

	layout := &FileLayout{}
	err := layout.ReadHeadIn(bytes.NewReader(buf))
	if err == nil {
		t.Error("expected error for invalid magic number")
	}
}

func TestCipherKeyData_NewSha256(t *testing.T) {
	key := &CipherKeyData{raw: []byte("test-key-32-bytes-long-for-testing")}
	h := key.NewSha256()
	if h == nil {
		t.Fatal("NewSha256 returned nil")
	}
	sum := h.Sum(nil)
	if len(sum) != 32 {
		t.Errorf("SHA-256 hash should be 32 bytes, got %d", len(sum))
	}
}

func TestAesKeyGeneration(t *testing.T) {
	// Verify generated AES-256 keys are 32 bytes
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("AES-256 key should be 32 bytes, got %d", len(key))
	}
}

func TestNewFortifierWithSss(t *testing.T) {
	f := NewFortifierWithSss(false, true, nil)
	if f == nil {
		t.Fatal("NewFortifierWithSss returned nil")
	}
	if f.meta == nil {
		t.Fatal("meta is nil")
	}
	if f.meta.Sss == nil {
		t.Fatal("SSS metadata is nil")
	}
	if f.meta.Sss.Parts != 2 {
		t.Errorf("expected default 2 parts, got %d", f.meta.Sss.Parts)
	}
	if f.meta.Sss.Threshold != 2 {
		t.Errorf("expected default 2 threshold, got %d", f.meta.Sss.Threshold)
	}
	if f.key.kind != CipherKeyKindSSS {
		t.Errorf("expected SSS key kind, got %s", f.key.kind)
	}
}

func TestFortifierSetupKey(t *testing.T) {
	f := NewFortifierWithSss(false, true, nil)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}

	if len(f.key.raw) != 32 {
		t.Errorf("expected 32-byte AES key, got %d bytes", len(f.key.raw))
	}

	if f.meta.Key != CipherKeyKindSSS {
		t.Errorf("expected key kind SSS, got %s", f.meta.Key)
	}

	if f.meta.Sss.Digest == "" {
		t.Error("SSS digest should not be empty after SetupKey")
	}
}

func TestEncryptDecrypt_WrongKey(t *testing.T) {
	plaintext := []byte("Test wrong key detection")

	dir := t.TempDir()
	inPath := filepath.Join(dir, "plain.txt")
	outPath := filepath.Join(dir, "encrypted.bin")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatalf("failed to write input: %v", err)
	}

	in, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("failed to open input: %v", err)
	}
	defer in.Close()

	f := NewFortifierWithSss(false, true, nil)
	if err := f.SetupKey(); err != nil {
		t.Fatalf("SetupKey failed: %v", err)
	}

	enc := NewEncrypter(CipherModeAes256CTR, f)
	out, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	if err := enc.EncryptFile(in, out); err != nil {
		out.Close()
		t.Fatalf("EncryptFile failed: %v", err)
	}
	out.Close()

	// Try to decrypt with a different key
	encIn, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("failed to open encrypted file: %v", err)
	}
	defer encIn.Close()

	layout := &FileLayout{}
	if err := layout.ReadHeadIn(encIn); err != nil {
		t.Fatalf("ReadHeadIn failed: %v", err)
	}

	wrongKey := make([]byte, 32)
	if _, err := rand.Read(wrongKey); err != nil {
		t.Fatalf("failed to generate wrong key: %v", err)
	}

	f2 := NewFortifierWithSss(false, true, nil)
	f2.key.raw = wrongKey
	if f2.block, err = aes.NewCipher(wrongKey); err != nil {
		t.Fatalf("failed to create cipher block: %v", err)
	}
	f2.meta = layout.Metadata()

	dec := NewDecrypter(CipherModeAes256CTR, f2)
	decOut, err := os.Create(filepath.Join(dir, "decrypted.txt"))
	if err != nil {
		t.Fatalf("failed to create decrypted file: %v", err)
	}
	defer decOut.Close()

	err = dec.DecryptFile(encIn, decOut, layout)
	if err == nil {
		// Decryption with wrong key should produce garbage but not error
		// The checksum should catch it
		t.Error("expected error when decrypting with wrong key")
	}
}