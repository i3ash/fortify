package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/i3ash/fortify/files"
	"github.com/i3ash/fortify/fortifier"
	"github.com/i3ash/fortify/sss"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{Use: "fortify", Short: "Enhance file security through encryption"}
var ssss = &cobra.Command{Use: "sss", Short: "Shamir's secret sharing"}

func init() {
	root.AddCommand(ssss)
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version of the command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", VersionString())
		},
	})
}

func newFortifier(
	kind fortifier.CipherKeyKind, meta *fortifier.Metadata, args []string,
) (*fortifier.Fortifier, []string, error) {
	switch kind {
	case fortifier.CipherKeyKindSSS:
		if parts, err := sss.CombineKeyFiles(args); err != nil {
			return nil, args, err
		} else {
			return fortifier.NewFortifierWithSss(flagVerbose, flagTruncate, parts), args[len(parts):], nil
		}
	case fortifier.CipherKeyKindRSA:
		if kb, err := readKeyFile(args); err != nil {
			return nil, args, err
		} else {
			return fortifier.NewFortifierWithRsa(flagVerbose, meta, kb), args[1:], nil
		}
	default:
		return nil, args, fmt.Errorf("unknown cipher key kind: %s", kind)
	}
}

func readKeyFile(args []string) (kb []byte, err error) {
	size := len(args)
	if size == 0 {
		return
	}
	var kCloseFn func()
	var kf *os.File
	if kf, kCloseFn, err = files.OpenInputFile(args[0]); err != nil {
		return
	}
	defer kCloseFn()
	if kb, err = io.ReadAll(kf); err != nil {
		return
	}
	return
}

func VersionString() string {
	commit := fmt.Sprintf("%s", CommitHash)
	if commit == "-" {
		return fmt.Sprintf("%s", Version)
	}
	return fmt.Sprintf("%s (%s)", Version, CommitHash)
}
