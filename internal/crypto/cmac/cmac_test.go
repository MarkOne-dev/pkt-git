package cmac

import (
	"crypto/aes"
	"encoding/hex"
	"testing"
)

// NIST SP 800-38B, Appendix D.1 (AES-128).
func TestNISTVectors(t *testing.T) {
	key := mustHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"empty", "", "bb1d6929e95937287fa37d129b756746"},
		{"one block", "6bc1bee22e409f96e93d7e117393172a", "070a16b46b4d4144f79bdd9dd04a287c"},
		{
			"partial block",
			"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51" +
				"30c81c46a35ce411",
			"dfa66747de9ae63030ca32611497c827",
		},
		{
			"multiple blocks",
			"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e51" +
				"30c81c46a35ce411e5fbc1191a0a52eff69f2445df4f9b17ad2b417be66c3710",
			"51f0bebf7e3b9d92fc49741779363cfe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sum(block, mustHex(t, tt.msg))
			if hex.EncodeToString(got) != tt.want {
				t.Fatalf("CMAC = %x, want %s", got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	if !Equal([]byte{1, 2, 3}, []byte{1, 2, 3}) {
		t.Fatal("Equal returned false for identical MACs")
	}
	if Equal([]byte{1, 2, 3}, []byte{1, 2, 4}) {
		t.Fatal("Equal returned true for different MACs")
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
