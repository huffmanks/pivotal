package link

import (
	"crypto/rand"
	"math/big"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateBase62ID(length int) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(base62Chars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b)
}
