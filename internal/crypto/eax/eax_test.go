package eax

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	block := mustAES(t)
	nonce := bytes.Repeat([]byte{0x10}, 16)
	header := []byte("header")
	plain := []byte("the quick brown fox jumps over the lazy dog")

	sealed := Encrypt(block, plain, nonce, header)
	if len(sealed) != len(plain)+16 {
		t.Fatalf("sealed length = %d, want %d", len(sealed), len(plain)+16)
	}

	got, err := Decrypt(block, sealed, nonce, header)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("plaintext = %q, want %q", got, plain)
	}
}

func TestEmptyPlaintext(t *testing.T) {
	block := mustAES(t)
	nonce := bytes.Repeat([]byte{0x10}, 16)
	sealed := Encrypt(block, nil, nonce, nil)
	if len(sealed) != 16 {
		t.Fatalf("sealed length = %d, want 16", len(sealed))
	}
	got, err := Decrypt(block, sealed, nonce, nil)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("plaintext length = %d, want 0", len(got))
	}
}

func TestTamperedTag(t *testing.T) {
	block := mustAES(t)
	nonce := bytes.Repeat([]byte{0x10}, 16)
	sealed := Encrypt(block, []byte("payload"), nonce, nil)
	sealed[len(sealed)-1] ^= 0x01

	if _, err := Decrypt(block, sealed, nonce, nil); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("Decrypt error = %v, want ErrAuthentication", err)
	}
}

func TestTamperedCiphertext(t *testing.T) {
	block := mustAES(t)
	nonce := bytes.Repeat([]byte{0x10}, 16)
	sealed := Encrypt(block, []byte("payload"), nonce, nil)
	sealed[0] ^= 0x01

	if _, err := Decrypt(block, sealed, nonce, nil); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("Decrypt error = %v, want ErrAuthentication", err)
	}
}

func TestShortCiphertext(t *testing.T) {
	block := mustAES(t)
	if _, err := Decrypt(block, make([]byte, 15), nil, nil); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("Decrypt error = %v, want ErrInvalidSize", err)
	}
}

func TestDifferentNonceFails(t *testing.T) {
	block := mustAES(t)
	sealed := Encrypt(block, []byte("payload"), bytes.Repeat([]byte{0x10}, 16), nil)
	if _, err := Decrypt(block, sealed, bytes.Repeat([]byte{0x11}, 16), nil); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("Decrypt error = %v, want ErrAuthentication", err)
	}
}

func mustAES(t *testing.T) cipher.Block {
	t.Helper()
	block, err := aes.NewCipher(bytes.Repeat([]byte{0x42}, 16))
	if err != nil {
		t.Fatal(err)
	}
	return block
}
