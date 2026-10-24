package twofish

import (
	"bytes"
	"testing"
)

func TestNewCipher(t *testing.T) {
	c, err := NewCipher(bytes.Repeat([]byte{0x89}, 16))
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	if got := c.BlockSize(); got != 16 {
		t.Fatalf("BlockSize = %d, want 16", got)
	}

	src := make([]byte, c.BlockSize())
	dst := make([]byte, c.BlockSize())
	c.Encrypt(dst, src)
	if bytes.Equal(src, dst) {
		t.Fatal("Encrypt returned the input unchanged")
	}
}

func TestNewCipherRejectsBadKey(t *testing.T) {
	if _, err := NewCipher([]byte("short")); err == nil {
		t.Fatal("NewCipher accepted an invalid key size")
	}
}
