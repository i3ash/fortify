package sss

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/i3ash/fortify/files"
	"github.com/i3ash/fortify/pkg/gf256"
	"github.com/i3ash/fortify/utils"
)

func Split(secret []byte, parts, threshold uint8) ([]Part, error) {
	out, err := SplitIntoShares(secret, parts, threshold)
	if err != nil {
		return nil, err
	}
	var outParts []Part
	digest := utils.ComputeDigest(secret)
	for index, share := range out {
		p := Part{
			Parts:     parts,
			Part:      index + 1,
			Payload:   base64.URLEncoding.EncodeToString(share),
			Timestamp: time.Now(),
			Threshold: threshold,
			Digest:    digest,
		}
		outParts = append(outParts, p)
	}
	return outParts, nil
}

func SplitIntoFiles(in string, parts, threshold uint8, prefix string, truncate, verbose bool) error {
	file, closer, err := files.OpenInputFile(in)
	if err != nil {
		return err
	}
	defer closer()
	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		return err
	}
	blocks := int(math.Ceil(float64(stat.Size()) / float64(fileBlockSize)))
	reader := bufio.NewReader(file)
	buffer := make([]byte, fileBlockSize)
	var bytesRead, block int
	var ps []Part
	for {
		bytesRead, err = reader.Read(buffer)
		if err != nil {
			return err
		}
		secret := buffer[:bytesRead]
		ps, err = Split(secret, parts, threshold)
		if err != nil {
			return err
		}
		err = AppendParts(ps, block, blocks, prefix, truncate)
		if err != nil {
			return err
		}
		block++
		if verbose {
			w := len(fmt.Sprintf("%d", blocks))
			fmt.Printf("Block %*d/%d OK\n", w, block, blocks)
		}
		if bytesRead < fileBlockSize {
			break
		}
	}
	return nil
}

func AppendParts(ps []Part, block, blocks int, prefix string, truncate bool) error {
	size := len(ps)
	var wg sync.WaitGroup
	wg.Add(size)
	errCh := make(chan error, len(ps))
	for i, p := range ps {
		{
			path := fmt.Sprintf("%s%dof%d.json", prefix, p.Part, p.Parts)
			file, err := OpenFileForWrite(path, truncate)
			if err != nil {
				return err
			}
			ps[i].file = file
			ps[i].Block = block + 1
			ps[i].Blocks = blocks
		}
		go func(wg *sync.WaitGroup, p Part) {
			defer wg.Done()
			if err := appendPart(&p, block); err != nil {
				errCh <- err
				return
			}
			//fmt.Printf("Part %d/%d: %s\n", p.Part, p.Parts, p.file.Name())
		}(&wg, ps[i])
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func appendPart(p *Part, block int) (err error) {
	file := p.file
	if block == 0 {
		if err = file.Truncate(0); err != nil {
			return
		}
	}
	var content []byte
	content, err = json.Marshal(p)
	if err != nil {
		return
	}
	if block > 0 {
		_, err = file.WriteString("\n\n")
		if err != nil {
			return
		}
	}
	_, err = file.Write(content)
	if err != nil {
		return
	}
	return nil
}

var openedFilesForWrite = make(map[string]*os.File)
var openedFilesForWriteCloser = make(map[string]func())
var openedFilesForWriteLock sync.Mutex

func OpenFileForWrite(path string, truncate bool) (*os.File, error) {
	openedFilesForWriteLock.Lock()
	defer openedFilesForWriteLock.Unlock()
	file, _ := openedFilesForWrite[path]
	if file != nil {
		return file, nil
	}
	var err error
	var closer func()
	if file, closer, err = files.OpenOutputFile(path, truncate); err != nil {
		return nil, err
	}
	openedFilesForWrite[path] = file
	openedFilesForWriteCloser[path] = closer
	return file, nil
}

func CloseAllFilesForWrite() {
	openedFilesForWriteLock.Lock()
	defer openedFilesForWriteLock.Unlock()
	for _, closer := range openedFilesForWriteCloser {
		closer()
	}
	clear(openedFilesForWrite)
	clear(openedFilesForWriteCloser)
}

var (
	ErrThresholdTooSmall     = errors.New("threshold must be at least 2")
	ErrInvalidPartsThreshold = errors.New("threshold cannot be greater than parts")
	ErrEmptySecret           = errors.New("secret is empty")
)

func SplitIntoShares(secret []byte, parts, threshold uint8) ([]Share, error) {
	if threshold < 2 {
		return nil, ErrThresholdTooSmall
	}
	if threshold > parts {
		return nil, ErrInvalidPartsThreshold
	}
	if len(secret) == 0 {
		return nil, ErrEmptySecret
	}
	var xs []uint8
	if xs0, err := generateSecureXCoordinates(parts); err != nil {
		return nil, err
	} else {
		xs = xs0
	}
	secretLen := len(secret)
	shares := make([]Share, parts)
	for i := range shares {
		shares[i] = make([]byte, secretLen+1)
		shares[i][secretLen] = xs[i]
	}
	for j, intercept := range secret {
		p, err := gf256.NewPolynomial(intercept, threshold-1, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create polynomial: %w", err)
		}
		for i := range shares {
			x := shares[i][secretLen]
			shares[i][j] = gf256.PolynomialEvaluate(p, x)
		}
	}
	return shares, nil
}

// generateSecureXCoordinates generates cryptographically secure x coordinates
func generateSecureXCoordinates(count uint8) ([]uint8, error) {
	// Generate random bytes for x coordinates
	xValues := make([]uint8, count)
	if _, err := io.ReadFull(rand.Reader, xValues); err != nil {
		return nil, fmt.Errorf("failed to generate secure x coordinates: %w", err)
	}
	// Ensure no duplicate values and no zeros
	used := make(map[uint8]bool)
	for i, x := range xValues {
		// Make sure it's not a zero, which is reserved for the secret
		for x == 0 || used[x] {
			// Get a new random byte
			buf := make([]byte, 1)
			if _, err := io.ReadFull(rand.Reader, buf); err != nil {
				return nil, fmt.Errorf("failed to generate unique x coordinate: %w", err)
			}
			x = buf[0]
		}
		used[x] = true
		xValues[i] = x
	}
	return xValues, nil
}
