package build

import (
	"fmt"
	"runtime/debug"
)

func PrintVersion() {
	fmt.Printf("%s\n", VersionString())
}

func PrintBuildInfo() {
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

func VersionString() string {
	commit := fmt.Sprintf("%s", CommitHash)
	if commit == "-" {
		return fmt.Sprintf("%s", Version)
	}
	return fmt.Sprintf("%s (%s)", Version, CommitHash)
}
