package sss

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/i3ash/fortify/files"
	"github.com/i3ash/fortify/pkg/gf256"
	"github.com/i3ash/fortify/utils"
)

func Combine(parts []Part) ([]byte, error) {
	var (
		secret []byte
		expect string
	)
	shares := make([]Share, len(parts))
	for index, i := range parts {
		if share, err := base64.URLEncoding.DecodeString(i.Payload); err != nil {
			return secret, err
		} else {
			shares[index] = share
			if len(expect) == 0 {
				expect = i.Digest
			} else {
				if expect != i.Digest {
					fmt.Printf("Expect secret digest: %s\n", expect)
					fmt.Printf("Actual secret digest: %s\n", i.Digest)
					return secret, fmt.Errorf("secret digest mismatch in file %v", index+1)
				}
			}
		}
	}
	var err error
	if secret, err = CombineFromShares(shares); err != nil {
		return secret, err
	}
	return secret, nil
}

func CombineKeyFiles(args []string) (parts []Part, err error) {
	size := len(args)
	if size == 0 {
		return nil, nil
	}
	kCloseFns := make([]func(), size)
	kParts := make([]Part, size)
	count := 0
	for i, name := range args {
		var kf *os.File
		if kf, kCloseFns[i], err = files.OpenInputFile(name); err != nil {
			break
		}
		count++
		var kb []byte
		if kb, err = io.ReadAll(kf); err != nil {
			return
		}
		if err = json.Unmarshal(kb, &kParts[i]); err != nil {
			err = fmt.Errorf("not a valid sss key part\nCaused by: %v", err)
			return
		}
	}
	kCloseFns = kCloseFns[:count]
	defer func() {
		for _, kCloseFn := range kCloseFns {
			kCloseFn()
		}
	}()
	return kParts[:count], nil
}

func CombinePartFiles(in []string, out string, truncate, verbose bool) error {
	size := len(in)
	if size == 0 {
		return errors.New("no input files")
	}
	var output *os.File = nil
	var oCloseFn func()
	if len(out) > 0 {
		var err error
		if output, oCloseFn, err = files.OpenOutputFile(out, truncate); err != nil {
			return err
		}
	}
	defer oCloseFn()
	iFiles := make([]*os.File, size)
	iCloseFn := make([]func(), size)
	for i, path := range in {
		var err error
		if iFiles[i], iCloseFn[i], err = files.OpenInputFile(path); err != nil {
			return err
		}
	}
	defer func() {
		for _, closer := range iCloseFn {
			closer()
		}
		clear(iCloseFn)
		clear(iFiles)
	}()
	scanners := make([]*bufio.Scanner, size)
	for i, file := range iFiles {
		buf := make([]byte, maxScannerTokenSize)
		scanners[i] = bufio.NewScanner(file)
		scanners[i].Buffer(buf, maxScannerTokenSize)
		scanners[i].Split(bufio.ScanLines)
	}
	parts := make([]Part, size)
	count := 0
	for {
		var err error
		var lines [][]byte
		for _, scanner := range scanners {
			if scanner.Scan() {
				line := scanner.Bytes()
				lines = append(lines, line)
			}
			if err = scanner.Err(); err != nil {
				return err
			}
		}
		if len(lines) != size {
			break
		}
		if len(lines[0]) == 0 {
			continue
		}
		for i, line := range lines {
			if err := json.Unmarshal(line, &parts[i]); err != nil {
				return err
			}
		}
		threshold := parts[0].Threshold
		if len(parts) < int(threshold) {
			return errors.New(fmt.Sprintf("need %d input files", threshold))
		}
		block := parts[0].Block
		blocks := parts[0].Blocks
		if block != count+1 {
			return errors.New("block mismatch")
		}
		var secret []byte
		if secret, err = Combine(parts); err != nil {
			return err
		}
		expect := parts[0].Digest
		actual := utils.ComputeDigest(secret)
		if expect != actual {
			fmt.Printf("Expect secret digest: %s\n", expect)
			fmt.Printf("Actual secret digest: %s\n", actual)
			return errors.New("secret digest mismatch")
		}
		if count == 0 && verbose {
			fmt.Printf("Blocks count: %d\n", blocks)
		}
		if output != nil {
			if count == 0 {
				var stat os.FileInfo
				if stat, err = output.Stat(); err != nil {
					return err
				}
				if stat.Size() > 0 {
					if truncate {
						if err = output.Truncate(0); err != nil {
							return err
						}
						fmt.Printf("Truncate output file: %s\n", out)
					} else {
						return errors.New("output file is not empty")
					}
				}
			}
			if _, err = output.Write(secret); err != nil {
				return err
			}
		}
		count++
		if verbose {
			l := len(secret)
			w := len(fmt.Sprintf("%d", blocks))
			if output != nil {
				fmt.Printf("Block %*d/%d OK -- recovered %6d bytes and appended them into %s\n", w, block, blocks, l, out)
			} else {
				fmt.Printf("Block %*d/%d OK -- recovered %6d bytes\n", w, block, blocks, l)
			}
		}
	}
	return nil
}

var (
	ErrShareCountNotEnough = errors.New("length of shares must be at least 2")
	ErrFirstShareInvalid   = errors.New("length of first share must be at least 2")
	ErrDuplicatedShare     = errors.New("duplicated share is disallowed")
)

func CombineFromShares(shares []Share) ([]byte, error) {
	if len(shares) < 2 {
		return nil, ErrShareCountNotEnough
	}
	shareLen := len(shares[0])
	if shareLen < 2 {
		return nil, ErrFirstShareInvalid
	}
	for i := 1; i < len(shares); i++ {
		if len(shares[i]) != shareLen {
			return nil, fmt.Errorf("length of shares[%d] must be %d", i, shareLen)
		}
	}
	xSet := map[uint8]bool{}
	xSamples := make([]uint8, len(shares))
	for i, share := range shares {
		x := share[shareLen-1]
		xSamples[i] = x
		xSet[x] = true
	}
	if len(xSet) != len(xSamples) {
		return nil, ErrDuplicatedShare
	}
	secret := make([]byte, shareLen-1)
	ySamples := make([]uint8, len(shares))
	for idx := range secret {
		for i, share := range shares {
			ySamples[i] = share[idx]
		}
		val := gf256.InterpolatePolynomial(xSamples, ySamples, 0)
		secret[idx] = val
	}
	return secret, nil
}
