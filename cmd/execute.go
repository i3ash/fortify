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

const tempFilePrefix = ".oO0"
const mountBinDir = "/usr/local/sbin"
const keyListFile = "/dev/shm/keys/k_fortify"

var cleanupOnce sync.Once
var cleanupDelaySeconds = 5

func init() {
	c := &cobra.Command{
		Short: "Execute a decrypted program from the fortified file",
		Use:   "execute -i <input-file> [flags] <key1> [key2] ... [-- [arg1] [arg2] ...]",
		Args:  cobra.MinimumNArgs(0),
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
	docker := dockerYes()
	merge := args
	if docker {
		paths := dockerReadKeyPaths(keyListFile)
		merge = make([]string, 0, len(args)+len(paths))
		merge = append(merge, paths...)
		merge = append(merge, args...)
	}
	meta := layout.Metadata()
	var f *fortifier.Fortifier
	var rest []string
	if f, rest, err = newFortifier(meta.Key, meta, merge); err != nil {
		return
	}
	var dec fortifier.Decrypter
	if dec = fortifier.NewDecrypter(meta.Mode, f); dec == nil {
		err = fmt.Errorf("unknown cipher mode name: %s", meta.Mode)
		return
	}
	var out *os.File
	var command string
	if docker {
		command = filepath.Base(in.Name())
		out, err = os.Create(mountBinDir + "/" + command)
	} else {
		if out, err = os.CreateTemp("", tempFilePrefix); err == nil {
			command = out.Name()
		}
	}
	defer func() { cleanupOnce.Do(func() { cleanup(out) }) }()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
		return nil
	}
	r := bufio.NewReaderSize(in, 128*1024)
	err = dec.Decrypt(r, out, layout)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to decrypt program: %v\n", err)
		cleanupOnce.Do(func() { cleanup(out) })
		return nil
	}
	path := out.Name()
	_ = out.Close()
	if err = clean(out.Name()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to setup cleaning: %v\n", err)
	}
	if err = permit(path); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to permit: %v\n", err)
	}
	argv := append([]string{command}, rest...)
	if err = syscall.Exec(command, argv, os.Environ()); err == nil {
		return nil
	}
	var wg sync.WaitGroup
	var process *os.Process
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt, syscall.SIGTERM)
	if process, err = start(command, out, &wg, chanSignal, rest...); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to run program: %v\n", err)
		cleanupOnce.Do(func() { cleanup(out) })
		return nil
	}
	sig := <-chanSignal
	_ = process.Signal(sig)
	wg.Wait()
	return
}

func clean(path string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("sleep %d && rm '%s'", cleanupDelaySeconds, path))
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func permit(path string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("chmod u+x '%s'", path))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chmod failed: %v", err)
	}
	return nil
}

func start(command string, out *os.File, wg *sync.WaitGroup, chanSignal chan os.Signal, arg ...string) (*os.Process, error) {
	cmd := exec.Command(command, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return cmd.Process, fmt.Errorf("failed to start program: %v", err)
	}
	select {
	case <-time.After(time.Duration(cleanupDelaySeconds) * time.Second):
		cleanupOnce.Do(func() { cleanup(out) })
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

func cleanup(out *os.File) {
	programPath := out.Name()
	if programPath == "" {
		return
	}
	if _, err := os.Stat(programPath); os.IsNotExist(err) {
		return
	}
	if err := os.Remove(programPath); err != nil {
		return
	}
}

func dockerYes() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func dockerReadKeyPaths(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{}
	}
	defer func() { _ = file.Close() }()
	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		elements := strings.Split(line, ",")
		for _, element := range elements {
			s := strings.TrimSpace(element)
			if s != "" {
				result = append(result, s)
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return []string{}
	}
	return result
}
