package crypto

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

const (
	chars      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bitsNeeded = 6 // len(chars) = 62 = 0b111110
	mask       = 1<<bitsNeeded - 1
	max        = 63 / bitsNeeded
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

// RandomString generate a random string of length n
// Modified from https://stackoverflow.com/a/31832326
func RandomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), max; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), max
		}
		if idx := int(cache & mask); idx < len(chars) {
			b[i] = chars[idx]
			i--
		}
		cache >>= bitsNeeded
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// RandomBytes uses crypto/rand to generate random bytes.
func RandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := crand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes from crypto/rand: %w", err)
	}
	return b, nil
}

// MustRandomBytes uses crypto/rand to generate random bytes. Panics on error.
func MustRandomBytes(n uint32) []byte {
	b, err := RandomBytes(n)
	if err != nil {
		panic(err)
	}
	return b
}
