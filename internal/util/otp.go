package util

import (
	"crypto/rand"
	"math/big"
)

func GenerateOTP6() (string, error) {
	min := int64(100000)
	max := int64(900000) // random [0..899999] + 100000 => 100000..999999

	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	code := min + nBig.Int64()
	return itoa6(code), nil
}

func itoa6(n int64) string {
	b := []byte("000000")
	for i := 5; i >= 0; i-- {
		b[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(b)
}
