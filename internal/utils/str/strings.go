package str

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomString() (string, error) {
	str := make([]byte, 32)
	_, err := rand.Read(str)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(str), nil
}
