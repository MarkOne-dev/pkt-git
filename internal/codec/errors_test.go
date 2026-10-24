package codec

import (
	"testing"
)

func TestDecodeReturnsNoXMLOnFailure(t *testing.T) {
	raw := loadSpecimen(t)
	tamperedTag := append([]byte(nil), raw...)
	tamperedTag[len(tamperedTag)-1] ^= 0x01
	tamperedBody := append([]byte(nil), raw...)
	tamperedBody[len(tamperedBody)/2] ^= 0x01

	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", nil},
		{"truncated", raw[:len(raw)-1]},
		{"bad tag", tamperedTag},
		{"bad body", tamperedBody},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.input)
			if err == nil {
				t.Fatal("Decode succeeded")
			}
			if got != nil {
				t.Fatalf("Decode returned %d bytes of XML on failure", len(got))
			}
		})
	}
}

func TestEncodeReturnsNoBinaryOnFailure(t *testing.T) {
	got, err := Encode([]byte("<root>"))
	if err == nil {
		t.Fatal("Encode succeeded with malformed XML")
	}
	if got != nil {
		t.Fatalf("Encode returned %d bytes on failure", len(got))
	}
}

func TestVerifyReturnsErrorOnFailure(t *testing.T) {
	result, err := Verify(nil)
	if err == nil {
		t.Fatal("Verify succeeded on an empty container")
	}
	if result.XMLSize != 0 || result.RebuiltSize != 0 {
		t.Fatalf("Verify returned partial result: %+v", result)
	}
}
