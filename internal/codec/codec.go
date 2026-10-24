// Package codec implements the reversible Packet Tracer container.
package codec

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"github.com/MarkOne-dev/pkt-git/internal/crypto/eax"
	"github.com/MarkOne-dev/pkt-git/internal/crypto/twofish"
)

var (
	// ErrInvalidXML reports an XML document that cannot be parsed.
	ErrInvalidXML = errors.New("codec: invalid XML document")
)

var (
	containerKey   = bytes.Repeat([]byte{0x89}, 16)
	containerNonce = bytes.Repeat([]byte{0x10}, 16)
)

// Decode converts a Packet Tracer binary into its UTF-8 XML document.
func Decode(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, ErrTruncated
	}

	block, err := twofish.NewCipher(containerKey)
	if err != nil {
		return nil, fmt.Errorf("codec: create cipher: %w", err)
	}

	data := unmaskStage4(raw)
	plain, err := eax.Decrypt(block, data, containerNonce, nil)
	if err != nil {
		return nil, fmt.Errorf("codec: decode container: %w", err)
	}
	return qUncompress(unmaskStage2(plain))
}

// Encode converts an XML document into a Packet Tracer binary.
func Encode(xmlData []byte) ([]byte, error) {
	if err := ValidateXML(xmlData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidXML, err)
	}

	block, err := twofish.NewCipher(containerKey)
	if err != nil {
		return nil, fmt.Errorf("codec: create cipher: %w", err)
	}

	payload, err := qCompress(xmlData)
	if err != nil {
		return nil, err
	}
	sealed := eax.Encrypt(block, maskStage2(payload), containerNonce, nil)
	return maskStage4(sealed), nil
}

// ValidateXML reports whether data is a complete, well-formed XML document.
func ValidateXML(data []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// VerifyResult contains the observable results of a round-trip check.
type VerifyResult struct {
	BinarySize    int
	XMLSize       int
	RebuiltSize   int
	XMLIsDocument bool
	XMLRoundTrip  bool
	ByteIdentical bool
}

// Verify decodes raw, re-encodes the XML, and decodes the rebuilt binary.
func Verify(raw []byte) (VerifyResult, error) {
	result := VerifyResult{BinarySize: len(raw)}

	xmlData, err := Decode(raw)
	if err != nil {
		return result, err
	}
	result.XMLSize = len(xmlData)
	result.XMLIsDocument = ValidateXML(xmlData) == nil
	if !result.XMLIsDocument {
		return result, nil
	}

	rebuilt, err := Encode(xmlData)
	if err != nil {
		return result, err
	}
	result.RebuiltSize = len(rebuilt)
	result.ByteIdentical = bytes.Equal(rebuilt, raw)

	xmlAgain, err := Decode(rebuilt)
	if err != nil {
		return result, err
	}
	result.XMLRoundTrip = bytes.Equal(xmlData, xmlAgain)
	return result, nil
}
