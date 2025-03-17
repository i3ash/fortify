package build

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
)

func VersionString() string {
	commit := fmt.Sprintf("%s", CommitHash)
	if commit == "-" {
		return fmt.Sprintf("%s", Version)
	}
	return fmt.Sprintf("%s (%s)", Version, CommitHash)
}

func PrintVersion() {
	fmt.Printf("%s\n", VersionString())
}

type VersionDetail struct {
	Version         string            `json:"version"`
	CommitHash      string            `json:"commit_hash"`
	BuildTime       string            `json:"build_time"`
	GoModule        map[string]string `json:"go_module"`
	GoVersion       string            `json:"go_version"`
	GoBuildSettings map[string]string `json:"go_build_settings"`
}

func NewVersionDetail() *VersionDetail {
	var data = &VersionDetail{
		Version:         Version,
		CommitHash:      CommitHash,
		BuildTime:       Time,
		GoVersion:       runtime.Version(),
		GoModule:        make(map[string]string),
		GoBuildSettings: make(map[string]string),
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		data.GoVersion = info.GoVersion
		data.GoModule["path"] = info.Main.Path
		data.GoModule["version"] = info.Main.Version
		data.GoModule["sum"] = info.Main.Sum
		for _, kv := range info.Settings {
			data.GoBuildSettings[kv.Key] = kv.Value
		}
	}
	return data
}

func PrintVersionDetail() {
	fmt.Printf("Version    : %s\n", Version)
	fmt.Printf("Git Commit : %s\n", CommitHash)
	fmt.Printf("Build Time : %s\n", Time)
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("Go Version : %s\n", info.GoVersion)
		for _, kv := range info.Settings {
			fmt.Printf("Go Build Setting : %s=%s\n", kv.Key, kv.Value)
		}
	}
}

func PrintJsonVersionDetail() {
	jsonStr, _ := json.Marshal(NewVersionDetail())
	fmt.Printf("%s\n", jsonStr)
}
