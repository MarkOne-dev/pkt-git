// Package codec implements the reversible Packet Tracer container.
package codec

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/crypto/eax"
	"github.com/MarkOne-dev/pkt-git/internal/crypto/twofish"
)

const (
	// SupportedVersion is the Packet Tracer version covered by the bundled
	// regression specimen.
	SupportedVersion = "9.0.1.0858"
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
		return nil, fmt.Errorf("codec: decode container: %w; check the producing Packet Tracer version (tested: %s)", err, SupportedVersion)
	}
	xmlData, err := qUncompress(unmaskStage2(plain))
	if err != nil {
		return nil, fmt.Errorf("codec: decompress container: %w; check the producing Packet Tracer version (tested: %s)", err, SupportedVersion)
	}
	return xmlData, nil
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

// ValidateXML reports whether data is a complete, well-formed XML document
// with exactly one root element. Comments, processing instructions, and
// whitespace around the root are allowed; character data outside the root and
// multiple roots are rejected.
func ValidateXML(data []byte) error {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	decoder := xml.NewDecoder(bytes.NewReader(data))
	depth := 0
	seenRoot := false
	seenToken := false
	seenDoctype := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			if !seenRoot {
				return errors.New("XML document has no root element")
			}
			return nil
		}
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				if seenRoot {
					return errors.New("XML document has multiple root elements")
				}
				seenRoot = true
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(token)) != "" {
				return errors.New("XML document has character data outside the root element")
			}
		case xml.ProcInst:
			if strings.EqualFold(token.Target, "xml") && seenToken {
				return errors.New("XML declaration is not at the start of the document")
			}
		case xml.Directive:
			if seenRoot || seenDoctype || !bytes.HasPrefix(bytes.TrimSpace(token), []byte("DOCTYPE")) {
				return errors.New("XML document has a misplaced or unsupported directive")
			}
			seenDoctype = true
		}
		seenToken = true
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
