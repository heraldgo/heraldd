package util

import (
	"crypto/sha256"
	"io"
	"os"
)

// SHA256Sum calculate the sha256 checksum from the Reader
func SHA256Sum(r io.Reader) ([]byte, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// SHA256SumFile calculate the sha256 checksum from the file
func SHA256SumFile(fn string) ([]byte, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return SHA256Sum(f)
}
