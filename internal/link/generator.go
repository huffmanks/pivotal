package link

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateBase62ID(length int) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(base62Chars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random ID: %w", err)
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b), nil
}
