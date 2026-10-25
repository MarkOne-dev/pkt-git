package codec

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrTruncated reports an input that cannot contain a complete container.
	ErrTruncated = errors.New("codec: truncated container")
	// ErrLengthMismatch reports a qCompress size header that disagrees with
	// the decompressed payload.
	ErrLengthMismatch = errors.New("codec: declared length mismatch")
)

func qCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))
	buf.Write(header[:])

	zw, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("codec: create zlib writer: %w", err)
	}
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return nil, fmt.Errorf("codec: compress payload: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("codec: close zlib writer: %w", err)
	}
	return buf.Bytes(), nil
}

func qUncompress(payload []byte) ([]byte, error) {
	if len(payload) < 4 {
		return nil, ErrTruncated
	}
	declared := binary.BigEndian.Uint32(payload[:4])

	zr, err := zlib.NewReader(bytes.NewReader(payload[4:]))
	if err != nil {
		return nil, fmt.Errorf("codec: open zlib stream: %w", err)
	}
	defer zr.Close()

	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("codec: decompress payload: %w", err)
	}
	if uint64(len(data)) != uint64(declared) {
		return nil, fmt.Errorf("%w: declared %d bytes, got %d", ErrLengthMismatch, declared, len(data))
	}
	return data, nil
}

func maskStage2(data []byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ byte((len(data)-i)&0xff)
	}
	return out
}

func unmaskStage2(data []byte) []byte {
	return maskStage2(data)
}

func maskStage4(data []byte) []byte {
	n := int64(len(data))
	out := make([]byte, len(data))
	for i, b := range data {
		out[n-1-int64(i)] = b ^ byte((n-int64(i)*n)&0xff)
	}
	return out
}

func unmaskStage4(data []byte) []byte {
	n := int64(len(data))
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[n-1-int64(i)] ^ byte((n-int64(i)*n)&0xff)
	}
	return out
}
