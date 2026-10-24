// Package cmac implements CMAC (OMAC1) as specified by NIST SP 800-38B.
package cmac

import (
	"crypto/cipher"
	"crypto/subtle"
)

// Sum returns the CMAC of msg using the given 128-bit block cipher.
func Sum(block cipher.Block, msg []byte) []byte {
	bs := block.BlockSize()
	if bs != 16 {
		panic("cmac: only 128-bit block ciphers are supported")
	}

	zero := make([]byte, bs)
	l := make([]byte, bs)
	block.Encrypt(l, zero)
	k1 := double(l)
	k2 := double(k1)

	var body, last []byte
	if len(msg) == 0 {
		last = xor(pad(nil, bs), k2)
	} else {
		rem := len(msg) % bs
		if rem == 0 {
			body = msg[:len(msg)-bs]
			last = xor(msg[len(msg)-bs:], k1)
		} else {
			body = msg[:len(msg)-rem]
			last = xor(pad(msg[len(msg)-rem:], bs), k2)
		}
	}

	acc := make([]byte, bs)
	for i := 0; i < len(body); i += bs {
		acc = encrypt(block, xor(acc, body[i:i+bs]))
	}
	return encrypt(block, xor(acc, last))
}

// Equal compares two MACs in constant time.
func Equal(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

func double(in []byte) []byte {
	out := make([]byte, len(in))
	carry := in[0] & 0x80
	for i := 0; i < len(in)-1; i++ {
		out[i] = in[i]<<1 | in[i+1]>>7
	}
	out[len(in)-1] = in[len(in)-1] << 1
	if carry != 0 {
		out[len(in)-1] ^= 0x87
	}
	return out
}

func pad(in []byte, bs int) []byte {
	out := make([]byte, bs)
	copy(out, in)
	out[len(in)] = 0x80
	return out
}

func xor(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := range a {
		out[i] = a[i] ^ b[i]
	}
	return out
}

func encrypt(block cipher.Block, in []byte) []byte {
	out := make([]byte, len(in))
	block.Encrypt(out, in)
	return out
}
