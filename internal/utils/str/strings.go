package str

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomString(byteCount int) (string, error) {
	str := make([]byte, byteCount)
	_, err := rand.Read(str)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(str), nil
}
