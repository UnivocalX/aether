package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"sync"
)

func ParallelProcess[T any, R any](
	items []T,
	numWorkers int,
	workerFunc func(T) (R, error),
) ([]R, []error) {
	if len(items) == 0 {
		return []R{}, []error{}
	}

	if numWorkers <= 0 {
		numWorkers = 1
	}
	if numWorkers > len(items) {
		numWorkers = len(items)
	}

	var (
		wg       sync.WaitGroup
		muResults sync.Mutex
		muErrors   sync.Mutex
		results   []R
		errors    []error
	)

	workCh := make(chan T, len(items))
	for _, item := range items {
		workCh <- item
	}
	close(workCh)

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for item := range workCh {
				result, err := workerFunc(item)
				if err != nil {
					muErrors.Lock()
					errors = append(errors, err)
					muErrors.Unlock()
				} else {
					muResults.Lock()
					results = append(results, result)
					muResults.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return results, errors
}

func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

func ValidateSHA256(hash string) error {
	h := strings.TrimSpace(hash)
	if len(h) != sha256.Size*2 {
		return fmt.Errorf("invalid SHA256 hash length: %d", len(h))
	}
	if _, err := hex.DecodeString(h); err != nil {
		return fmt.Errorf("invalid SHA256 hex: %w", err)
	}
	return nil
}

// hexToBase64SHA256 converts hex-encoded SHA256 to base64
func HexToBase64SHA256(hexSHA256 string) (string, error) {
	decoded, err := hex.DecodeString(hexSHA256)
	if err != nil {
		return "", fmt.Errorf("invalid hex: %w", err)
	}
	return base64.StdEncoding.EncodeToString(decoded), nil
}
