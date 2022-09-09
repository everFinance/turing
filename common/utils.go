package common

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashBytes(data []byte) string {
	shaSum := sha256.Sum256(data)
	hash := hex.EncodeToString(shaSum[:])
	return hash
}
