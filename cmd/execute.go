package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/i3ash/fortify/files"
	"github.com/i3ash/fortify/fortifier"
	"github.com/spf13/cobra"
)

var cleanupOnce sync.Once
var cleanupDelaySeconds = 5

func init() {
	c := &cobra.Command{
		Short: "Execute a decrypted program from the fortified file",
		Use:   "execute -i <input-file> [flags] <key1> [key2] ... [-- [arg1] [arg2] ...]",
		Args:  cobra.MinimumNArgs(1),
		RunE:  func(_ *cobra.Command, args []string) error { return execute(flagIn, args) },
	}
	c.SetUsageTemplate(fmt.Sprintf(`%s
Required Arguments:
  <key1>   Path to the first secret share file or private key file if cipher key kind of <input-file> is 'rsa'
  [key2]   [Required cipher key kind of <input-file> is 'sss'] Path to the second secret share file
  ...      Additional paths to secret share files (all files remain unmodified)
`, c.UsageTemplate()))
	root.AddCommand(c)
	initFlagHelp(c)
	initFlagVerbose(c)
	initFlagIn(c, "[Required] Path of the fortified/encrypted input file")
	_ = c.MarkFlagRequired("in")
	c.Flags().IntVarP(&cleanupDelaySeconds, "cleanup-delay", "", 5,
		"Number of seconds to wait before performing the cleanup operation")
	if cleanupDelaySeconds < 1 {
		cleanupDelaySeconds = 1
	}
}

func execute(input string, args []string) (err error) {
	files.SetVerbose(flagVerbose)
	var in *os.File
	var iCloseFn func()
	if in, iCloseFn, err = files.OpenInputFile(input); err != nil {
		return
	}
	defer iCloseFn()
	layout := &fortifier.FileLayout{}
	if err = layout.ReadHeadIn(in); err != nil {
		return
	}
	//fmt.Printf("%s\n", layout)
	meta := layout.Metadata()
	var f *fortifier.Fortifier
	var rest []string
	if f, rest, err = newFortifier(meta.Key, meta, args); err != nil {
		return
	}
	var dec fortifier.Decrypter
	if dec = fortifier.NewDecrypter(meta.Mode, f); dec == nil {
		err = fmt.Errorf("unknown cipher mode name: %s", meta.Mode)
		return
	}
	var out *os.File
	out, err = os.CreateTemp("", ".bin")
	defer cleanupOnce.Do(func() { cleanup(out.Name()) })
	go func() {
		select {
		case <-time.After(time.Duration(cleanupDelaySeconds) * time.Second):
			cleanupOnce.Do(func() { cleanup(out.Name()) })
		}
	}()
	r := bufio.NewReaderSize(in, 128*1024)
	err = dec.Decrypt(r, out, layout)
	_ = out.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to decrypt program: %v\n", err)
		os.Exit(1)
		return nil
	}
	var wg sync.WaitGroup
	var process *os.Process
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt, syscall.SIGTERM)
	if process, err = start(out.Name(), &wg, chanSignal, rest...); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to run program: %v\n", err)
		os.Exit(2)
		return nil
	}
	defer func() { fmt.Println("Executed") }()
	sig := <-chanSignal
	_ = process.Signal(sig)
	wg.Wait()
	return
}

func permit(path string) error {
	cmd := exec.Command("/bin/sh", "-c", "chmod u+x "+path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chmod failed: %v", err)
	}
	return nil
}

func start(path string, wg *sync.WaitGroup, chanSignal chan os.Signal, arg ...string) (*os.Process, error) {
	if err := permit(path); err != nil {
		fmt.Printf("failed to add permission: %v\n", err)
		return nil, err
	}
	cmd := exec.Command(path, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return cmd.Process, fmt.Errorf("failed to start program: %v", err)
	}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		if err := cmd.Wait(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		}
		wg.Done()
		chanSignal <- syscall.SIGTERM
	}(wg)
	return cmd.Process, nil
}

func cleanup(programPath string) {
	programName := filepath.Base(programPath)
	if !strings.HasPrefix(programName, ".bin") {
		return
	}
	if programPath == "" {
		return
	}
	if _, err := os.Stat(programPath); os.IsNotExist(err) {
		return
	}
	_ = os.Remove(programPath)
}
