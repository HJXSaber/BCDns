package utils

import (
	"crypto/sha256"
)

func SHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}