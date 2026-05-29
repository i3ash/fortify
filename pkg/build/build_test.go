package build

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionString(t *testing.T) {
	vs := VersionString()
	if !strings.Contains(vs, Version) {
		t.Errorf("VersionString should contain version %s, got %s", Version, vs)
	}
}

func TestNewVersionDetail(t *testing.T) {
	vd := NewVersionDetail()
	if vd == nil {
		t.Fatal("NewVersionDetail returned nil")
	}
	if vd.Version != Version {
		t.Errorf("expected version %s, got %s", Version, vd.Version)
	}
	if vd.CommitHash != CommitHash {
		t.Errorf("expected commit hash %s, got %s", CommitHash, vd.CommitHash)
	}
	if vd.BuildTime != Time {
		t.Errorf("expected build time %s, got %s", Time, vd.BuildTime)
	}
}

func TestNewVersionDetail_GoVersion(t *testing.T) {
	vd := NewVersionDetail()
	if vd.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if !strings.Contains(vd.GoVersion, "go") {
		t.Errorf("GoVersion should contain 'go', got %s", vd.GoVersion)
	}
}

func TestNewVersionDetail_GoModule(t *testing.T) {
	vd := NewVersionDetail()
	if vd.GoModule == nil {
		t.Error("GoModule should not be nil")
	}
	path, ok := vd.GoModule["path"]
	if ok && path != "github.com/i3ash/fortify" {
		t.Errorf("expected module path 'github.com/i3ash/fortify', got %s", path)
	}
}

func TestNewVersionDetail_JsonMarshal(t *testing.T) {
	vd := NewVersionDetail()
	data, err := json.Marshal(vd)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded VersionDetail
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Version != Version {
		t.Errorf("json round-trip version mismatch: got %s, expected %s", decoded.Version, Version)
	}
}

func TestPrintVersion(t *testing.T) {
	// Just verify it doesn't panic - will output to stdout
	PrintVersion()
}

func TestPrintVersionDetail(t *testing.T) {
	// Just verify it doesn't panic
	PrintVersionDetail()
}

func TestPrintJsonVersionDetail(t *testing.T) {
	// Just verify it doesn't panic
	PrintJsonVersionDetail()
}