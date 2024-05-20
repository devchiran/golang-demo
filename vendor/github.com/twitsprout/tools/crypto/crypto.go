package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	prand "math/rand"
	"sync"
	"time"
)

// ReadRand fills the provided buffer with cryptographically random bytes. In
// the case that random data cannot be retrieved, an error is returned.
func ReadRand(buf []byte) error {
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return fmt.Errorf("unable to read random bytes: %s", err.Error())
	}
	return nil
}

// ReadRandUnsafe fills the provided buffer with cryptographically random bytes.
// In the case that random data cannot be retrieved, ReadRandUnsafe panics.
func ReadRandUnsafe(buf []byte) {
	if err := ReadRand(buf); err != nil {
		panic(err)
	}
}

var (
	muPRand sync.Mutex
	// Disable gosec linting here because we're using prand on purpose.
	//nolint:gosec
	pRandSrc = prand.New(prand.NewSource(time.Now().UnixNano()))
)

// ReadPRand fills the provided buffer with random bytes from a a pseduo-random
// source.
func ReadPRand(buf []byte) error {
	muPRand.Lock()
	defer muPRand.Unlock()
	_, err := io.ReadFull(pRandSrc, buf)
	return err
}

// PRandInt64 returns a pseduo-random 64-bit integer in the range of [min, max).
// It panics if max <= min.
func PRandInt64(min, max int64) int64 {
	muPRand.Lock()
	n := pRandSrc.Int63n(max - min)
	muPRand.Unlock()
	return min + n
}

// Encode uses the provided cipher block to encode 'data', returning the result.
func Encode(c cipher.Block, data []byte) []byte {
	// Generate random IV.
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	ReadRandUnsafe(iv)

	// Encrypt the provided id.
	stream := cipher.NewCFBEncrypter(c, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	// Return the encoded ciphertext.
	return ciphertext
}

// Decode uses the provided cipher block to decode 'data', returning the result
// and any error encountered.
func Decode(c cipher.Block, data []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("crypto: length of data too short: %d", len(data))
	}

	// Decrypt ciphertext.
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(c, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}
