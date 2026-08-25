package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func FileChecksum(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:]), nil
}
