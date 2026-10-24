// Package eax implements the Crypto++-style EAX mode used by Packet Tracer:
// the 16-byte authentication tag is appended to the ciphertext.
package eax

import (
	"crypto/cipher"
	"crypto/subtle"
	"errors"

	"github.com/MarkOne-dev/pkt-git/internal/crypto/cmac"
)

var (
	// ErrAuthentication reports a tag mismatch.
	ErrAuthentication = errors.New("eax: authentication failed")
	// ErrInvalidSize reports an input shorter than the authentication tag.
	ErrInvalidSize = errors.New("eax: ciphertext is shorter than the tag")
)

const blockSize = 16

// Encrypt returns ciphertext || tag.
func Encrypt(block cipher.Block, plain, nonce, header []byte) []byte {
	if block.BlockSize() != blockSize {
		panic("eax: only 128-bit block ciphers are supported")
	}

	n := omac(block, 0, nonce)
	stream := ctr(block, n, len(plain))
	out := make([]byte, len(plain)+blockSize)
	for i := range plain {
		out[i] = plain[i] ^ stream[i]
	}
	tag := xor(n, xor(omac(block, 1, header), omac(block, 2, out[:len(plain)])))
	copy(out[len(plain):], tag)
	return out
}

// Decrypt verifies the trailing tag and returns the plaintext.
func Decrypt(block cipher.Block, data, nonce, header []byte) ([]byte, error) {
	if block.BlockSize() != blockSize {
		panic("eax: only 128-bit block ciphers are supported")
	}
	if len(data) < blockSize {
		return nil, ErrInvalidSize
	}

	ciphertext := data[:len(data)-blockSize]
	tag := data[len(data)-blockSize:]
	n := omac(block, 0, nonce)
	stream := ctr(block, n, len(ciphertext))
	plain := make([]byte, len(ciphertext))
	for i := range ciphertext {
		plain[i] = ciphertext[i] ^ stream[i]
	}
	expected := xor(n, xor(omac(block, 1, header), omac(block, 2, ciphertext)))
	if subtle.ConstantTimeCompare(expected, tag) != 1 {
		return nil, ErrAuthentication
	}
	return plain, nil
}

// omac implements EAX's OMAC^t(x): CMAC over 00*15 || t || x.
func omac(block cipher.Block, tweak byte, data []byte) []byte {
	prefix := make([]byte, blockSize)
	prefix[blockSize-1] = tweak
	return cmac.Sum(block, append(prefix, data...))
}

func ctr(block cipher.Block, initial []byte, length int) []byte {
	counter := append([]byte(nil), initial...)
	out := make([]byte, 0, length)
	for len(out) < length {
		blockOut := make([]byte, blockSize)
		block.Encrypt(blockOut, counter)
		out = append(out, blockOut...)
		increment(counter)
	}
	return out[:length]
}

func increment(counter []byte) {
	for i := len(counter) - 1; i >= 0; i-- {
		counter[i]++
		if counter[i] != 0 {
			return
		}
	}
}

func xor(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := range a {
		out[i] = a[i] ^ b[i]
	}
	return out
}
