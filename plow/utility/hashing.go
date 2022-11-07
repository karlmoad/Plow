package utility

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func Sha256Hashf(format string, args ...interface{}) string {
	s := fmt.Sprintf(format, args)
	return Sha256Hash(s)
}

func Sha256Hash(s string) string {
	bytes := []byte(s)
	b := sha256.Sum256(bytes)
	return hex.EncodeToString(b[:])
}

func Sha256HashBytes(bytes []byte) string {
	b := sha256.Sum256(bytes)
	return hex.EncodeToString(b[:])
}
