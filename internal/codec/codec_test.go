package codec

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/MarkOne-dev/pkt-git/internal/crypto/eax"
	"github.com/MarkOne-dev/pkt-git/internal/crypto/twofish"
)

func TestSpecimenRoundTrip(t *testing.T) {
	raw := loadSpecimen(t)

	xmlData, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if err := ValidateXML(xmlData); err != nil {
		t.Fatalf("decoded XML is invalid: %v", err)
	}
	if !bytes.Contains(xmlData, []byte("<PACKETTRACER5")) {
		t.Fatal("decoded XML does not contain a PACKETTRACER5 root")
	}

	rebuilt, err := Encode(xmlData)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	xmlAgain, err := Decode(rebuilt)
	if err != nil {
		t.Fatalf("Decode rebuilt: %v", err)
	}
	if !bytes.Equal(xmlData, xmlAgain) {
		t.Fatal("XML round trip mismatch")
	}
}

func TestVerifySpecimen(t *testing.T) {
	result, err := Verify(loadSpecimen(t))
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !result.XMLIsDocument {
		t.Fatal("XMLIsDocument = false")
	}
	if !result.XMLRoundTrip {
		t.Fatal("XMLRoundTrip = false")
	}
	if result.BinarySize == 0 || result.XMLSize == 0 || result.RebuiltSize == 0 {
		t.Fatalf("unexpected sizes: %+v", result)
	}
}

func TestEncodeIsDeterministic(t *testing.T) {
	xmlData, err := Decode(loadSpecimen(t))
	if err != nil {
		t.Fatal(err)
	}
	a, err := Encode(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encode(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("Encode output is not deterministic")
	}
}

func TestEncodeRejectsInvalidXML(t *testing.T) {
	if _, err := Encode([]byte("<root>")); !errors.Is(err, ErrInvalidXML) {
		t.Fatalf("Encode error = %v, want ErrInvalidXML", err)
	}
}

func TestDecodeRejectsTruncated(t *testing.T) {
	raw := loadSpecimen(t)
	for _, input := range [][]byte{nil, raw[:len(raw)-1]} {
		if _, err := Decode(input); err == nil {
			t.Fatal("Decode accepted a truncated container")
		}
	}
}

func TestDecodeRejectsTamperedTag(t *testing.T) {
	raw := append([]byte(nil), loadSpecimen(t)...)
	raw[len(raw)-1] ^= 0x01
	if _, err := Decode(raw); !errors.Is(err, eax.ErrAuthentication) {
		t.Fatalf("Decode error = %v, want ErrAuthentication", err)
	}
}

func TestDecodeRejectsTamperedCiphertext(t *testing.T) {
	raw := append([]byte(nil), loadSpecimen(t)...)
	raw[len(raw)/2] ^= 0x01
	if _, err := Decode(raw); !errors.Is(err, eax.ErrAuthentication) {
		t.Fatalf("Decode error = %v, want ErrAuthentication", err)
	}
}

func TestDecodeRejectsLengthMismatch(t *testing.T) {
	xmlData, err := Decode(loadSpecimen(t))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := qCompress(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	payload[0] ^= 0xff

	block, err := twofish.NewCipher(containerKey)
	if err != nil {
		t.Fatal(err)
	}
	sealed := eax.Encrypt(block, maskStage2(payload), containerNonce, nil)
	if _, err := Decode(maskStage4(sealed)); !errors.Is(err, ErrLengthMismatch) {
		t.Fatalf("Decode error = %v, want ErrLengthMismatch", err)
	}
}

func TestValidateXML(t *testing.T) {
	valid := []string{
		"<root><child/></root>",
		"<?xml version=\"1.0\"?>\n<!-- comment -->\n<root/>\n",
		"<!DOCTYPE root><root/>",
		"<root/><?xml-stylesheet href=\"style.xsl\"?>",
		"\xEF\xBB\xBF<root/>",
	}
	for _, input := range valid {
		if err := ValidateXML([]byte(input)); err != nil {
			t.Errorf("ValidateXML(%q): %v", input, err)
		}
	}

	invalid := []string{
		"",
		"   \n",
		"<!-- comment only -->",
		"<root>",
		"<a/><b/>",
		"<a/>junk",
		"<a></a>junk",
		"<a/><?xml version=\"1.0\"?>",
		"<a/><!DOCTYPE foo>",
		"<!DOCTYPE a><!DOCTYPE b><root/>",
		"<!ENTITY foo><root/>",
		"<!-- comment --><?xml version=\"1.0\"?><root/>",
	}
	for _, input := range invalid {
		if err := ValidateXML([]byte(input)); err == nil {
			t.Errorf("ValidateXML accepted invalid document %q", input)
		}
	}
}

func loadSpecimen(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "specimens", "Caso1_Grupo09-2.pkt")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read specimen: %v", err)
	}
	return raw
}
