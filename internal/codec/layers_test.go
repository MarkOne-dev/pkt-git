package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestStage2RoundTrip(t *testing.T) {
	for _, size := range []int{0, 1, 2, 255, 256, 257, 1024} {
		in := make([]byte, size)
		for i := range in {
			in[i] = byte(i)
		}
		got := unmaskStage2(maskStage2(in))
		if !bytes.Equal(got, in) {
			t.Fatalf("size %d: round trip mismatch", size)
		}
	}
}

func TestStage4RoundTrip(t *testing.T) {
	for _, size := range []int{0, 1, 2, 3, 16, 255, 256, 257, 1024} {
		in := make([]byte, size)
		for i := range in {
			in[i] = byte(i * 7)
		}
		got := unmaskStage4(maskStage4(in))
		if !bytes.Equal(got, in) {
			t.Fatalf("size %d: round trip mismatch", size)
		}
	}
}

func TestStage4Reverses(t *testing.T) {
	in := []byte{1, 2, 3, 4}
	got := maskStage4(in)
	want := []byte{252, 255, 2, 5}
	if !bytes.Equal(got, want) {
		t.Fatalf("masked = %v, want %v", got, want)
	}
}

func TestQCompressRoundTrip(t *testing.T) {
	for _, in := range [][]byte{
		nil,
		[]byte("hello"),
		bytes.Repeat([]byte("packet-tracer"), 1000),
	} {
		compressed, err := qCompress(in)
		if err != nil {
			t.Fatalf("qCompress: %v", err)
		}
		if len(compressed) < 4 {
			t.Fatalf("compressed length = %d, want >= 4", len(compressed))
		}
		if got := binary.BigEndian.Uint32(compressed[:4]); got != uint32(len(in)) {
			t.Fatalf("declared size = %d, want %d", got, len(in))
		}
		got, err := qUncompress(compressed)
		if err != nil {
			t.Fatalf("qUncompress: %v", err)
		}
		if !bytes.Equal(got, in) {
			t.Fatalf("round trip mismatch")
		}
	}
}

func TestQCompressIsDeterministic(t *testing.T) {
	in := bytes.Repeat([]byte("deterministic"), 100)
	a, err := qCompress(in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := qCompress(in)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("qCompress output is not deterministic")
	}
}

func TestQUncompressTruncated(t *testing.T) {
	if _, err := qUncompress(nil); !errors.Is(err, ErrTruncated) {
		t.Fatalf("error = %v, want ErrTruncated", err)
	}
	if _, err := qUncompress([]byte{0, 0, 0}); !errors.Is(err, ErrTruncated) {
		t.Fatalf("error = %v, want ErrTruncated", err)
	}
}

func TestQUncompressLengthMismatch(t *testing.T) {
	compressed, err := qCompress([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	binary.BigEndian.PutUint32(compressed[:4], 99)
	if _, err := qUncompress(compressed); !errors.Is(err, ErrLengthMismatch) {
		t.Fatalf("error = %v, want ErrLengthMismatch", err)
	}
}
