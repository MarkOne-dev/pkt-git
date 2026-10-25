// Package twofish exposes the Twofish block cipher used by the Packet Tracer
// container. It keeps the external dependency behind a small internal boundary.
package twofish

import (
	"crypto/cipher"

	xcrypto "golang.org/x/crypto/twofish"
)

// NewCipher creates a Twofish cipher with a 16, 24, or 32-byte key.
func NewCipher(key []byte) (cipher.Block, error) {
	return xcrypto.NewCipher(key)
}
